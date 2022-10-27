#启动
>./steemsync

#环境和配置文件
>项目启动时会按顺序找相应配置文件(./app_dev.toml,./app.toml)

#编译
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build; aws s3 cp ./steemsync s3://piupiupiu/;

#启动
pkill -9 steemsync;
rm -rf ./steemsync;
sudo wget -O steemsync "https://piupiupiu.s3.ap-northeast-1.amazonaws.com/steemsync";
sudo chmod +x ./steemsync;
nohup ./steemsync &