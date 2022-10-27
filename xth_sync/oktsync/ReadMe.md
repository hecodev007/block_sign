#启动
##1. dataServer启动
>./oktsync

#环境
>./environment 文件指定所在环境（dev,test,pro）,项目启动时会找相应配置文件启动
>主要是为了防止误操作


#编译
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build; aws s3 cp ./oktsync s3://piupiupiu/;

#启动
pkill -9 oktsync;
rm -rf ./oktsync;
sudo wget -O oktsync "https://piupiupiu.s3.ap-northeast-1.amazonaws.com/oktsync";
sudo chmod +x ./oktsync;
sudo supervisorctl restart oktsync

