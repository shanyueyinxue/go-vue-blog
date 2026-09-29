package utils

// ConcurrencyLimiter 非阻塞并发限制器
type ConcurrencyLimiter struct {
	sem chan struct{}
}

func NewConcurrencyLimiter(max int) *ConcurrencyLimiter {
	return &ConcurrencyLimiter{sem: make(chan struct{}, max)}
}

// TryAcquire 尝试获取许可，成功返回 release 函数，失败返回 false
func (l *ConcurrencyLimiter) TryAcquire() (release func(), ok bool) {
	select {
	case l.sem <- struct{}{}:
		return func() { <-l.sem }, true
	default:
		return nil, false
	}
}

/**
 * 示例：
 *
 * func main() {
 *     limiter := utils.NewConcurrencyLimiter(2)
 *     for i := 0; i < 10; i++ {
 *         go func() {
 *             release, ok := limiter.TryAcquire()
 *             if ok {
 *                 defer release()
 *                 // do something
 *             } else {
 *                 // do something else
 *             }
 *         }()
 *     }
 * }
 */
