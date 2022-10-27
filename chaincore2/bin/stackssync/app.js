


const app=require("express")();
c32 = require('c32check')
const bodyParser = require('body-parser');
app.use(bodyParser.json())
port=10004
app.post("/conversionAddress",function (req,res) {
    console.log(req.body)
    const body = req.body
    if(body.CoinName=="btc"){
        stackAddress = c32.b58ToC32(body.Address)
        console.log("转换结果:::"+stackAddress)
        res.json({"code":0,"message":"生成地址成功","data":stackAddress})
        return
    }
    if(body.CoinName=="stacks"){
        btcAddress=c32.c32ToB58(body.Address)
        console.log("转换结果:::"+btcAddress)
        res.json({"code":0,"message":"生成地址成功","data":btcAddress})
        return
    }
    res.json({"code":0,"message":"币名错误","data":btcAddress})

})



app.listen(port,()=>console.log("server starts "+port))
