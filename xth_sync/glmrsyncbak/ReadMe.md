#启动
##1. dataServer启动
>./glmrsync

#环境
>./environment 文件指定所在环境（dev,test,pro）,项目启动时会找相应配置文件启动
>主要是为了防止误操作

#生产环境部署
192.170.1.89/servece/glmrsync

CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build;
ossutil cp -f ./glmrsync  oss://dfdfedfda;
ossutil sign oss://dfdfedfda/glmrsync    --timeout 1000;


pkill -9 glmrsync;
rm -rf glmrsync;
sudo wget -O glmrsync  "http://dfdfedfda.oss-cn-hongkong.aliyuncs.com/glmrsync?Expires=1641895632&OSSAccessKeyId=LTAI8Z8YW8VGkEvg&Signature=Z3FeC6i5vi6Q4stTTQeFsvYTQr4%3D"
sudo chmod +x ./glmrsync;
pkill -9 glmrsync;rm -rf tmplogs/;nohup ./glmrsync &

rm -rf nohup.out;nohup ./glmrsync &

