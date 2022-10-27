#启动
##1. dataServer启动
>./iostsync

#环境
>./environment 文件指定所在环境（dev,test,pro）,项目启动时会找相应配置文件启动
>主要是为了防止误操作


CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build; ossutil cp -f ./iostsync  oss://dfdfedfda; ossutil sign oss://dfdfedfda/iostsync    --timeout 1000;

pkill -9 iostsync;
rm -rf ./iostsync;
sudo wget -O iostsync  "http://dfdfedfda.oss-cn-hongkong.aliyuncs.com/iostsync?Expires=1631959693&OSSAccessKeyId=LTAI8Z8YW8VGkEvg&Signature=xnGbktP5Vn%2Fia1ey0rovyLFfSXU%3D";
sudo chmod +x ./iostsync;
./iostsync
