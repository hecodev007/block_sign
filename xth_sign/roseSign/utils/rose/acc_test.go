package rose

import "testing"

func Test_acc(t *testing.T){
	//acc_test.go:6: 6081efb4d9329fcc7b6b083c3c6343cc49ffd7af087508584c690f0e4e394f52
	//oasis1qzv9v8cseda43z6v76yphjap22s4yuv65yszztx9
	t.Log(GenAccount())
	t.Log(PriToAddr("6081efb4d9329fcc7b6b083c3c6343cc49ffd7af087508584c690f0e4e394f52"))
	t.Log(BuildTx("","oasis1qzv9v8cseda43z6v76yphjap22s4yuv65yszztx9",0,0))
}