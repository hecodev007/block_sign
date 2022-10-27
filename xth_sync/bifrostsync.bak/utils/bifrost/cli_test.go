package bifrost

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"reflect"
	"testing"

	"github.com/centrifuge/go-substrate-rpc-client/v3/types"
)

func Test_cli(t *testing.T) {
	log.SetFlags(log.Llongfile)
	cli, err := NewClient("wss://bifrost-rpc.liebi.com/ws")
	if err != nil {
		t.Fatal(err.Error())
	}
	hash, _ := types.NewHashFromHexString("0x9554e1675361d64d83122d2abff9381caccf61632663867664e056c203dd4196")
	meta, err := cli.Api.RPC.State.GetMetadata(hash)
	if err != nil {
		t.Fatal(err.Error())
	}
	ioutil.WriteFile("./meta.json", []byte(String(meta.AsMetadataV13.Modules)), 0777)
	for _, v := range meta.AsMetadataV13.Modules {
		t.Log(v.Name)
		if v.HasEvents {
			ioutil.WriteFile("./"+string(v.Name)+"_meta.json", []byte(String(v)), 0777)
		}
	}
	return

	t.Log(cli.GetBestHeight())
	i := int64(640174)
	//for i:=int64(640170);i<=647000;i++ {
	block, err := cli.GetBlockByNum(i)
	if err != nil {
		t.Fatal(err.Error())
	}
	t.Log(String(block))
	//}
}

func Test_decode(t *testing.T) {
	log.SetFlags(log.Llongfile)
	mainnet := "Bifrost"
	cli, err := NewClient("wss://bifrost-rpc.liebi.com/ws")
	if err != nil {
		t.Fatal(err.Error())
	}
	//hash,_ := types.NewHashFromHexString("0x9554e1675361d64d83122d2abff9381caccf61632663867664e056c203dd4196")
	meta, err := cli.Api.RPC.State.GetMetadataLatest()
	if err != nil {
		t.Fatal(err.Error())
	}
	if !meta.IsMetadataV13 {
		t.Fatal("只支持meta13")
	}
	var EventRecords [][]string

	EventStructs := make(map[string][]types.Type)

	for _, mod := range meta.AsMetadataV13.Modules {
		if !mod.HasEvents {
			continue
		}
		for _, event := range mod.Events {
			var tmpEventRecords []string
			tmpEventRecords = append(tmpEventRecords, string(mod.Name)+"_"+string(event.Name))
			if !StructInDefault(tmpEventRecords[0]) {
				//log.Println(tmpEventRecords[0])
				continue
			}
			tmpEventRecords = append(tmpEventRecords, "Event"+string(mod.Name)+string(event.Name))
			EventRecords = append(EventRecords, tmpEventRecords)
			EventStructs["Event"+string(mod.Name)+string(event.Name)] = event.Args
		}
	}
	var EventRecordsText string
	EventRecordsText = "type " + mainnet + "EventRecords struct {\n"
	for _, v := range EventRecords {
		EventRecordsText += v[0] + "		[]" + v[1] + "\n"
	}
	EventRecordsText += "}\n"
	//t.Log(EventRecordsText)

	for k, v := range EventStructs {

		ToStruct(k, v)
	}

}

var num = make(map[string]int)

func ToStruct(name string, elems []types.Type) (ret string, err error) {
	ret = "type " + name + " struct {\n"
	ret += "Phase         types.Phase\n"

	for _, v := range elems {
		vstr := string(v)
		if _, ok := num[vstr]; ok {
			continue
		}

		num[vstr] = 1
		log.Println(name, vstr)
	}

	ret += "}\n"
	return ret, nil
}
func Test_refect(t *testing.T) {
	record := types.EventRecords{}
	fv := reflect.TypeOf(record)
	n := fv.NumField()
	for i := 0; i < n; i++ {
		t.Log(fv.Field(i).Name)
	}
}
func StructInDefault(name string) bool {
	record := types.EventRecords{}
	fv := reflect.TypeOf(record)
	n := fv.NumField()
	for i := 0; i < n; i++ {
		if name == fv.Field(i).Name {
			return true
		}
	}
	return false
}
func String(d interface{}) string {
	str, _ := json.Marshal(d)
	return string(str)
}
