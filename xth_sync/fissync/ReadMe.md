#启动
##1. dataServer启动
>./fissync

#环境
>./environment 文件指定所在环境（dev,test,pro）,项目启动时会找相应配置文件启动
>主要是为了防止误操作

#生产环境部署
192.170.1.89/servece/dotsync

CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build;
ossutil cp -f ./fissync  oss://dfdfedfda;
ossutil sign oss://dfdfedfda/fissync    --timeout 1000;


pkill -9 fissync;
rm -rf fissync;

sudo wget -O fissync  "http://dfdfedfda.oss-cn-hongkong.aliyuncs.com/fissync?Expires=1636773249&OSSAccessKeyId=LTAI8Z8YW8VGkEvg&Signature=hlryd8%2BDm%2BnooSVkjsT2g7Vd54g%3D"
sudo chmod +x ./fissync;

pkill -9 fissync;rm -rf tmplogs/;nohup ./fissync &

rm -rf nohup.out;nohup ./fissync &

