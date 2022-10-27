#启动
##1. dataServer启动
>./sgbsync

#环境
>./environment 文件指定所在环境（dev,test,pro）,项目启动时会找相应配置文件启动
>主要是为了防止误操作

#生产环境部署
192.170.1.89/servece/dotsync

CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build;
ossutil cp -f ./sgbsync  oss://dfdfedfda;
ossutil sign oss://dfdfedfda/sgbsync    --timeout 1000;


pkill -9 sgbsync;
rm -rf sgbsync;
sudo wget -O sgbsync  "http://dfdfedfda.oss-cn-hongkong.aliyuncs.com/sgbsync?Expires=1638930677&OSSAccessKeyId=LTAI8Z8YW8VGkEvg&Signature=pFi%2FHthIBs%2FtCrrH9Zw3LORI8u4%3D"
sudo chmod +x ./sgbsync;
pkill -9 sgbsign;./sgbsync

rm -rf nohup.out;nohup ./sgbsync &

