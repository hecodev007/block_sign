#启动
##1. dataServer启动
>./oktsync

#环境
>./environment 文件指定所在环境（dev,test,pro）,项目启动时会找相应配置文件启动
>主要是为了防止误操作


CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build; ossutil cp -f ./oktsync  oss://dfdfedfda; ossutil sign oss://dfdfedfda/oktsync    --timeout 1000;

pkill -9 oktsync;
rm -rf ./oktsync;
sudo wget -O oktsync  "http://dfdfedfda.oss-cn-hongkong.aliyuncs.com/oktsync?Expires=1628043632&OSSAccessKeyId=LTAI8Z8YW8VGkEvg&Signature=dlo7USUHKlSu65DAQzO0viiydwI%3D";
sudo chmod 777 ./oktsync;
./oktsync
