#启动
##1. dataServer启动
>./bifrostsync

#环境
>./environment 文件指定所在环境（dev,test,pro）,项目启动时会找相应配置文件启动
>主要是为了防止误操作

#生产环境部署
192.170.1.89/servece/bifrostsync

CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build;
ossutil cp -f ./bifrostsync  oss://dfdfedfda;
ossutil sign oss://dfdfedfda/bifrostsync    --timeout 1000;


pkill -9 bifrostsync;
rm -rf bifrostsync;
sudo wget -O bifrostsync  "http://dfdfedfda.oss-cn-hongkong.aliyuncs.com/bifrostsync?Expires=1627546055&OSSAccessKeyId=LTAI8Z8YW8VGkEvg&Signature=63JFr3KsIiKXlapSdnCMi1yz%2B0o%3D"
sudo chmod +x ./bifrostsync;
pkill -9 bifrostsync;rm -rf tmplogs/;nohup ./bifrostsync &

rm -rf nohup.out;nohup ./bifrostsync &

