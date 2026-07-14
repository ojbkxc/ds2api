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
)

// pooledSession 会话池中的单个会话
type pooledSession struct {
	SessionID       string
	MessageCount    int
	LastMessageID   int
	CreatedAt       time.Time
	LastUsedAt      time.Time
	TTL             time.Duration
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

// Acquire 获取或创建会话。返回 sessionID 和 parentMessageID（0 表示新会话）
func (p *SessionPool) Acquire(accountID string, maxMessages int) (sessionID string, parentMessageID int) {
	if maxMessages <= 0 {
		maxMessages = defaultMaxMessagesPerSession
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	entry, ok := p.entries[accountID]
	if ok && !entry.isExpired() && !entry.isFull(maxMessages) {
		// 复用现有会话
		entry.MessageCount++
		entry.LastUsedAt = time.Now()
		config.Logger.Debug("[session_pool] reusing session",
			"account", accountID, "session_id", entry.SessionID,
			"messages", entry.MessageCount, "parent_msg_id", entry.LastMessageID)
		return entry.SessionID, entry.LastMessageID
	}

	// 需要新会话：返回空 sessionID 让调用方创建，parentMessageID=0
	if ok {
		config.Logger.Debug("[session_pool] session full or expired, creating new",
			"account", accountID, "old_session", entry.SessionID,
			"messages", entry.MessageCount, "expired", entry.isExpired())
	}
	// 先占位，避免并发重复创建
	p.entries[accountID] = &pooledSession{
		SessionID:    "", // 待填充
		MessageCount: 1,
		CreatedAt:    time.Now(),
		LastUsedAt:   time.Now(),
		TTL:          defaultSessionTTL,
	}
	return "", 0
}

// Register 注册新创建的会话 ID（Acquire 返回空时由调用方回填）
func (p *SessionPool) Register(accountID string, sessionID string) {
	if accountID == "" || sessionID == "" {
		return
	}
	p.mu.Lock()
	defer p.mu.Unlock()

	entry, ok := p.entries[accountID]
	if !ok {
		p.entries[accountID] = &pooledSession{
			SessionID:    sessionID,
			MessageCount: 1,
			CreatedAt:    time.Now(),
			LastUsedAt:   time.Now(),
			TTL:          defaultSessionTTL,
		}
		return
	}
	// 如果是从 Acquire 预占位的，回填 sessionID
	if entry.SessionID == "" {
		entry.SessionID = sessionID
	}
	config.Logger.Debug("[session_pool] registered new session",
		"account", accountID, "session_id", sessionID)
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
		return
	}
	entry.LastMessageID = responseMessageID
	entry.LastUsedAt = time.Now()
	config.Logger.Debug("[session_pool] updated last message id",
		"account", accountID, "session_id", sessionID, "msg_id", responseMessageID)
}

// Invalidate 使某个会话失效（如遇到错误需要重建）
func (p *SessionPool) Invalidate(accountID string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	delete(p.entries, accountID)
	config.Logger.Debug("[session_pool] invalidated", "account", accountID)
}

// Remove 删除指定会话条目
func (p *SessionPool) Remove(accountID string) {
	p.mu.Lock()
	defer p.mu.Unlock()
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
		if entry.isExpired() {
			config.Logger.Debug("[session_pool] cleanup expired session",
				"account", id, "session_id", entry.SessionID, "messages", entry.MessageCount)
			delete(p.entries, id)
		}
	}
}