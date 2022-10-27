1.common 下放项目公共的组件

2.utils 放model用到的组件,子文件夹可以用对应的model命名

3.app.toml配置文件放根目录，部署只需要拷贝xlmSign,app.toml，2个文件

4.启动 ./xlmSign;

5,创建账户
```
curl -X POST --url  http://127.0.0.1:18096/v1/xlm/createaddr -d '{"num":1,"coin_name":"xlm","order_no":"test","mch_name":"goapi"}'
```
6,签名
```
curl -X POST --url  http://127.0.0.1:18095/v1/bcha/sign -d '{"coinName":"dash","orderId":"test","mchId":"goapi","txIns":[{"fromAddr":"XawwU58trgyp8wMQtPRkcQUCxJQpwmCv31","fromTxid":"3f70bbbef46bbd6ad70e9c7308d9e57e05c5b1804bcd1fe369570fca0388d176","FromIndex":0,"fromAmount":2800000}],"txOuts":[{"toAddr":"XawwU58trgyp8wMQtPRkcQUCxJQpwmCv31","toAmount":2700000},{"toAddr":"7pG9k7MmH54GTx5aFBCxim9sCEAsQwU3Bt","toAmount":89999}]}'
```
7,签名并发送交易
```
curl -X POST -basic -u rylink:rylink@telos@2020 --url  http://127.0.0.1:18078/v1/biw/transfer -d '{"coinName":"dash","orderId":"test","mchId":"goapi","txIns":[{"fromAddr":"XawwU58trgyp8wMQtPRkcQUCxJQpwmCv31","fromTxid":"3f70bbbef46bbd6ad70e9c7308d9e57e05c5b1804bcd1fe369570fca0388d176","FromIndex":0,"fromAmount":2800000}],"txOuts":[{"toAddr":"XawwU58trgyp8wMQtPRkcQUCxJQpwmCv31","toAmount":2700000},{"toAddr":"7pG9k7MmH54GTx5aFBCxim9sCEAsQwU3Bt","toAmount":89999}]}'
```
curl --location --request POST 'http://127.0.0.1:18096/v1/xlm/trustline' \
--header 'Content-Type: application/json' \
--data-raw '{
"orderId": "test",
"mch_name": "635985570@qq.com",
"address":"GCU4W74RDYNM2WZQ7LROGRDZHMK4YT3JTVSSEANVAZ4ZXOJV6VMDMVXO",
"token":"LSP-GAB7STHVD5BDH3EEYXPI3OM7PCS4V443PYB5FNT6CFGJVPDLMKDM24WK"
}'



curl --location --request POST 'http://127.0.0.1:18096/v1/xlm/trustline' \
--header 'Content-Type: application/json' \
--data-raw '{
"orderId": "test",
"mch_name": "hoo",
"address":"GABQ3CP6KRXFXVHGTBVCB7KGFBUGLRJHYPI3JIFFRNYDOUCZULPUGRMT",
"token":"LSP-GAB7STHVD5BDH3EEYXPI3OM7PCS4V443PYB5FNT6CFGJVPDLMKDM24WK"
}'


CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build;ossutil cp -f ./xlmSign  oss://dfdfedfda;ossutil sign oss://dfdfedfda/xlmSign    --timeout 1000;

pkill -9 xlmSign;
rm -rf xlmSign;
sudo wget -O xlmSign  "http://dfdfedfda.oss-cn-hongkong.aliyuncs.com/xlmSign?Expires=1642126094&OSSAccessKeyId=LTAI8Z8YW8VGkEvg&Signature=0HpM5f6QPkoake7CiRbwISflf4I%3D";
sudo chmod +x ./xlmSign;

