1.common 下放项目公共的组件
2.utils 放model用到的组件,子文件夹可以用对应的model命名
3.app.toml配置文件放根目录，部署只需要拷贝demo,app.toml，2个文件
4.启动 ./demo
curl -X POST --url  http://127.0.0.1:18074/v1/dash/createaddr -d '{"num":1,"coinName":"dash","orderId":"main","mchId":"goapi"}'

curl -X POST --url  http://127.0.0.1:18074/v1/dash/sign -d '{"coinName":"dash","orderId":"test","mchId":"goapi","fromAddr":"X-avax10dafagra6vg2x3d3899nvetkuh8hfq5umexngn","toAddr":"x-avax1eg8cwusc09vpvrruy3gvllj7lehs8hurkyfxpm","changeAddr":"x-avax10dafagra6vg2x3d3899nvetkuh8hfq5umexngn","amount":100000,"fee":1000000,"utxos":["114joKFFaUKF5ybUf3oAoka3AM2fdt5on66gR5TugeM64CA16SmaBSSoNPfp4LFcRbDCTexg6PkC78Eow3JiErgGneq5QHWPnM9tn73DAdCMmBS8myiNotKautFg17pXKPnrsUZ39uhuJpJFCs7gLkqPG5nwtanaSZo58s"]}'

curl -X POST --url  http://127.0.0.1:18074/v1/dash/sign -d '{"coinName":"dash","orderId":"test","mchId":"goapi","txIns":[{"fromAddr":"XawwU58trgyp8wMQtPRkcQUCxJQpwmCv31","fromTxid":"3f70bbbef46bbd6ad70e9c7308d9e57e05c5b1804bcd1fe369570fca0388d176","FromIndex":0,"fromAmount":2800000}],"txOuts":[{"toAddr":"XawwU58trgyp8wMQtPRkcQUCxJQpwmCv31","toAmount":2700000},{"toAddr":"7pG9k7MmH54GTx5aFBCxim9sCEAsQwU3Bt","toAmount":89999}]}'


curl -X POST -basic -u rylink:rylink@telos@2020 --url  http://127.0.0.1:18074/v1/dash/transfer -d '{"coinName":"dash","orderId":"test","mchId":"goapi","txIns":[{"fromAddr":"XawwU58trgyp8wMQtPRkcQUCxJQpwmCv31","fromTxid":"3f70bbbef46bbd6ad70e9c7308d9e57e05c5b1804bcd1fe369570fca0388d176","FromIndex":0,"fromAmount":2800000}],"txOuts":[{"toAddr":"XawwU58trgyp8wMQtPRkcQUCxJQpwmCv31","toAmount":2700000},{"toAddr":"7pG9k7MmH54GTx5aFBCxim9sCEAsQwU3Bt","toAmount":89999}]}'

