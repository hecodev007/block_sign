package bifrost

import (
	"bncsign/common/conf"
	"encoding/json"
	"github.com/yanyushr/go-substrate-rpc-client/v3/client"
	"github.com/yanyushr/go-substrate-rpc-client/v3/rpc/state"
	"github.com/yanyushr/go-substrate-rpc-client/v3/types"
	"testing"
)

func Test_acc(t *testing.T) {
	pri, pub, err := GenerateKey()
	t.Log(pri, pub, err)
	addr, err := CreateAddress(pub, BNCPrefix)
	t.Log(addr, err)
}

func Test_Account_info(t *testing.T){
	Client, err := client.Connect("wss://bifrost-rpc.liebi.com/ws")
	if err != nil {
		t.Fatal(err.Error())
	}
	//Chain := chain.NewChain(Client)
	State := state.NewState(Client)
	address := "gG6gy9mTQSpkNenHy8PygX5GD8FCuPZmRdGse7uhKH8AN2B"
	var accountInfo types.AccountInfo

	pubKey := GetPublicFromAddr(address, BNCPrefix)
	entryMeta := types.StorageFunctionMetadataV13{}
	entryData :=[]byte("{\"Name\":\"Account\",\"Modifier\":{\"IsOptional\":false,\"IsDefault\":true,\"IsRequired\":false},\"Type\":{\"IsType\":false,\"AsType\":\"\",\"IsMap\":true,\"AsMap\":{\"Hasher\":{\"IsBlake2_128\":false,\"IsBlake2_256\":false,\"IsBlake2_128Concat\":true,\"IsTwox128\":false,\"IsTwox256\":false,\"IsTwox64Concat\":false,\"IsIdentity\":false},\"Key\":\"T::AccountId\",\"Value\":\"AccountInfo\\u003cT::Index, T::AccountData\\u003e\",\"Linked\":false},\"IsDoubleMap\":false,\"AsDoubleMap\":{\"Hasher\":{\"IsBlake2_128\":false,\"IsBlake2_256\":false,\"IsBlake2_128Concat\":false,\"IsTwox128\":false,\"IsTwox256\":false,\"IsTwox64Concat\":false,\"IsIdentity\":false},\"Key1\":\"\",\"Key2\":\"\",\"Value\":\"\",\"Key2Hasher\":{\"IsBlake2_128\":false,\"IsBlake2_256\":false,\"IsBlake2_128Concat\":false,\"IsTwox128\":false,\"IsTwox256\":false,\"IsTwox64Concat\":false,\"IsIdentity\":false}},\"IsNMap\":false,\"AsNMap\":{\"Keys\":null,\"Hashers\":null,\"Value\":\"\"}},\"Fallback\":\"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=\",\"Documentation\":[\" The full account information for a particular account ID.\"]}")
	json.Unmarshal(entryData,&entryMeta)
	key, err := types.CreateStorageKey(entryMeta, "System", "Account", pubKey, nil)
	if err != nil {
		t.Fatal()
	}

	_, err = State.GetStorageLatest(key, &accountInfo)
	if err != nil {
		t.Fatal(err.Error())
	}
	t.Log(accountInfo.Nonce,accountInfo.Data.Free.String())
}

func Test_chain(t *testing.T){
	Client, err := client.Connect(conf.GetConfig().Node.Url)
	if err != nil {
		t.Fatal(err.Error())
	}
	//Chain := chain.NewChain(Client)
	//State := state.NewState(Client)

	var res map[string]interface{}
	//err = Client.Call(&res, "chain_getBlock", "0x1868a075691a6a94283cea91cb8c30cc21be213e51e852be90b5c7ed6a7a2702")
	err = Client.Call(&res, "rpc_methods")
	if err != nil {
		t.Fatal(err.Error())
	}
	t.Log(res)
}