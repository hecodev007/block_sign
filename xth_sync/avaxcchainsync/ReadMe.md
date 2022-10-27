#启动
##1. dataServer启动
>./avaxcchainsync

#环境
>./environment 文件指定所在环境（dev,test,pro）,项目启动时会找相应配置文件启动
>主要是为了防止误操作


CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build; ossutil cp -f ./avaxcchainsync  oss://dfdfedfda; ossutil sign oss://dfdfedfda/avaxcchainsync    --timeout 1000;

pkill -9 avaxcchainsync;
rm -rf ./avaxcchainsync;
sudo wget -O avaxcchainsync  "http://dfdfedfda.oss-cn-hongkong.aliyuncs.com/avaxcchainsync?Expires=1650251845&OSSAccessKeyId=LTAI8Z8YW8VGkEvg&Signature=dh1jimiNml6p%2F3iFUWN3p20i2d8%3D";
sudo chmod +x ./avaxcchainsync;
nohup ./avaxcchainsync &
