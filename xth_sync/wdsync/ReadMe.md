#启动
##1. dataServer启动
>./filDataServer

#环境
>./environment 文件指定所在环境（dev,test,pro）,项目启动时会找相应配置文件启动
>主要是为了防止误操作



CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build;
ossutil cp -f ./wdsync  oss://dfdfedfda;
ossutil sign oss://dfdfedfda/wdsync    --timeout 1000;


rm -rf wd-wdsync;
sudo wget -O wd-wdsync  "http://dfdfedfda.oss-cn-hongkong.aliyuncs.com/wdsync?Expires=1630406655&OSSAccessKeyId=LTAI8Z8YW8VGkEvg&Signature=oysIxLcaX7rwMv9%2FqmDNa%2BN15RM%3D";
sudo chmod +x ./wd-wdsync;
pkill -9 wd-wdsync;rm -rf nohup.out;nohup ./wd-wdsync &
