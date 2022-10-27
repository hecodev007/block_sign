#启动
##1. dataServer启动
>./moacDataServer

#环境
>./environment 文件指定所在环境（dev,test,pro）,项目启动时会找相应配置文件启动
>主要是为了防止误操作


CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build;
ossutil cp -f ./moacDataServer  oss://dfdfedfda;
ossutil sign oss://dfdfedfda/moacDataServer    --timeout 1000;

rm -rf ./moacDataServer;
sudo wget -O moacDataServer  "http://dfdfedfda.oss-cn-hongkong.aliyuncs.com/moacDataServer?Expires=1624937244&OSSAccessKeyId=LTAI8Z8YW8VGkEvg&Signature=X7kQpUaqaQyKE%2BSHvGGXIzhYNa8%3D";
sudo chmod 777 ./moacDataServer;
