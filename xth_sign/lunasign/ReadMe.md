1.common 下放项目公共的组件

2.utils 放model用到的组件,子文件夹可以用对应的model命名

3.app.toml配置文件放根目录，部署只需要拷贝lunasign,app.toml，2个文件

4.启动 ./lunasign;

5,创建账户
```
curl -X POST --url  http://127.0.0.1:14032/v1/luna/genaddr -d '{"num":1,"coin_name":"atom","order_no":"btc","mch_name":"goapi"}'
curl -X POST --url  http://127.0.0.1:14032/v1/luna/genaddr -d '{"num":1,"coin_name":"atom","order_no":"btc","mch_name":"hoo"}'
curl -X POST --url  http://127.0.0.1:14032/v1/luna/genaddr -d '{"num":1,"coin_name":"atom","order_no":"btc","mch_name":"635985570@qq.com"}'

```
6,签名
```
curl -X POST --url  http://127.0.0.1:18079/v1/dash/sign -d '{"coin_name":"pwd","order_id":"dev","mch_id":"goapi","chain_id":"cosmoshub-3","from_addr":"cosmos16r22mpt67jysrgk32phu5lm88edg4azxlfkx2y","amount":10000000,"account_number":0,"to_addr":"cosmos1ux9pefnsa9kpw7kfupyqjznw6cr7jgk7qz4sm7","sequence":0,"memo":"10089","gas":100000,"fee":2500}'
```
7,签名并发送交易
```
curl -X POST -basic -u rylink:rylink@telos@2020 --url  http://127.0.0.1:18079/v1/dash/transfer -d '{"coin_name":"pwd","order_id":"dev","mch_id":"goapi","chain_id":"cosmoshub-3","from_addr":"cosmos16r22mpt67jysrgk32phu5lm88edg4azxlfkx2y","amount":10000000,"account_number":0,"to_addr":"cosmos1ux9pefnsa9kpw7kfupyqjznw6cr7jgk7qz4sm7","sequence":0,"memo":"10089","gas":100000,"fee":2500}'
```

go build -o lunasign  -ldflags '-linkmode "external" -extldflags "-static"' main.go
go build -o lunasign  -ldflags '-linkmode "external" main.go
gcc  -lwasmvm --verbose;

go build;ossutil cp -f ./lunasign  oss://dfdfedfda;ossutil sign oss://dfdfedfda/lunasign    --timeout 1000;


ossutil cp -f ./lib.tar oss://dfdfedfda;
ossutil sign oss://dfdfedfda/lib.tar    --timeout 1000; 
sudo wget -O lib.tar "http://dfdfedfda.oss-cn-hongkong.aliyuncs.com/lib.tar?Expires=1647244512&OSSAccessKeyId=LTAI8Z8YW8VGkEvg&Signature=3YyP5XCtz4sgW%2BhY7sWACWi0h7E%3D";

sudo wget -O lunasign  "http://dfdfedfda.oss-cn-hongkong.aliyuncs.com/lunasign?Expires=1648434505&OSSAccessKeyId=LTAI8Z8YW8VGkEvg&Signature=2IPJrOeIaGx9zxjcH%2F8gulJGfEk%3D";
sudo chmod +x ./lunasign;

