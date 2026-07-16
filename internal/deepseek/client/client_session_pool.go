package client

import (
	"strings"
	"sync"
	"time"

	"ds2api/internal/config"
)

const (
	// 默认每个会话最大消息数
	defaultMaxMessagesPerSession = 50
	// 会话池过期清理间隔
	poolCleanupInterval = 5 * time.Minute
	// 会话默认 TTL（72小时，DeepSeek 会话本身 TTL 为 3 天）
	defaultSessionTTL = 72 * time.Hour
	// 等待会话创建的超时时间
	sessionCreateTimeout = 30 * time.Second
)

// poolKey 构建会话池的复合键：accountID + ":" + modelType
// 确保不同 model_type 的会话相互隔离，避免 DeepSeek 上游因 session
// 绑定的 model_type 与请求 payload 中的 model_type 不一致而路由到错误模型。
func poolKey(accountID, modelType string) string {
	return accountID + ":" + modelType
}

// pooledSession 会话池中的单个会话
type pooledSession struct {
	SessionID    string
	ModelType    string
	MessageCount int
	LastMessageID int
	CreatedAt time.Time
	LastUsedAt time.Time
	TTL time.Duration
	ready chan struct{} // 关闭表示会话已创建完成（sessionID 已填充）
}

func (s *pooledSession) isExpired() bool {
	return time.Since(s.CreatedAt) > s.TTL
}

func (s *pooledSession) isFull(maxMessages int) bool {
	return s.MessageCount >= maxMessages
}

// SessionPool 按账号+模型类型管理会话复用
type SessionPool struct {
	mu sync.RWMutex
	entries map[string]*pooledSession // key: accountID:modelType
	stopCh chan struct{}
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
// modelType 用于区分不同模型类型的会话（default/expert/vision），确保
// DeepSeek 上游 session 绑定的 model_type 与请求 payload 一致。
//
// 并发安全：当多个 goroutine 同时为同一个 key 调用 Acquire 时，
// 第一个会创建占位符并返回空 sessionID，后续的会阻塞等待占位符被 Register 填充，
// 确保不会重复创建会话。
func (p *SessionPool) Acquire(accountID string, modelType string, maxMessages int) (sessionID string, parentMessageID int) {
	if maxMessages <= 0 {
		maxMessages = defaultMaxMessagesPerSession
	}

	key := poolKey(accountID, modelType)

	p.mu.Lock()
	entry, ok := p.entries[key]

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
					"account", accountID, "model_type", modelType)
				p.mu.Lock()
				if p.entries[key] == entry {
					delete(p.entries, key)
				}
				p.mu.Unlock()
				return "", 0
			}
			// 重新加锁获取
			p.mu.Lock()
			entry, ok = p.entries[key]
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
				"account", accountID, "model_type", modelType, "session_id", entry.SessionID,
				"messages", entry.MessageCount, "parent_msg_id", entry.LastMessageID)
			sid := entry.SessionID
			pmid := entry.LastMessageID
			p.mu.Unlock()
			return sid, pmid
		}

		// 会话已满或过期，需要新建
		config.Logger.Debug("[session_pool] session full or expired, creating new",
			"account", accountID, "model_type", modelType, "old_session", entry.SessionID,
			"messages", entry.MessageCount, "expired", entry.isExpired())
	}

	// 创建占位符：ready channel 未关闭 = 正在创建中
	entry = &pooledSession{
		SessionID:    "", // 待填充
		ModelType:    modelType,
		MessageCount: 1,
		CreatedAt:    time.Now(),
		LastUsedAt:   time.Now(),
		TTL:          defaultSessionTTL,
		ready:        make(chan struct{}),
	}
	p.entries[key] = entry
	p.mu.Unlock()
	return "", 0
}

// Register 注册新创建的会话 ID（Acquire 返回空时由调用方回填）。
// 关闭 ready channel 唤醒所有等待的 goroutine。
func (p *SessionPool) Register(accountID string, modelType string, sessionID string) {
	if accountID == "" || sessionID == "" {
		return
	}
	key := poolKey(accountID, modelType)
	p.mu.Lock()
	entry, ok := p.entries[key]
	if !ok {
		p.entries[key] = &pooledSession{
			SessionID:    sessionID,
			ModelType:    modelType,
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
		entry.ModelType = modelType
		close(entry.ready)
		entry.ready = nil
		config.Logger.Debug("[session_pool] registered new session",
			"account", accountID, "model_type", modelType, "session_id", sessionID)
	} else {
		// 并发冲突：另一个 goroutine 已经填充了不同的 sessionID
		config.Logger.Warn("[session_pool] register conflict, session already filled",
			"account", accountID, "existing_session", entry.SessionID, "new_session", sessionID)
	}
	p.mu.Unlock()
}

// Update 更新会话的最新消息 ID（用于下一次请求的 parent_message_id）
func (p *SessionPool) Update(accountID string, modelType string, sessionID string, responseMessageID int) {
	if accountID == "" || sessionID == "" || responseMessageID <= 0 {
		return
	}
	key := poolKey(accountID, modelType)
	p.mu.Lock()
	defer p.mu.Unlock()

	entry, ok := p.entries[key]
	if !ok || entry.SessionID != sessionID {
		config.Logger.Warn("[session_pool] update skipped, session mismatch",
			"account", accountID, "model_type", modelType, "expected_session", sessionID,
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
		"account", accountID, "model_type", modelType, "session_id", sessionID, "msg_id", responseMessageID)
}

// Invalidate 使指定账号+模型类型的会话失效。
// 如果 modelType 为空，则使该账号下所有模型类型的会话失效。
func (p *SessionPool) Invalidate(accountID string, modelType string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if modelType != "" {
		key := poolKey(accountID, modelType)
		entry, ok := p.entries[key]
		if ok && entry.ready != nil {
			close(entry.ready)
		}
		delete(p.entries, key)
		config.Logger.Debug("[session_pool] invalidated", "account", accountID, "model_type", modelType)
		return
	}
	// Invalidate all entries for the account
	prefix := accountID + ":"
	for key, entry := range p.entries {
		if strings.HasPrefix(key, prefix) {
			if entry.ready != nil {
				close(entry.ready)
			}
			delete(p.entries, key)
		}
	}
	config.Logger.Debug("[session_pool] invalidated all", "account", accountID)
}

// Remove 删除指定账号+模型类型的会话条目
func (p *SessionPool) Remove(accountID string, modelType string) {
	key := poolKey(accountID, modelType)
	p.mu.Lock()
	defer p.mu.Unlock()
	entry, ok := p.entries[key]
	if ok && entry.ready != nil {
		close(entry.ready)
	}
	delete(p.entries, key)
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
