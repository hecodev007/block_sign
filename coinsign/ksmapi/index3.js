const express = require('express')
const app = express()
//引入body-parser
var bodyParser = require('body-parser');
app.use(express.static('public'));
//需要use的
app.use(bodyParser.json()); // for parsing application/json
app.use(bodyParser.urlencoded({
    extended: true
})); // for parsing application/x-www-form-urlencoded


const { ApiPromise, WsProvider } = require('@polkadot/api');
const { Keyring,encodeAddress } = require('@polkadot/keyring');
const { decodeAddress,checkAddress }  = require('@polkadot/util-crypto');


//===============================config==============================================//
const NodeWsUrl = 'ws://127.0.0.1:9944';
const Port = 9000;



//====================================路由定义====================================
app.get('/ping', (req, res) => res.send('Hello World!'))

app.post('/vaildaddr', function (req, res) {
    let addr = req.body.addr
    var o = {};
    if (typeof (addr) == "undefined" || addr == "" ) {
        o.code = 1000
        o.message = "empty addr"
        res.json(o)
        return
    }
    console.log("addr:", addr)
    var myPattern = new RegExp("^[a-zA-Z]"); // 以英文字母开头
    if(!myPattern.exec(addr)) {
        o.code = 1002
        o.message = "fail"
        res.json(o)
        return
    }
    if (checkAddress(addr,2)[0]){
        o.code = 200
        o.message = "ok"
        res.json(o)
        return
    }else{
        o.code = 1002
        o.message = "fail"
        res.json(o)
        return
    }
})

app.post('/balance', function (req, res) {
    let addr = req.body.addr
    var o = {};
    if (typeof (addr) == "undefined" || addr == "" ) {
        o.code = 1000
        o.message = "empty addr"
        res.json(o)
        return
    }
    console.log("addr:", addr)
    ksmGetbalance(api, addr).then(balance => {
        o.code = 200
        o.message = "ok"
        o.data = String(balance)
        res.json(o)
        return
    })
})


//todo 金额传值是整数,会有精度溢出问题。后续需要调整
app.post('/transfer', function (req, res) {
    let fromSeed = req.body.fromSeed
    let fromAddr = req.body.fromAddr
    let toAddr = req.body.toAddr
    let toAmount = req.body.toAmount
    console.log(toAmount)
    console.log(toAddr)
    var o = {};
    if (typeof (fromSeed) == "undefined" || typeof (fromAddr) == "undefined"
        || typeof (toAddr) == "undefined" || typeof (toAmount) == "undefined") {
        o.code = 1000
        o.message = "error params"
        res.json(o)
        return
    }
    if (fromSeed == "" || fromAddr == "" || toAddr == "" || toAmount == "") {
        o.code = 1000
        o.message = "error params"
        res.json(o)
        return
    }
    if (!checkAddress(toAddr,2)[0]) {
        o.code = 1001
        o.message = "error to address"
        res.json(o)
        return;
    }
    if (isContains(toAmount,'.')){
        o.code = 1001
        o.message = "error to amount,,amount is not integer"
        res.json(o)
        return;
    }


    ksmTransfer(api,fromSeed,fromAddr,toAddr,toAmount).then(hex =>{
        if (hex == "") {
            o.code = 2000
            o.message = "transfer error"
            res.json(o)
            return
        }
        o.code = 200
        o.message = "ok"
        o.data = hex
        res.json(o)
        return
    })


})
//====================================路由定义====================================


//==============================方法分离，方便以后分离文件===========================
let api
async function initApi() {
    const wsProvider = new WsProvider(NodeWsUrl);
    api = await ApiPromise.create({ provider: wsProvider });
    const now = await api.query.timestamp.now();
    console.log(`api:${now}`);
}

async function ksmGetbalance(api, addr) {
    let { data: { free: previousFree }, nonce: previousNonce } = await api.query.system.account(addr);
    console.log(`${addr} has a balance of ${previousFree}, nonce ${previousNonce}`);
    return previousFree

    // const balance = await api.query.balances.freeBalance(addr);
    //
    // console.log(`balance of ${balance}`);
    // return balance
}

async function ksmTransfer (api,fromKey,fromAddr,toAddr,toAmount) {
    const keyring = new Keyring({ type: 'sr25519' });
    const fromPair = keyring.addFromUri(fromKey);
    const pub  = fromPair.publicKey
    const fromPairAddr =  encodeAddress(pub, 2);
    console.log(fromPairAddr);
    if (fromPairAddr == fromAddr) {

        const transfer = api.tx.balances.transfer(toAddr, toAmount);
        const hash = await transfer.signAndSend(fromPair);
        console.log('Transfer sent with hash', hash.toHex());
        return hash.toHex()
    }
    return ""

}

function isContains(str, substr) {
    return str.indexOf(substr) >= 0;
}
//==============================方法分离，方便以后分离文件============================



//==============================启动服务器==============================
initApi().then(() => {
    app.listen(Port, () => console.log('Example app listening on port ',Port))
}).catch(console.error)
// //==============================启动服务器==============================