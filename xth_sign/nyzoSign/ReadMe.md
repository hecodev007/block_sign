#### 1.启动 ./cfxSign

#### 2.私钥放在csv目录下

#### 3.app.toml配置文件放根目录，部署只需要拷贝cfxSign,app.toml，2个文件

#### 4.common 下放项目公共的组件

#### 5.utils 放model用到的组件,子文件夹可以用对应的model命名

# 测试用例：

##### 1.创建账户:

>curl --location --request POST 'http://127.0.0.1:22010/v1/nyzo/createaddr' \
--header 'Content-Type: application/json' \
--data-raw '{ "num": 1, "coin_name": "nyzo", "order_no": "20201211", "mch_name": "635985570@qq.com" }'


