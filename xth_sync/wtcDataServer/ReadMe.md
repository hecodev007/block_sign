#启动
##1. dataServer启动
>./wtcDataServer

#环境
>./environment 文件指定所在环境（dev,test,pro）,项目启动时会找相应配置文件启动
>主要是为了防止误操作


CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build;
ossutil cp -f ./wtcDataServer  oss://dfdfedfda;
ossutil sign oss://dfdfedfda/wtcDataServer    --timeout 1000;


sudo wget -O wtcDataServer  "http://dfdfedfda.oss-cn-hongkong.aliyuncs.com/wtcDataServer?Expires=1625812640&OSSAccessKeyId=LTAI8Z8YW8VGkEvg&Signature=gu45Ift0NJLBSktlJyk5q2g9j7c%3D"

