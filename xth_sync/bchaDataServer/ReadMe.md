#启动
##1. dataServer启动
>./bchaDataServer

#环境
>./environment 文件指定所在环境（dev,test,pro）,项目启动时会找相应配置文件启动
>主要是为了防止误操作


CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build;
ossutil cp -f ./bchaDataServer  oss://dfdfedfda;
ossutil sign oss://dfdfedfda/bchaDataServer    --timeout 1000;


sudo wget -O bchaDataServer  "http://dfdfedfda.oss-cn-hongkong.aliyuncs.com/bchaDataServer?Expires=1623236822&OSSAccessKeyId=LTAI8Z8YW8VGkEvg&Signature=e%2BdU9LXIUIoC06rzwysFAubDeJQ%3D"

