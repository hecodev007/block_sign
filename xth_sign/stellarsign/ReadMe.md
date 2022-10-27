####1.启动 ./telosSign
####2.私钥放在csv目录下
####3.app.toml配置文件放根目录，部署只需要拷贝telosSign,app.toml，2个文件
####4.common 下放项目公共的组件
####5.utils 放model用到的组件,子文件夹可以用对应的model命名

#测试用例：
#####1.创建账户:
>curl -X POST --url http://127.0.0.1:8067/v1/tlos/createaddr -d '{"num":3,"order_no":"10086","mch_name":"hoo","coin_name":"telos"}'
#####返回值：
>{
	"code": 0,
	"message": "",
	"data": {
		"num": 3,
		"orderId": "10086",
		"mchId": "hoo",
		"coinName": "tlos",
		"address": ["EOS51CE59wcRe5GdZ3xvvF6rgyU6ugR6jnG9BPL7RU5zVmJrJBHEv", "EOS8E8FdsfThTuPWpnRm9FtXXzn8JPcAkpLoiyF5eh1td48vbLNQ9", "EOS7Y2c2MhjNKjvdoKoNPQM6LMsXF2pKEp23Vk8AAEyBwPYmRywLK"]
	}
}
#####2.离线签名：
>curl -X POST --url http://127.0.0.1:8069/v1/tlos/sign -d '{"mch_id":1063,"order_no":"10086","mch_name":"635985570@qq.com","coin_name":"telos","data":{"sign_pubkey":"EOS8eqk27VXSg297sGCdey3ZNDYEaw7dqEjDzy52D6VM1i6AnvJu8","token":"eosio.token","from_address":"xutonghua123","memo":"test","quantity":"0.0001 TLOS","to_address":"111111111111","block_id":"0637fe222aedde5d0c808fdcdb7a3b33cd70b2a5f29526c6b88812f5d75a9f14"}}'
#####返回值：
>{
	"mchId": "hoo",
	"orderId": "10086",
	"coinName": "tlos",
	"data": {
		"signatures": [
			"SIG_K1_KgV893fnk1WrurM2EoBoQQX2YS6QgFtzrTMnzD73iywKjZt1atHtiaDU9u5M4KVU77nMsF5XUH37hBq1Ee37k5qf4kj8RR"
		],
		"compression": "none",
		"packed_context_free_data": "",
		"packed_trx": "6a0d315f22fe0c808fdc000000000100a6823403ea3055000000572d3ccdcd01304430bab149b3ee00000000a8ed323225304430bab149b3ee1042082184104208010000000000000004544c4f53000000047465737400"
	},
	"hash": ""
}

#####1.创建账户:
curl --location --request POST 'http://127.0.0.1:22004/v1/okt/createaddr' \
--header 'Content-Type: application/json' \
--data-raw '{"num":50,"coin_name":"test","order_no":"20210804002","mch_name":"635985570@qq.com"}'

curl --location --request POST 'http://127.0.0.1:22004/v1/okt/createaddr' \
--header 'Content-Type: application/json' \
--data-raw '{"num":10000,"coin_name":"test","order_no":"20210804001","mch_name":"hoo"}'

curl --location --request POST 'http://127.0.0.1:22004/v1/okt/createaddr' \
--header 'Content-Type: application/json' \
--data-raw '{"num":10000,"coin_name":"test","order_no":"20210804002","mch_name":"hoo"}'


curl --location --request POST 'http://127.0.0.1:22004/v1/okt/createaddr' \
--header 'Content-Type: application/json' \
--data-raw '{"num":10000,"coin_name":"test","order_no":"20210804003","mch_name":"hoo"}'


curl --location --request POST 'http://127.0.0.1:22004/v1/okt/createaddr' \
--header 'Content-Type: application/json' \
--data-raw '{"num":10000,"coin_name":"test","order_no":"20210804004","mch_name":"hoo"}'


go build;
ossutil cp -f ./stellarsign  oss://dfdfedfda;
ossutil sign oss://dfdfedfda/stellarsign    --timeout 1000;

pkill -9 stellarsign;
rm -rf ./stellarsign;
sudo wget -O stellarsign  "http://dfdfedfda.oss-cn-hongkong.aliyuncs.com/stellarsign?Expires=1632652384&OSSAccessKeyId=LTAI8Z8YW8VGkEvg&Signature=WadG0GETuWofEdojrrWk0rSIQHs%3D";
sudo chmod +x ./stellarsign;
nohup ./stellarsign &