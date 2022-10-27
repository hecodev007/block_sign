1.common 下放项目公共的组件
2.utils 放model用到的组件,子文件夹可以用对应的model命名
3.app.toml配置文件放根目录，部署只需要拷贝demo,app.toml，2个文件
4.启动 ./neoSign
5.创建账户地址
```
curl -X POST  --url http://127.0.0.1:18075/v1/neo/createaddr -d '{"num":10,"coin_name":"neo","order_no":"test","mch_name":"goapi"}'
curl -X POST --url http://127.0.0.1:18075/v1/neo/createaddr -d '{"num":1,"order_no":"test","mch_name":"hoo","coin_name":"neo"}'

curl -X POST --url http://127.0.0.1:18075/v1/neo/sign -d '{"coinName":"dash","orderId":"test","mchId":"goapi","txIns":[{"fromAddr":"AUnkg6U6hoDo7WCYpe6TUf9LhDp4JCSnFK","fromTxid":"c769217050a4ba70e748e66be4b6013c603a6566b090b8d7d76d0634a2d78087","FromIndex":0,"fromAmount":100000000}],"txOuts":[{"toAddr":"ASss3bMFeeVoN9noMzVXaCMukSjn7YhDia","toAmount":100000000}]}'

curl --location --request POST 'http://127.0.0.1:18075/v1/neo/createaddr' \
--header 'Content-Type: application/json' \
--data-raw '{"num":10,"coin_name":"neo","order_no":"test","mch_name":"goapi"}'
```
6.签名
```
curl -X POST --url  http://127.0.0.1:18075/v1/neo/sign -d '{"coinName":"dash","orderId":"test","mchId":"goapi","txIns":[{"fromAddr":"AUnkg6U6hoDo7WCYpe6TUf9LhDp4JCSnFK","fromTxid":"c769217050a4ba70e748e66be4b6013c603a6566b090b8d7d76d0634a2d78087","FromIndex":0,"fromAmount":100000000}],"txOuts":[{"toAddr":"ASss3bMFeeVoN9noMzVXaCMukSjn7YhDia","toAmount":100000000}]}'
```
7,签名并上链
```
curl -X POST -basic -u rylink:rylink@telos@2020 --url  http://127.0.0.1:18074/v1/dash/transfer -d '
{
	"coinName": "neo",
	"orderId": "test",
	"mchId": "goapi",
	"txIns": [
		{
			"fromAddr": "AUnkg6U6hoDo7WCYpe6TUf9LhDp4JCSnFK",
			"fromTxid": "c769217050a4ba70e748e66be4b6013c603a6566b090b8d7d76d0634a2d78087",
			"FromIndex": 0,
			"fromAmount": 100000000
		},
		{
			"fromAddr": "AKM4mRWWFWZBVgxk18cBy2jYJT9s6aZCty",
			"fromTxid": "7cf0819ec16093d0b327d7a293838bc660aee0b497d651fa699cd816b4200558",
			"FromIndex": 0,
			"fromAmount": 100000000
		}
	],
	"txOuts": [
		{
			"toAddr": "ASss3bMFeeVoN9noMzVXaCMukSjn7YhDia",
			"toAmount": 200000000
		}
	]
}
'
```


