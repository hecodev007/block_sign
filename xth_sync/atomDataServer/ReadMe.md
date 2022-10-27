#启动
##1. dataServer启动
>./atomDataServer

#环境
>./environment 文件指定所在环境（dev,test,pro）,项目启动时会找相应配置文件启动
>主要是为了防止误操作



CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build;
ossutil cp -f ./atomDataServer  oss://dfdfedfda;
ossutil sign oss://dfdfedfda/atomDataServer    --timeout 1000;


rm -rf atomDataServer;
sudo wget -O atomDataServer  "http://dfdfedfda.oss-cn-hongkong.aliyuncs.com/atomDataServer?Expires=1643097800&OSSAccessKeyId=LTAI8Z8YW8VGkEvg&Signature=ZKcod82SmtRPgGcOqLerU08la7M%3D";
sudo chmod +x ./atomDataServer;

