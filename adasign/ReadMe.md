####1.启动 ./adasign
####2.私钥放在csv目录下
####3.app.toml配置文件放根目录，部署只需要拷贝adasign,app.toml，2个文件
####4.common 下放项目公共的组件
####5.utils 放model用到的组件,子文件夹可以用对应的model命名

#测试用例：
#####1.创建账户:
curl --location --request POST 'http://127.0.0.1:14014/v1/ada/createaddr' \
--header 'Content-Type: application/json' \
--data-raw '{"num":50,"coinName":"test","orderId":"20211209002","mchId":"635985570@qq.com"}'

curl --location --request POST 'http://127.0.0.1:14014/v1/ada/createaddr' \
--header 'Content-Type: application/json' \
--data-raw '{"num":10000,"coinName":"test","orderId":"20211209001","mchId":"hoo"}'

curl --location --request POST 'http://127.0.0.1:14014/v1/ada/createaddr' \
--header 'Content-Type: application/json' \
--data-raw '{"num":10000,"coinName":"test","orderId":"20211209002","mchId":"hoo"}'


curl --location --request POST 'http://127.0.0.1:14014/v1/ada/createaddr' \
--header 'Content-Type: application/json' \
--data-raw '{"num":10000,"coinName":"test","orderId":"20211209003","mchId":"hoo"}'


curl --location --request POST 'http://127.0.0.1:14014/v1/ada/createaddr' \
--header 'Content-Type: application/json' \
--data-raw '{"num":10000,"coinName":"test","orderId":"20211209004","mchId":"hoo"}'



go build;ossutil cp -f ./adasign  oss://dfdfedfda;ossutil sign oss://dfdfedfda/adasign    --timeout 1000;

pkill -9 adasign;
rm -rf ./adasign;
wget -O adasign "http://dfdfedfda.oss-cn-hongkong.aliyuncs.com/adasign?Expires=1641472271&OSSAccessKeyId=LTAI8Z8YW8VGkEvg&Signature=ojUA0tg17wCzhbtFk4erTVQ4630%3D";
sudo chmod u+x ./adasign;
nohup ./adasign &
