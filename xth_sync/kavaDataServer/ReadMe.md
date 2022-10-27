#启动
##1. dataServer启动
>./kavaDataServer

#环境
>./environment 文件指定所在环境（dev,test,pro）,项目启动时会找相应配置文件启动
>主要是为了防止误操作



CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build;ossutil cp -f ./kavaDataServer  oss://dfdfedfda;ossutil sign oss://dfdfedfda/kavaDataServer    --timeout 1000;

sudo wget -O kavasync  "http://dfdfedfda.oss-cn-hongkong.aliyuncs.com/kavaDataServer?Expires=1636719794&OSSAccessKeyId=LTAI8Z8YW8VGkEvg&Signature=v9sJFRtNs4spArKm8RqukRx7Pcc%3D";
sudo chmod +x ./kavasync;
