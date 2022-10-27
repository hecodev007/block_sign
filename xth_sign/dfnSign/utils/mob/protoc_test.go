package mob

import (

	"context"
	"encoding/hex"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
	mobapi "dfnSign/utils/mob/protoc"
	"net/rpc"
	"testing"
)
func Test_cli_yblock(t *testing.T){
	pwd,err := hex.DecodeString("c7f04fcd40d093ca6578b13d790df0790c96e94a77815e5052993af1b9d12923")
	if err != nil{
		panic(err.Error())
	}

	conn,err := grpc.Dial("18.182.64.34:24444",grpc.WithInsecure())
	if err != nil {
		panic(err.Error())
	}
	t.Log(conn.GetState().String())
	client :=mobapi.NewMobilecoindAPIClient(conn)
	req := &mobapi.SetDbPasswordRequest{
		Password: pwd,
	}
	res ,err := client.SetDbPassword(context.Background(), req)
	//res ,err := client.UnlockDb(context.Background(), req)
	if err != nil {
		t.Error(err.Error())
	}
	_ ,err = client.GetLedgerInfo(context.Background(), &emptypb.Empty{})
	if err != nil {
		panic(err.Error())
	}
	t.Log(res)
}
func Test_cli(t *testing.T){
	conn,err := grpc.Dial("18.182.64.34:24444",grpc.WithInsecure())
	if err != nil {
		panic(err.Error())
	}
	t.Log(conn.GetState().String())
	client :=mobapi.NewMobilecoindAPIClient(conn)
	response ,err := client.GetLedgerInfo(context.Background(), &emptypb.Empty{})
	if err != nil {
		panic(err.Error())
	}

	t.Log(response)
}
func Test_prc(t *testing.T){
	client, err := rpc.Dial("tcp", "18.182.64.34:24444")
	if err != nil {
		t.Error("dialing:", err)
	}

	var reply = &mobapi.GetLedgerInfoResponse{
	}
	//var param = &go_protoc.String{
	//	Value:"hello",
	//}
	//var params  = new(struct{})

	err = client.Call("MobilecoindAPI.GetLedgerInfo", &struct{}{}, reply)
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(reply)
}