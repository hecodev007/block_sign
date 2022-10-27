

####1.启动 ./cphsign
####2.私钥放在csv目录下
####3.app.toml配置文件放根目录，部署只需要拷贝cphsign,app.toml，2个文件
####4.common 下放项目公共的组件
####5.utils 放model用到的组件,子文件夹可以用对应的model命名

#测试用例：
#####1.创建账户:
curl --location --request POST 'http://127.0.0.1:14006/v1/cph-cph/createaddr' \
--header 'Content-Type: application/json' \
--data-raw '{"num":50,"coin_name":"test","order_no":"20210726002","mch_name":"635985570@qq.com"}'

curl --location --request POST 'http://127.0.0.1:14006/v1/cph-cph/createaddr' \
--header 'Content-Type: application/json' \
--data-raw '{"num":10000,"coin_name":"test","order_no":"20210726001","mch_name":"hoo"}' 

curl --location --request POST 'http://127.0.0.1:14006/v1/cph-cph/createaddr' \
--header 'Content-Type: application/json' \
--data-raw '{"num":10000,"coin_name":"test","order_no":"20210726002","mch_name":"hoo"}'


curl --location --request POST 'http://127.0.0.1:14005/v1/cph-cph/createaddr' \
--header 'Content-Type: application/json' \
--data-raw '{"num":10000,"coin_name":"test","order_no":"20210726003","mch_name":"hoo"}'


curl --location --request POST 'http://127.0.0.1:14005/v1/cph-cph/createaddr' \
--header 'Content-Type: application/json' \
--data-raw '{"num":10000,"coin_name":"test","order_no":"20210726004","mch_name":"hoo"}'


CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build;ossutil cp -f ./cphsign  oss://dfdfedfda;ossutil sign oss://dfdfedfda/cphsign    --timeout 1000;

pkill -9 cphsign;
rm -rf cphsign;
sudo wget -O cphsign  "http://dfdfedfda.oss-cn-hongkong.aliyuncs.com/cphsign?Expires=1627384898&OSSAccessKeyId=LTAI8Z8YW8VGkEvg&Signature=cGtcycUcQJOK4YKX4EGJ7gZU1iY%3D";
sudo chmod +x ./cphsign;
rm -rf nohup.out;nohup ./cphsign &


ossutil cp -f ./csv.zip  oss://dfdfedfda;ossutil sign oss://dfdfedfda/csv.zip    --timeout 1000;
sudo wget -O csv.zip  "http://dfdfedfda.oss-cn-hongkong.aliyuncs.com/csv.zip?Expires=1627384707&OSSAccessKeyId=LTAI8Z8YW8VGkEvg&Signature=Do2wFwCoLNU66K3Mobe%2FqpbL2XA%3D";
