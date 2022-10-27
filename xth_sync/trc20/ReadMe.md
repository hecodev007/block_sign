#启动
##1. dataServer启动
>./lunasync

#环境
>./environment 文件指定所在环境（dev,test,pro）,项目启动时会找相应配置文件启动
>主要是为了防止误操作



go build;ossutil cp -f ./lunasync  oss://dfdfedfda;ossutil sign oss://dfdfedfda/lunasync    --timeout 1000;

rm -rf ./lunasync;
sudo wget -O lunasync  "http://dfdfedfda.oss-cn-hongkong.aliyuncs.com/lunasync?Expires=1648276585&OSSAccessKeyId=LTAI8Z8YW8VGkEvg&Signature=gElwJvgq8KAetQzO1pILV49Anh0%3D";
sudo chmod +x ./lunasync;
