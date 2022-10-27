1.common 下放项目公共的组件

2.utils 放model用到的组件,子文件夹可以用对应的model命名

3.app.toml配置文件放根目录，部署只需要拷贝terrasign,app.toml，2个文件

4.启动 ./terrasign;

5,创建账户
```
curl -X POST --url  http://127.0.0.1:18079/v1/atom/createaddr -d '{"num":1,"coin_name":"atom","order_no":"btc","mch_name":"goapi"}'
```
6,签名
```
curl -X POST --url  http://127.0.0.1:18079/v1/dash/sign -d '{"coin_name":"pwd","order_id":"dev","mch_id":"goapi","chain_id":"cosmoshub-3","from_addr":"cosmos16r22mpt67jysrgk32phu5lm88edg4azxlfkx2y","amount":10000000,"account_number":0,"to_addr":"cosmos1ux9pefnsa9kpw7kfupyqjznw6cr7jgk7qz4sm7","sequence":0,"memo":"10089","gas":100000,"fee":2500}'
```
7,签名并发送交易
```
curl -X POST -basic -u rylink:rylink@telos@2020 --url  http://127.0.0.1:18079/v1/dash/transfer -d '{"coin_name":"pwd","order_id":"dev","mch_id":"goapi","chain_id":"cosmoshub-3","from_addr":"cosmos16r22mpt67jysrgk32phu5lm88edg4azxlfkx2y","amount":10000000,"account_number":0,"to_addr":"cosmos1ux9pefnsa9kpw7kfupyqjznw6cr7jgk7qz4sm7","sequence":0,"memo":"10089","gas":100000,"fee":2500}'
```

CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build;ossutil cp -f ./terrasign  oss://dfdfedfda;ossutil sign oss://dfdfedfda/terrasign    --timeout 1000;




sudo wget -O kava-sign  "http://dfdfedfda.oss-cn-hongkong.aliyuncs.com/terrasign?Expires=1632376987&OSSAccessKeyId=LTAI8Z8YW8VGkEvg&Signature=yJUyJM76aZMwMQ0PGQJWGtFVyos%3D";
sudo chmod +x ./kava-sign;