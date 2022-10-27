# 数据服务注册中间件接口

## 一 接口返回类型
     正确返回：
        {
            "code": 0,                          //正确返回code=0
            "message": "ok"
        }
    错误返回：
        {
            "code": 2,                          //错误code码
            "msg": "参数错误",                   //错误类型
            "message": "插入参数币种名字为空"     //详细错误原因
        }

## 二 地址操作

### Insert
    Method: POST
    Content-Type: application/json
    路径：/v1/address/insert
    请求参数：
        {
            "name":"ht-heco",                       //主链名字
            "uid":12,                               //userId
            "url":"http://127.0.0.1",               //url
            "addresses":[
                "0x65c9067bfbe61ad035ff332148852bbe32d50270",
                "0x4ca8ec49d742885f7d48d47ffb5fe6fb7eb59ff0"
            ]                                       //需要监听的地址，每次最多只能1000个
        }

## 三 合约操作
### Insert
     Method: POST
    Content-Type: application/json
    路径：/v1/contract/insert
    请求参数：
        {
            "name":"test-heco",                     // token名字
            "contract_address":"0xtest12345678",    // 合约地址
            "coin_type":"ht-heco",                  // 主链名字
            "decimal":12                            // 精度
        }