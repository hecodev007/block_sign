1.common 下放项目公共的组件
2.utils 放model用到的组件,子文件夹可以用对应的model命名
3.app.toml配置文件放根目录，部署只需要拷贝avaxSign,app.toml，2个文件
4.启动 ./avaxSign
curl -X POST --url  http://127.0.0.1:18071/v1/avax/sign -d '{"coinName":"avax","orderNo":"test","mchName":"goapi","fromAddr":"X-avax10dafagra6vg2x3d3899nvetkuh8hfq5umexngn","toAddr":"x-avax1eg8cwusc09vpvrruy3gvllj7lehs8hurkyfxpm","changeAddr":"x-avax10dafagra6vg2x3d3899nvetkuh8hfq5umexngn","amount":100000,"fee":1000000,"utxos":["114joKFFaUKF5ybUf3oAoka3AM2fdt5on66gR5TugeM64CA16SmaBSSoNPfp4LFcRbDCTexg6PkC78Eow3JiErgGneq5QHWPnM9tn73DAdCMmBS8myiNotKautFg17pXKPnrsUZ39uhuJpJFCs7gLkqPG5nwtanaSZo58s"]}'

curl -X POST --url  http://127.0.0.1:18071/v1/avax/sign -d '{"coinName":"avax","orderNo":"test","mchName":"hoo","fromAddr":"X-avax10dafagra6vg2x3d3899nvetkuh8hfq5umexngn","toAddr":"X-avax1eg8cwusc09vpvrruy3gvllj7lehs8hurkyfxpm","changeAddr":"X-avax10dafagra6vg2x3d3899nvetkuh8hfq5umexngn","amount":99999990,"fee":1000000,"utxos":["11RyXqXJYZXWYG6Vf8aDn74ZxZNcKzvJ8spETfVoskC8BeGxWJnpDYG6q2j9UVxcwjHYxdJCJtPkW85qr6eLkLcfqvmRERjE9VbjH1wZoXttujtQHxvW3pYdBHSx7tpdkmguSXEzkwhs4R3cRFqMSB2Y1ehw1NdLaBAz2W","11EXjbyDLQDTFcqALp15TzbA1ThGDnJSK8h6bhhZtSY3vUZY6DHKaQEftzfaPzEvuoP4j4ixiMeRAN4bMegCLphAbTLdJ7eJoRYNG99zzP4jV2P6WFoEdxtTPmMFo1Knk6f7mi1RDPLdj888qbLJxbdzXvP1nwQdLZdsFd"]}'
