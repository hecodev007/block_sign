#启动
##1. dataServer启动
>./filDataServer

#环境
>./environment 文件指定所在环境（dev,test,pro）,项目启动时会找相应配置文件启动
>主要是为了防止误操作




CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build;ossutil cp -f ./filDataServer  oss://dfdfedfda
ossutil sign oss://dfdfedfda/filDataServer    --timeout 1000;

rm -rf filDataServer;
wget -O filDataServer  "http://dfdfedfda.oss-cn-hongkong.aliyuncs.com/filDataServer?Expires=1646897960&OSSAccessKeyId=LTAI8Z8YW8VGkEvg&Signature=BEN2gSchZTqPsIh0hqg2zbdexVM%3D";
chmod u+x filDataServer;

