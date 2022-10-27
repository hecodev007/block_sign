package utils

import (
	"sync"
)

type TokenPool struct {
	capacity int
	tokens   chan struct{}
	lock     sync.Mutex
	wg       sync.WaitGroup
}

//创建并初始化令牌池
func NewTokenPool(capacity int) *TokenPool {
	tokens := make(chan struct{}, capacity)
	//for i := 0; i < capacity; i++ {
	//	tokens <- struct{}{}
	//}
	return &TokenPool{
		capacity,
		tokens,
		sync.Mutex{},
		sync.WaitGroup{}}
}

//消费指定数量的令牌,如果没有则会等待
func (t *TokenPool) SpendToken(n int) error {
	t.lock.Lock()
	defer t.lock.Unlock()

	for i := 0; i < n; i++ {
		t.tokens <- struct{}{}
		t.wg.Add(1)
	}
	return nil
}

//归还指定数量的令牌令牌
func (t *TokenPool) ReturnToken(n int) error {
	t.lock.Lock()
	defer t.lock.Unlock()

	for i := 0; i < n; i++ {
		<-t.tokens
		t.wg.Done()
	}
	return nil
}

//释放所有令牌
func (t *TokenPool) WaitForDown() {
	t.wg.Wait()
}
