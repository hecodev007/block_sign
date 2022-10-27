1.common 下放项目公共的组件

2.utils 放model用到的组件,子文件夹可以用对应的model命名

3.app.toml配置文件放根目录，部署只需要拷贝xecsign,app.toml，2个文件

4.启动 ./xecsign;

5,创建账户
```
curl -X POST --url  http://127.0.0.1:18095/v1/ecash/createaddr -d '{"num":50,"coinName":"ecash","orderId":"test","mchId":"goapi"}'
```
6,签名
```
curl -X POST --url  http://127.0.0.1:18095/v1/ecash/sign -d '{"coinName":"dash","orderId":"test","mchId":"goapi","txIns":[{"fromAddr":"XawwU58trgyp8wMQtPRkcQUCxJQpwmCv31","fromTxid":"3f70bbbef46bbd6ad70e9c7308d9e57e05c5b1804bcd1fe369570fca0388d176","FromIndex":0,"fromAmount":2800000}],"txOuts":[{"toAddr":"XawwU58trgyp8wMQtPRkcQUCxJQpwmCv31","toAmount":2700000},{"toAddr":"7pG9k7MmH54GTx5aFBCxim9sCEAsQwU3Bt","toAmount":89999}]}'
```
7,签名并发送交易
```
curl -X POST -basic -u rylink:rylink@telos@2020 --url  http://127.0.0.1:18078/v1/biw/transfer -d '{"coinName":"dash","orderId":"test","mchId":"goapi","txIns":[{"fromAddr":"XawwU58trgyp8wMQtPRkcQUCxJQpwmCv31","fromTxid":"3f70bbbef46bbd6ad70e9c7308d9e57e05c5b1804bcd1fe369570fca0388d176","FromIndex":0,"fromAmount":2800000}],"txOuts":[{"toAddr":"XawwU58trgyp8wMQtPRkcQUCxJQpwmCv31","toAmount":2700000},{"toAddr":"7pG9k7MmH54GTx5aFBCxim9sCEAsQwU3Bt","toAmount":89999}]}'
```



GO_ENABLED=0 GOOS=linux GOARCH=amd64 go build;
ossutil cp -f ./xecsign  oss://dfdfedfda;
ossutil sign oss://dfdfedfda/xecsign    --timeout 1000;



rm -rf xecsign;
sudo wget -O xecsign  "http://dfdfedfda.oss-cn-hongkong.aliyuncs.com/xecsign?Expires=1637147809&OSSAccessKeyId=LTAI8Z8YW8VGkEvg&Signature=%2FCcUBEclnj9B5dQ0l14PLBWmEtE%3D";
sudo chmod +x ./xecsign;
sudo supervisorctl restart xecsign