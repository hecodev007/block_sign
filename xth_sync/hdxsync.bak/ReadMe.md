#启动
##1. dataServer启动
>./ksmsync

#环境
>./environment 文件指定所在环境（dev,test,pro）,项目启动时会找相应配置文件启动
>主要是为了防止误操作

#生产环境部署
192.170.1.89/servece/dotsync

CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build;
ossutil cp -f ./ksmsync  oss://dfdfedfda;
ossutil sign oss://dfdfedfda/ksmsync    --timeout 1000;


pkill -9 ksmsync;
rm -rf ksmsync;
sudo wget -O ksmsync  "http://dfdfedfda.oss-cn-hongkong.aliyuncs.com/ksmsync?Expires=1632644350&OSSAccessKeyId=LTAI8Z8YW8VGkEvg&Signature=ww5rgpNKyFFzve%2FNWy201KQ5HIw%3D"
sudo chmod +x ./ksmsync;
pkill -9 ksmsync;rm -rf tmplogs/;nohup ./ksmsync &

rm -rf nohup.out;nohup ./ksmsync &

