package luna

import (
	"net/http"
	"testing"
	"time"
)

func Test_acc(t *testing.T){
	t.Log(GenAccount())
}
func Test_http(t *testing.T){
	for i:=0;i<100;i++ {
		req, err := http.Get("https://lcd.terra.dev/wasm/contracts/terra1rl20t79ffsrqfa29rke48tj05gj9jxumm92vg8/store?query_msg={%22balance%22:{%22address%22:%20%22terra15srwqp98mv36hf6cs9heu5pz9qa59hl8km5q2d%22}}")
		if err != nil {
			t.Fatal(err.Error())
		}
		if req.StatusCode !=200{
			t.Log("faild")
		}
		time.Sleep(time.Second)
	}
}