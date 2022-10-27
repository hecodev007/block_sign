1.common 下放项目公共的组件

2.utils 放model用到的组件,子文件夹可以用对应的model命名

3.app.toml配置文件放根目录，部署只需要拷贝biwSign,app.toml，2个文件

4.启动 ./biwSign;

5,创建账户
```
curl -X POST --url  http://127.0.0.1:18078/v1/biw/createaddr -d '{"num":1,"coinName":"dash","orderId":"main","mchId":"goapi"}'
```
6,签名
```
curl -X POST --url  http://127.0.0.1:18078/v1/biw/sign -d '{"coinName":"dash","orderId":"test","mchId":"goapi","txIns":[{"fromAddr":"XawwU58trgyp8wMQtPRkcQUCxJQpwmCv31","fromTxid":"3f70bbbef46bbd6ad70e9c7308d9e57e05c5b1804bcd1fe369570fca0388d176","FromIndex":0,"fromAmount":2800000}],"txOuts":[{"toAddr":"XawwU58trgyp8wMQtPRkcQUCxJQpwmCv31","toAmount":2700000},{"toAddr":"7pG9k7MmH54GTx5aFBCxim9sCEAsQwU3Bt","toAmount":89999}]}'
```
7,签名并发送交易
```
curl -X POST -basic -u rylink:rylink@telos@2020 --url  http://127.0.0.1:18078/v1/biw/transfer -d '{"coinName":"dash","orderId":"test","mchId":"goapi","txIns":[{"fromAddr":"XawwU58trgyp8wMQtPRkcQUCxJQpwmCv31","fromTxid":"3f70bbbef46bbd6ad70e9c7308d9e57e05c5b1804bcd1fe369570fca0388d176","FromIndex":0,"fromAmount":2800000}],"txOuts":[{"toAddr":"XawwU58trgyp8wMQtPRkcQUCxJQpwmCv31","toAmount":2700000},{"toAddr":"7pG9k7MmH54GTx5aFBCxim9sCEAsQwU3Bt","toAmount":89999}]}'
```
