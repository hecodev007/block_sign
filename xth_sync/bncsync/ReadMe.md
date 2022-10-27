#启动
##1. dataServer启动
>./bncsync

#环境
>./environment 文件指定所在环境（dev,test,pro）,项目启动时会找相应配置文件启动
>主要是为了防止误操作

#生产环境部署
192.170.1.89/servece/dotsync

CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build;
ossutil cp -f ./bncsync  oss://dfdfedfda;
ossutil sign oss://dfdfedfda/bncsync    --timeout 1000;


pkill -9 bncsync;
rm -rf bncsync;
sudo wget -O bncsync  "http://dfdfedfda.oss-cn-hongkong.aliyuncs.com/bncsync?Expires=1649208727&OSSAccessKeyId=LTAI8Z8YW8VGkEvg&Signature=n0BjD7bD4%2FYjlJZnbTpKBPgEoPI%3D"
sudo chmod +x ./bncsync;
pkill -9 bncsync;rm -rf tmplogs/;nohup ./bncsync &

rm -rf nohup.out;nohup ./bncsync &

