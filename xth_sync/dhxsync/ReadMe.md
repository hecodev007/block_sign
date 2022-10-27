#启动
##1. dataServer启动
>./dhxsync

#环境
>./environment 文件指定所在环境（dev,test,pro）,项目启动时会找相应配置文件启动
>主要是为了防止误操作

#生产环境部署
192.170.1.89/servece/dotsync

CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build;
ossutil cp -f ./dhxsync  oss://dfdfedfda;
ossutil sign oss://dfdfedfda/dhxsync    --timeout 1000;


pkill -9 dhxsync;
rm -rf dhxsync;
sudo wget -O dhx-sync  "http://dfdfedfda.oss-cn-hongkong.aliyuncs.com/dhxsync?Expires=1631004622&OSSAccessKeyId=LTAI8Z8YW8VGkEvg&Signature=3W%2F9VWE%2BzRH0EfdpnRfxbkmv4sc%3D"
sudo chmod +x ./dhx-sync;
pkill -9 dhxsync;rm -rf tmplogs/;nohup ./dhxsync &

rm -rf nohup.out;nohup ./dhxsync &

