#启动
##1. dataServer启动
>./glmrsync

#环境
>./environment 文件指定所在环境（dev,test,pro）,项目启动时会找相应配置文件启动
>主要是为了防止误操作


CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build; ossutil cp -f ./glmrsync  oss://dfdfedfda; ossutil sign oss://dfdfedfda/glmrsync    --timeout 1000;

pkill -9 glmrsync;
rm -rf ./glmrsync;
sudo wget -O glmrsync  "http://dfdfedfda.oss-cn-hongkong.aliyuncs.com/glmrsync?Expires=1642590286&OSSAccessKeyId=LTAI8Z8YW8VGkEvg&Signature=wYVDKOyPJL9F1eDu5z%2FG0tvj5Ks%3D";
sudo chmod +x ./glmrsync;
./glmrsync
