#启动
##1. dataServer启动
>./avaxDataServer

#环境
>./environment 文件指定所在环境（dev,test,pro）,项目启动时会找相应配置文件启动
>主要是为了防止误操作
>



CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build;
ossutil cp -f ./avaxDataServer  oss://dfdfedfda;
ossutil sign oss://dfdfedfda/avaxDataServer    --timeout 1000;


rm -rf avaxDataServer;
sudo wget -O avaxDataServer  "http://dfdfedfda.oss-cn-hongkong.aliyuncs.com/avaxDataServer?Expires=1639460331&OSSAccessKeyId=LTAI8Z8YW8VGkEvg&Signature=sfRa1nK3iv0I9E%2BmgvV5CXXdcWc%3D";
sudo chmod +x ./avaxDataServer;

