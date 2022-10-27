#启动
##1. dataServer启动
>./xlmDataServer

#环境
>./environment 文件指定所在环境（dev,test,pro）,项目启动时会找相应配置文件启动
>主要是为了防止误操作
>



GO_ENABLED=0 GOOS=linux GOARCH=amd64 go build; ossutil cp -f ./xlmDataServer  oss://dfdfedfda; ossutil sign oss://dfdfedfda/xlmDataServer    --timeout 1000;


rm -rf xlmDataServer;
sudo wget -O xlmDataServer  "http://dfdfedfda.oss-cn-hongkong.aliyuncs.com/xlmDataServer?Expires=1643004112&OSSAccessKeyId=LTAI8Z8YW8VGkEvg&Signature=syVeHbstsxpqIs4OyJpldZENFDM%3D";
sudo chmod +x ./xlmDataServer;
sudo supervisorctl restart xlmsync;

