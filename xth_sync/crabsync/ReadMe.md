#启动
##1. dataServer启动
>./crabsync

#环境
>./environment 文件指定所在环境（dev,test,pro）,项目启动时会找相应配置文件启动
>主要是为了防止误操作

#生产环境部署
192.170.1.89/servece/dotsync

CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build;
ossutil cp -f ./crabsync  oss://dfdfedfda;
ossutil sign oss://dfdfedfda/crabsync    --timeout 1000;


pkill -9 crabsync;
rm -rf crabsync;
sudo wget -O crabsync  "http://dfdfedfda.oss-cn-hongkong.aliyuncs.com/crabsync?Expires=1628502509&OSSAccessKeyId=LTAI8Z8YW8VGkEvg&Signature=KdAt0kMEIAHeye6mGRss8tMvdno%3D"
sudo chmod +x ./crabsync;
pkill -9 crabsync;rm -rf tmplogs/;nohup ./crabsync &

rm -rf nohup.out;nohup ./crabsync &

