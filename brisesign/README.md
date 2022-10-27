#opsign
##转账（金额进度为18位,需要传递outer_order_no）
````
curl --location --request POST '127.0.0.1:12101/v1/op/transfer' \
--header 'Content-Type: application/json' \
--data-raw '{
"outer_order_no": "13",
"from_address": "0xc2bba9ae55bc6a7d13797e3fd393c406702aaf60",
"to_address": "0xafd454fce595f050aa3ab0e0ff69a246cf7ad29e",
"amount": "18000000000000"
}'
````
##归集
````
curl --location --request POST '127.0.0.1:12101/v1/op/transferCollect' \
--header 'Content-Type: application/json' \
--data-raw '{
    "from_address": "0xc2bba9ae55bc6a7d13797e3fd393c406702aaf60",
    "to_address": "0xafd454fce595f050aa3ab0e0ff69a246cf7ad29e",
    "amount": "18000000000000"
}'
````
##查询余额（返回的数据要右移动18位）
````
curl --location --request POST '127.0.0.1:12101/v1/op/getBalance' \
--header 'Content-Type: application/json' \
--data-raw '{
"coin_name":"op",
"address":"0xC2Bba9Ae55bc6A7d13797e3FD393C406702aAF60"
}'
````
##创建地址(batchNo订单号不能相同，可以为空)
````
curl --location --request POST '127.0.0.1:12101/v1/op/createAddr' \
--header 'Content-Type: application/json' \
--data-raw '{
    "mch":"hoo",
    "coinCode":"op",
    "count":1,
    "batchNo":"nan1"
}'
````