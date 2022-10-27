1.common 下放项目公共的组件
2.utils 放model用到的组件,子文件夹可以用对应的model命名
3.app.toml配置文件放根目录，部署只需要拷贝filSign,app.toml，2个文件
4.启动 ./filSign
5,创建地址
```
curl -X POST --url  http://127.0.0.1:18072/v1/fil/createaddr -d '{"num":3,"coin_name":"fil","order_no":"12306","mch_name":"goapi"}'
```
6,签名
```
curl -X POST --url  http://127.0.0.1:18072/v1/fil/sign -d '{"coin_name":"fil","order_no":"12306","mch_name":"goapi","from_addr":"f1ixaijgeg4tsdepfgloljqboe65auz65alkfhkxq","to_addr":"f1s7qtoabqeh7qmfddpr7pdu6m63gthaufkaxsnhy","amount":"1","nonce":0,"gas_limit":4000000,"gas_premium":200000,"gas_fee_cap":100000000}'
{
	"coin_name": "fil",
	"order_no": "12306",
	"mch_name": "goapi",
	"from_addr": "f1ixaijgeg4tsdepfgloljqboe65auz65alkfhkxq",
	"to_addr": "f1s7qtoabqeh7qmfddpr7pdu6m63gthaufkaxsnhy",
	"amount": "1",
	"nonce": 0, //必填
	"gas_limit": 4000000,  //非必填,默认4000000
	"gas_premium": 200000, //非必填,默认200000
	"gas_fee_cap": 1000000000 //非必填,默认1000000000
}
```
6,热签名
```
curl -X POST --url  http://127.0.0.1:18072/v1/fil/sign -d '{"coin_name":"fil","order_no":"12306","mch_name":"goapi","from_addr":"f1ixaijgeg4tsdepfgloljqboe65auz65alkfhkxq","to_addr":"f1s7qtoabqeh7qmfddpr7pdu6m63gthaufkaxsnhy","amount":"1","nonce":0,"gas_limit":4000000,"gas_premium":200000,"gas_fee_cap":100000000}'
{
	"coin_name": "fil",
	"order_no": "12306",
	"mch_name": "goapi",
	"from_addr": "f1ixaijgeg4tsdepfgloljqboe65auz65alkfhkxq",
	"to_addr": "f1s7qtoabqeh7qmfddpr7pdu6m63gthaufkaxsnhy",
	"amount": "1",
	"nonce": 0, //非必填,自动获取
	"gas_limit": 4000000,  //非必填,默认4000000
	"gas_premium": 200000, //非必填,默认200000
	"gas_fee_cap": 1000000000 //非必填,默认1000000000
}
```


CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build;
ossutil cp -f ./filSign  oss://dfdfedfda;
ossutil sign oss://dfdfedfda/filSign    --timeout 1000;


sudo wget -O filSign  "http://dfdfedfda.oss-cn-hongkong.aliyuncs.com/filSign?Expires=1630735903&OSSAccessKeyId=LTAI8Z8YW8VGkEvg&Signature=rVWX1rfIEGuSyZNP3a7gIY1ZUKg%3D"

go build;
ossutil cp -f ./filSign  oss://dfdfedfda;
ossutil sign oss://dfdfedfda/filSign    --timeout 1000;

