##crustio
1.common 下放项目公共的组件
2.utils 放model用到的组件,子文件夹可以用对应的model命名
3.app.toml配置文件放根目录，部署只需要拷贝cruSign,app.toml，2个文件
4.启动 ./cruSign
curl -X POST --url  http://127.0.0.1:18070/v1/ghost/createaddr -d '{ "num":50,"coin_name":"ghost","order_no":"12306","mch_name": "goapi"}'
