package xkutils

import "reflect"

type TryCatch struct {
	errChan      chan interface{}
	catches      map[reflect.Type]func(err error)
	defaultCatch func(err error)
}

type MyError struct {
	error
}

func (t TryCatch) Try(block func()) TryCatch {
	t.errChan = make(chan interface{})
	t.catches = map[reflect.Type]func(err error){}
	t.defaultCatch = func(err error) {}
	go func() {
		defer func() {
			t.errChan <- recover()
		}()
		block()
	}()
	return t
}

func (t TryCatch) CatchAll(block func(err error)) TryCatch {
	t.defaultCatch = block
	return t
}

func (t TryCatch) Catch(e error, block func(err error)) TryCatch {
	errorType := reflect.TypeOf(e)
	t.catches[errorType] = block
	return t
}

func (t TryCatch) Finally(block func()) TryCatch {
	err := <-t.errChan
	if err != nil {
		catch := t.catches[reflect.TypeOf(err)]
		if catch != nil {
			catch(err.(error))
		} else {
			t.defaultCatch(err.(error))
		}
	}
	block()
	return t
}
