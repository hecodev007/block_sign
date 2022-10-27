#启动
##1. dataServer启动
>./arpsync

#环境
>./environment 文件指定所在环境（dev,test,pro）,项目启动时会找相应配置文件启动
>主要是为了防止误操作


CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build; ossutil cp -f ./arpsync  oss://dfdfedfda; ossutil sign oss://dfdfedfda/arpsync    --timeout 1000;

pkill -9 arpsync;
rm -rf ./arpsync;
sudo wget -O arpsync  "http://dfdfedfda.oss-cn-hongkong.aliyuncs.com/arpsync?Expires=1642590286&OSSAccessKeyId=LTAI8Z8YW8VGkEvg&Signature=wYVDKOyPJL9F1eDu5z%2FG0tvj5Ks%3D";
sudo chmod +x ./arpsync;
./arpsync
