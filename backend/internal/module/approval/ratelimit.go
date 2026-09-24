package approval

import (
	"sync"
	"time"
)

// rateLimiter 固定窗口频控, 用于限制单个审批人的指令频率.
//
// 审批指令会触发数据库写入与 Bot 发送, 若机器人被恶意刷指令, 频控是必要的保护;
// 桶按 key(审批人会话)分开计数, 互相不影响.
type rateLimiter struct {
	mu      sync.Mutex
	window  time.Duration
	limit   int
	buckets map[string]*rateBucket
}

// rateBucket 某个 key 在当前窗口内的计数
type rateBucket struct {
	start time.Time
	count int
}

// newRateLimiter 创建固定窗口频控器
func newRateLimiter(limit int, window time.Duration) *rateLimiter {
	if limit <= 0 {
		limit = 5
	}
	if window <= 0 {
		window = 10 * time.Second
	}
	return &rateLimiter{
		window:  window,
		limit:   limit,
		buckets: make(map[string]*rateBucket),
	}
}

// allow 判断本次操作是否被允许; 被限流时同时返回建议的等待秒数
func (l *rateLimiter) allow(key string) (allowed bool, retryAfterSeconds int) {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	bucket, ok := l.buckets[key]
	if !ok || now.Sub(bucket.start) >= l.window {
		l.pruneLocked(now)
		l.buckets[key] = &rateBucket{start: now, count: 1}
		return true, 0
	}

	if bucket.count >= l.limit {
		left := int((l.window - now.Sub(bucket.start)).Seconds()) + 1
		return false, left
	}

	bucket.count++
	return true, 0
}

// pruneLocked 清理已过期的桶, 避免长期运行时 map 无界增长, 调用方必须持有锁
func (l *rateLimiter) pruneLocked(now time.Time) {
	for key, bucket := range l.buckets {
		if now.Sub(bucket.start) >= l.window {
			delete(l.buckets, key)
		}
	}
}
