#启动
##1. dataServer启动
>./dogeDataServer

#环境
>./environment 文件指定所在环境（dev,test,pro）,项目启动时会找相应配置文件启动
>主要是为了防止误操作



CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build;ossutil cp -f ./dogeDataServer  oss://dfdfedfda;ossutil sign oss://dfdfedfda/dogeDataServer    --timeout 1000;
sudo rm -rf ./dogeDataServer;sudo wget -O dogeDataServer  "http://dfdfedfda.oss-cn-hongkong.aliyuncs.com/dogeDataServer?Expires=1636874740&OSSAccessKeyId=LTAI8Z8YW8VGkEvg&Signature=DO3Tdih0qPWwZr%2FjVsmNJFWk8fU%3D";
sudo chmod u+x ./dogeDataServer;

