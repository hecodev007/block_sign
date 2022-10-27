####1.启动 ./telosSign
####2.私钥放在csv目录下
####3.app.toml配置文件放根目录，部署只需要拷贝telosSign,app.toml，2个文件
####4.common 下放项目公共的组件
####5.utils 放model用到的组件,子文件夹可以用对应的model命名

#测试用例：
#####1.创建账户:
curl -X POST --url http://127.0.0.1:14031/v1/wax/createaddr -d '{"num":1,"coin_name":"wax","order_no":"20201211","mch_name":"635985570@qq.com"}'

curl -X POST --url http://127.0.0.1:14031/v1/wax/createaddr -d '{"num":1,"coin_name":"wax","order_no":"20201211","mch_name":"hoo"}'



CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build;ossutil cp -f ./waxsign  oss://dfdfedfda;ossutil sign oss://dfdfedfda/waxsign    --timeout 1000;

pkill -9 waxsign;
rm -rf waxsign;
sudo wget -O waxsign  "http://dfdfedfda.oss-cn-hongkong.aliyuncs.com/waxsign?Expires=1642409515&OSSAccessKeyId=LTAI8Z8YW8VGkEvg&Signature=7qvaAqQq6%2Bwon5oPCie%2F1YaTrAk%3D";
sudo chmod +x ./waxsign;
