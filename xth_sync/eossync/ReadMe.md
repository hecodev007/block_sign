#启动
##1. dataServer启动
>./eossync

#环境
>./environment 文件指定所在环境（dev,test,pro）,项目启动时会找相应配置文件启动
>主要是为了防止误操作



CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build;ossutil cp -f ./eossync  oss://dfdfedfda;ossutil sign oss://dfdfedfda/eossync    --timeout 1000;

pkill -9 eossync;
rm -rf eossync;
sudo wget -O eossync  "http://dfdfedfda.oss-cn-hongkong.aliyuncs.com/eossync?Expires=1642672660&OSSAccessKeyId=LTAI8Z8YW8VGkEvg&Signature=2d8XiSUt6B%2BmCtPLrLAV%2FM8XmXU%3D";
sudo chmod +x ./eossync;
