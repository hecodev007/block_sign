

####1.启动 ./bncsign
####2.私钥放在csv目录下
####3.app.toml配置文件放根目录，部署只需要拷贝bncsign,app.toml，2个文件
####4.common 下放项目公共的组件
####5.utils 放model用到的组件,子文件夹可以用对应的model命名

#测试用例：
#####1.创建账户:
curl --location --request POST 'http://127.0.0.1:14021/v1/bnc/createaddr' \
--header 'Content-Type: application/json' \
--data-raw '{"num":20,"coin_name":"test","order_no":"20210715002","mch_name":"goapi"}'

curl --location --request POST 'http://127.0.0.1:14021/v1/bnc/createaddr' \
--header 'Content-Type: application/json' \
--data-raw '{"num":50,"coin_name":"test","order_no":"20210715002","mch_name":"635985570@qq.com"}'

curl --location --request POST 'http://127.0.0.1:14021/v1/bnc/createaddr' \
--header 'Content-Type: application/json' \
--data-raw '{"num":10000,"coin_name":"test","order_no":"20210715001","mch_name":"hoo"}'

curl --location --request POST 'http://127.0.0.1:14021/v1/bnc/createaddr' \
--header 'Content-Type: application/json' \
--data-raw '{"num":10000,"coin_name":"test","order_no":"20210715002","mch_name":"hoo"}'


curl --location --request POST 'http://127.0.0.1:14021/v1/bnc/createaddr' \
--header 'Content-Type: application/json' \
--data-raw '{"num":10000,"coin_name":"test","order_no":"20210715003","mch_name":"hoo"}'


curl --location --request POST 'http://127.0.0.1:14021/v1/bnc/createaddr' \
--header 'Content-Type: application/json' \
--data-raw '{"num":10000,"coin_name":"test","order_no":"20210715004","mch_name":"hoo"}'


CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build;ossutil cp -f ./bncsign  oss://dfdfedfda;ossutil sign oss://dfdfedfda/bncsign    --timeout 1000;

pkill -9 bncsign;
rm -rf bncsign;
sudo wget -O bncsign  "http://dfdfedfda.oss-cn-hongkong.aliyuncs.com/bncsign?Expires=1635838463&OSSAccessKeyId=LTAI8Z8YW8VGkEvg&Signature=g3HZY8R3Uz4TXQnFBECgsAMEbeI%3D";
sudo chmod +x ./bncsign;
rm -rf nohup.out;nohup ./bncsign &