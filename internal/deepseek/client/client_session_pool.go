package client

import (
	"sync"
	"time"

	"ds2api/internal/config"
)

const (
	// 默认每个会话最大消息数
	defaultMaxMessagesPerSession = 50
	// 会话池过期清理间隔
	poolCleanupInterval = 5 * time.Minute
	// 会话默认 TTL（12小时，与 auto_delete delay_hours 保持一致）
	defaultSessionTTL = 12 * time.Hour
	// 等待会话创建的超时时间
	sessionCreateTimeout = 30 * time.Second
)

// pooledSession 会话池中的单个会话
type pooledSession struct {
	SessionID     string
	MessageCount  int
	LastMessageID int
	CreatedAt     time.Time
	LastUsedAt    time.Time
	TTL           time.Duration
	ready         chan struct{} // 关闭表示会话已创建完成（sessionID 已填充）
}

func (s *pooledSession) isExpired() bool {
	return time.Since(s.CreatedAt) > s.TTL
}

func (s *pooledSession) isFull(maxMessages int) bool {
	return s.MessageCount >= maxMessages
}

// SessionPool 按账号管理会话复用
type SessionPool struct {
	mu       sync.RWMutex
	entries  map[string]*pooledSession // key: accountID
	stopCh   chan struct{}
	stopOnce sync.Once
}

// NewSessionPool 创建会话池并启动后台清理
func NewSessionPool() *SessionPool {
	p := &SessionPool{
		entries: make(map[string]*pooledSession),
		stopCh:  make(chan struct{}),
	}
	go p.cleanupLoop()
	return p
}

// Acquire 获取或创建会话。返回 sessionID 和 parentMessageID（0 表示新会话）。
//
// 并发安全：当多个 goroutine 同时为同一个 accountID 调用 Acquire 时，
// 第一个会创建占位符并返回空 sessionID，后续的会阻塞等待占位符被 Register 填充，
// 确保不会重复创建会话。
func (p *SessionPool) Acquire(accountID string, maxMessages int) (sessionID string, parentMessageID int) {
	if maxMessages <= 0 {
		maxMessages = defaultMaxMessagesPerSession
	}

	p.mu.Lock()
	entry, ok := p.entries[accountID]

	if ok {
		// 已有条目：检查是否正在创建中（ready 未关闭 = 占位符，sessionID 为空）
		if entry.ready != nil {
			ready := entry.ready
			p.mu.Unlock()
			// 等待创建完成或超时
			select {
			case <-ready:
				// 创建完成，重新获取
			case <-time.After(sessionCreateTimeout):
				// 超时：创建方可能失败了，清理并当作新会话
				config.Logger.Warn("[session_pool] acquire timeout waiting for session creation",
					"account", accountID)
				p.mu.Lock()
				if p.entries[accountID] == entry {
					delete(p.entries, accountID)
				}
				p.mu.Unlock()
				return "", 0
			}
			// 重新加锁获取
			p.mu.Lock()
			entry, ok = p.entries[accountID]
			if !ok {
				p.mu.Unlock()
				return "", 0
			}
		}

		if !entry.isExpired() && !entry.isFull(maxMessages) {
			// 复用现有会话
			entry.MessageCount++
			entry.LastUsedAt = time.Now()
			config.Logger.Debug("[session_pool] reusing session",
				"account", accountID, "session_id", entry.SessionID,
				"messages", entry.MessageCount, "parent_msg_id", entry.LastMessageID)
			sid := entry.SessionID
			pmid := entry.LastMessageID
			p.mu.Unlock()
			return sid, pmid
		}

		// 会话已满或过期，需要新建
		config.Logger.Debug("[session_pool] session full or expired, creating new",
			"account", accountID, "old_session", entry.SessionID,
			"messages", entry.MessageCount, "expired", entry.isExpired())
	}

	// 创建占位符：ready channel 未关闭 = 正在创建中
	entry = &pooledSession{
		SessionID:    "", // 待填充
		MessageCount: 1,
		CreatedAt:    time.Now(),
		LastUsedAt:   time.Now(),
		TTL:          defaultSessionTTL,
		ready:        make(chan struct{}),
	}
	p.entries[accountID] = entry
	p.mu.Unlock()
	return "", 0
}

// Register 注册新创建的会话 ID（Acquire 返回空时由调用方回填）。
// 关闭 ready channel 唤醒所有等待的 goroutine。
func (p *SessionPool) Register(accountID string, sessionID string) {
	if accountID == "" || sessionID == "" {
		return
	}
	p.mu.Lock()
	entry, ok := p.entries[accountID]
	if !ok {
		p.entries[accountID] = &pooledSession{
			SessionID:    sessionID,
			MessageCount: 1,
			CreatedAt:    time.Now(),
			LastUsedAt:   time.Now(),
			TTL:          defaultSessionTTL,
		}
		p.mu.Unlock()
		return
	}
	// 如果是从 Acquire 预占位的，回填 sessionID 并唤醒等待者
	if entry.SessionID == "" && entry.ready != nil {
		entry.SessionID = sessionID
		close(entry.ready)
		entry.ready = nil
		config.Logger.Debug("[session_pool] registered new session",
			"account", accountID, "session_id", sessionID)
	} else {
		// 并发冲突：另一个 goroutine 已经填充了不同的 sessionID
		config.Logger.Warn("[session_pool] register conflict, session already filled",
			"account", accountID, "existing_session", entry.SessionID, "new_session", sessionID)
	}
	p.mu.Unlock()
}

// Update 更新会话的最新消息 ID（用于下一次请求的 parent_message_id）
func (p *SessionPool) Update(accountID string, sessionID string, responseMessageID int) {
	if accountID == "" || sessionID == "" || responseMessageID <= 0 {
		return
	}
	p.mu.Lock()
	defer p.mu.Unlock()

	entry, ok := p.entries[accountID]
	if !ok || entry.SessionID != sessionID {
		config.Logger.Warn("[session_pool] update skipped, session mismatch",
			"account", accountID, "expected_session", sessionID,
			"pool_session", func() string {
				if ok {
					return entry.SessionID
				}
				return "<none>"
			}(),
			"msg_id", responseMessageID)
		return
	}
	entry.LastMessageID = responseMessageID
	entry.LastUsedAt = time.Now()
	config.Logger.Debug("[session_pool] updated last message id",
		"account", accountID, "session_id", sessionID, "msg_id", responseMessageID)
}

// Invalidate 使某个账号的会话失效（如账号被封禁或切换账号）
func (p *SessionPool) Invalidate(accountID string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	entry, ok := p.entries[accountID]
	if ok {
		// 如果正在创建中，关闭 ready channel 唤醒等待者
		if entry.ready != nil {
			close(entry.ready)
		}
	}
	delete(p.entries, accountID)
	config.Logger.Debug("[session_pool] invalidated", "account", accountID)
}

// Remove 删除指定会话条目
func (p *SessionPool) Remove(accountID string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	entry, ok := p.entries[accountID]
	if ok && entry.ready != nil {
		close(entry.ready)
	}
	delete(p.entries, accountID)
}

// Stop 停止后台清理
func (p *SessionPool) Stop() {
	p.stopOnce.Do(func() {
		close(p.stopCh)
	})
}

func (p *SessionPool) cleanupLoop() {
	ticker := time.NewTicker(poolCleanupInterval)
	defer ticker.Stop()
	for {
		select {
		case <-p.stopCh:
			return
		case <-ticker.C:
			p.cleanup()
		}
	}
}

func (p *SessionPool) cleanup() {
	p.mu.Lock()
	defer p.mu.Unlock()
	for id, entry := range p.entries {
		// 跳过正在创建中的占位符（由创建方负责）
		if entry.ready != nil {
			continue
		}
		if entry.isExpired() {
			config.Logger.Debug("[session_pool] cleanup expired session",
				"account", id, "session_id", entry.SessionID, "messages", entry.MessageCount)
			delete(p.entries, id)
		}
	}
}