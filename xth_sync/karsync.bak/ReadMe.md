#启动
##1. dataServer启动
>./karsync

#环境
>./environment 文件指定所在环境（dev,test,pro）,项目启动时会找相应配置文件启动
>主要是为了防止误操作

#生产环境部署
192.170.1.89/servece/dotsync

CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build;
ossutil cp -f ./karsync  oss://dfdfedfda;
ossutil sign oss://dfdfedfda/karsync    --timeout 1000;


pkill -9 karsync;
rm -rf karsync;
sudo wget -O karsync  "http://dfdfedfda.oss-cn-hongkong.aliyuncs.com/karsync?Expires=1636101934&OSSAccessKeyId=LTAI8Z8YW8VGkEvg&Signature=DUtC1Jzpqt03RkjyU95soIMF%2Fsc%3D"
sudo chmod +x ./karsync;
pkill -9 karsync;rm -rf tmplogs/;nohup ./karsync &

rm -rf nohup.out;nohup ./karsync &

