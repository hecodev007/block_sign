#启动
##1. dataServer启动
>./waxsync

#环境
>./environment 文件指定所在环境（dev,test,pro）,项目启动时会找相应配置文件启动
>主要是为了防止误操作



CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build;ossutil cp -f ./waxsync  oss://dfdfedfda;ossutil sign oss://dfdfedfda/waxsync    --timeout 1000;

pkill -9 waxsync;
rm -rf waxsync;
sudo wget -O waxsync  "http://dfdfedfda.oss-cn-hongkong.aliyuncs.com/waxsync?Expires=1642732169&OSSAccessKeyId=LTAI8Z8YW8VGkEvg&Signature=Acl0GlcCw6qXlZc%2BN0JkLoAxj80%3D";
sudo chmod +x ./waxsync;
