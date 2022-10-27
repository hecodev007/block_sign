#启动
##1. dataServer启动
>./adasync

#环境
>./environment 文件指定所在环境（dev,test,pro）,项目启动时会找相应配置文件启动
>主要是为了防止误操作


CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build; ossutil cp -f ./adasync  oss://dfdfedfda; ossutil sign oss://dfdfedfda/adasync    --timeout 1000;

pkill -9 adasync;
rm -rf ./adasync;
sudo wget -O adasync  "http://dfdfedfda.oss-cn-hongkong.aliyuncs.com/adasync?Expires=1639035243&OSSAccessKeyId=LTAI8Z8YW8VGkEvg&Signature=%2BUIYnZC4wzLApWBaXY6lSdwRBpw%3D";
sudo chmod +x ./adasync;
nohup ./adasync &
