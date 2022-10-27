编译前要改包  go-address.TestnetPrefix = 'w'
github.com/filecoin-project/go-address
1.common 下放项目公共的组件
2.utils 放model用到的组件,子文件夹可以用对应的model命名
3.app.toml配置文件放根目录，部署只需要拷贝wdsign,app.toml，2个文件
4.启动 ./wdsign
5,创建地址
```bash
curl --location --request POST 'http://127.0.0.1:22026/v1/wd-wd/createaddr' \
--header 'Content-Type: application/json' \
--data-raw '{"coinName":"wd-wd","mchId":"635985570@qq.com","orderId":"20210706001","num":50}'

```


```bash
curl --location --request POST 'http://127.0.0.1:22026/v1/wd-wd/createaddr' \
--header 'Content-Type: application/json' \
--data-raw '{"coinName":"wd-wd","mchId":"hoo","orderId":"20210706002","num":20000}'

```

```bash
curl --location --request POST 'http://127.0.0.1:22026/v1/wd-wd/createaddr' \
--header 'Content-Type: application/json' \
--data-raw '{"coinName":"wd-wd","mchId":"hoo","orderId":"20210706001","num":20000}'

```

6,签名
```
curl -X POST --url  http://127.0.0.1:22026/v1/wd-wd/sign -d '{"coin_name":"fil","order_no":"12306","mch_name":"goapi","from_addr":"f1ixaijgeg4tsdepfgloljqboe65auz65alkfhkxq","to_addr":"f1s7qtoabqeh7qmfddpr7pdu6m63gthaufkaxsnhy","amount":"1","nonce":0,"gas_limit":4000000,"gas_premium":200000,"gas_fee_cap":100000000}'
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
curl -X POST --url  http://127.0.0.1:22026/v1/wd-wd/sign -d '{"coin_name":"fil","order_no":"12306","mch_name":"goapi","from_addr":"f1ixaijgeg4tsdepfgloljqboe65auz65alkfhkxq","to_addr":"f1s7qtoabqeh7qmfddpr7pdu6m63gthaufkaxsnhy","amount":"1","nonce":0,"gas_limit":4000000,"gas_premium":200000,"gas_fee_cap":100000000}'
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

zip ../wdsign.zip ../wdsign;
ossutil cp -f ../wdsign.zip  oss://dfdfedfda;
ossutil sign oss://dfdfedfda/wdsign.zip    --timeout 1000;


sudo wget -O wdsign.zip  "http://dfdfedfda.oss-cn-hongkong.aliyuncs.com/wdsign?Expires=1625546378&OSSAccessKeyId=LTAI8Z8YW8VGkEvg&Signature=AacvVX%2FYC4L7NPOdpVvPxOpe9L0%3D";

go build;
ossutil cp -f ./wdsign  oss://dfdfedfda;
ossutil sign oss://dfdfedfda/wdsign    --timeout 1000;


rm -rf ./wdsign;
sudo wget -O wdsign  "http://dfdfedfda.oss-cn-hongkong.aliyuncs.com/wdsign?Expires=1627278120&OSSAccessKeyId=LTAI8Z8YW8VGkEvg&Signature=EwD50i8D0qkn4JZOjdSko6MyqGo%3D";
sudo chmod +x ./wdsign

//私钥
w1qs5wnpdxqfdnw6utgjcngulax5f4pug2325vt2a
MTfjnOMNZH8pkvnvXGdDn7Ewx7LWB7urXlSFJkkZnZI=
w123456