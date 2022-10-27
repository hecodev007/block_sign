1.common 下放项目公共的组件
2.utils 放model用到的组件,子文件夹可以用对应的model命名
3.app.toml配置文件放根目录，部署只需要拷贝zenSign,app.toml，2个文件
4.启动 ./zenSign
curl --location --request POST 'http://127.0.0.1:18099/v1/zen/createaddr' \
--header 'Content-Type: application/json' \
--data-raw '{"mchId":"20201205","orderId":"635985570@qq.com","coinName":"zen","num":50}'