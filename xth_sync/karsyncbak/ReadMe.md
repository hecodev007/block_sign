#启动
##1. dataServer启动
>./karsync

#环境
>./environment 文件指定所在环境（dev,test,pro）,项目启动时会找相应配置文件启动
>主要是为了防止误操作

#生产环境部署
192.170.1.89/servece/dotsync

CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build;
ossutil cp -f ./karsync  oss://dfdfedfda;
ossutil sign oss://dfdfedfda/karsync    --timeout 1000;


sudo wget -O dotsync  "http://dfdfedfda.oss-cn-hongkong.aliyuncs.com/karsync?Expires=1623231327&OSSAccessKeyId=LTAI8Z8YW8VGkEvg&Signature=M459jhMVI2DUcXVgBt7VxJ68QsA%3D"


