#启动
##1. dataServer启动
>./xecsync

#环境
>./environment 文件指定所在环境（dev,test,pro）,项目启动时会找相应配置文件启动
>主要是为了防止误操作


CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build;
ossutil cp -f ./xecsync  oss://dfdfedfda;
ossutil sign oss://dfdfedfda/xecsync    --timeout 1000;

rm -rf xecsync;
sudo wget -O xecsync  "http://dfdfedfda.oss-cn-hongkong.aliyuncs.com/xecsync?Expires=1637375723&OSSAccessKeyId=LTAI8Z8YW8VGkEvg&Signature=rd0tphxtQV5mmf%2Bg4qjc1c%2BEjgU%3D"
sudo  chmod u+x xecsync;

