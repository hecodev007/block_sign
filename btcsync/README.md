#启动
##1. dataServer启动
>./btcsync

#环境
>./environment 文件指定所在环境（dev,test,pro）,项目启动时会找相应配置文件启动
>主要是为了防止误操作


CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build; ossutil cp -f ./btcsync  oss://dfdfedfda; ossutil sign oss://dfdfedfda/btcsync    --timeout 1000;

pkill -9 btcsync;
rm -rf ./btcsync;
sudo wget -O btcsync  "http://dfdfedfda.oss-cn-hongkong.aliyuncs.com/btcsync?Expires=1651018865&OSSAccessKeyId=LTAI8Z8YW8VGkEvg&Signature=ytHsaQErDxaRjbASp3qQbb%2Bxar4%3D";
sudo chmod +x ./btcsync;
./btcsync
