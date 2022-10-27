####1.启动 ./moacSign
####2.私钥放在csv目录下
####3.app.toml配置文件放根目录，部署只需要拷贝moacSign,app.toml，2个文件
####4.common 下放项目公共的组件
####5.utils 放model用到的组件,子文件夹可以用对应的model命名

#测试用例：
#####1.创建账户:
>curl --location --request POST 'http://127.0.0.1:22022/v1/moac/createaddr' \
--header 'Content-Type: application/json' \
--data-raw '{"num":50,"coin_name":"moac","order_no":"test","mch_name":"635985570@qq.com"}'
#####返回值：
>
#####2.离线签名：
>curl --location --request POST 'http://127.0.0.1:22018/v1/wtc/sign' \
--header 'Content-Type: application/json' \
--data-raw '{
"coin_name": "wtc",
"order_no": "test",
"mch_name":"goapi",
"nonce": 0,
"from_address": "0xd44E9BDb4b7f8f54d3E85CfFe7Df326b897a7589",
"to_address": "0x0A9cEe1FE13788CC75F00c7BFD5A9e2b856274B2",
"token": "0x0A9cEe1FE13788CC75F00c7BFD5A9e2b856274B2",
"value": "1",
"gas_limit": 0,
"gas_price":"2"
}'


CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build;
ossutil cp -f ./moacSign  oss://dfdfedfda;
ossutil sign oss://dfdfedfda/moacSign    --timeout 1000;

sudo rm -rf ./moacSign;
sudo wget -O moacSign  "http://dfdfedfda.oss-cn-hongkong.aliyuncs.com/moacSign?Expires=1624266575&OSSAccessKeyId=LTAI8Z8YW8VGkEvg&Signature=8seVvWXev7a4KL%2BkyKIUoPKNZfA%3D"
sudo chmod 777 ./moacSign;

sudo pkill -9 moacSign;
sudo nohup ./moacSign &

