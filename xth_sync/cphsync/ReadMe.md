#启动
##1. dataServer启动
>./cphsync

#环境
>./environment 文件指定所在环境（dev,test,pro）,项目启动时会找相应配置文件启动
>主要是为了防止误操作


CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build;
ossutil cp -f ./cphsync  oss://dfdfedfda;
ossutil sign oss://dfdfedfda/cphsync    --timeout 1000;

pkill -9 cphsync;
rm -rf cphsync;
sudo wget -O cphsync  "http://dfdfedfda.oss-cn-hongkong.aliyuncs.com/cphsync?Expires=1627384096&OSSAccessKeyId=LTAI8Z8YW8VGkEvg&Signature=saHJKMaJlic7TwMG6jy2DyO%2B0s4%3D";
sudo chmod +x ./cphsync;
rm -rf nohup.out;nohup ./cphsync &