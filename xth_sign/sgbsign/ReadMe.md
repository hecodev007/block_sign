

####1.启动 ./sgbsign
####2.私钥放在csv目录下
####3.app.toml配置文件放根目录，部署只需要拷贝sgbsign,app.toml，2个文件
####4.common 下放项目公共的组件
####5.utils 放model用到的组件,子文件夹可以用对应的model命名

#测试用例：
#####1.创建账户:
curl --location --request POST 'http://127.0.0.1:14005/v1/sgb-sgb/createaddr' \
--header 'Content-Type: application/json' \
--data-raw '{"num":50,"coin_name":"test","order_no":"20210715002","mch_name":"635985570@qq.com"}'

curl --location --request POST 'http://127.0.0.1:14005/v1/sgb-sgb/createaddr' \
--header 'Content-Type: application/json' \
--data-raw '{"num":10000,"coin_name":"test","order_no":"20210715001","mch_name":"hoo"}' 

curl --location --request POST 'http://127.0.0.1:14005/v1/sgb-sgb/createaddr' \
--header 'Content-Type: application/json' \
--data-raw '{"num":10000,"coin_name":"test","order_no":"20210715002","mch_name":"hoo"}'


curl --location --request POST 'http://127.0.0.1:14005/v1/sgb-sgb/createaddr' \
--header 'Content-Type: application/json' \
--data-raw '{"num":10000,"coin_name":"test","order_no":"20210715003","mch_name":"hoo"}'


curl --location --request POST 'http://127.0.0.1:14005/v1/sgb-sgb/createaddr' \
--header 'Content-Type: application/json' \
--data-raw '{"num":10000,"coin_name":"test","order_no":"20210715004","mch_name":"hoo"}'


go build;ossutil cp -f ./sgbsign  oss://dfdfedfda;ossutil sign oss://dfdfedfda/sgbsign    --timeout 1000;

pkill -9 sgbsign;
rm -rf sgbsign;
sudo wget -O sgbsign  "http://dfdfedfda.oss-cn-hongkong.aliyuncs.com/sgbsign?Expires=1631692658&OSSAccessKeyId=LTAI8Z8YW8VGkEvg&Signature=aSFhlmPBQ4n02exPztPBChHoL6o%3D";
sudo chmod +x ./sgbsign;
rm -rf nohup.out;nohup ./sgbsign &