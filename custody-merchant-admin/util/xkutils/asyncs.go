package xkutils

func AsyncBackCall(call func()) {
	go func() {
		call()
	}()
}
