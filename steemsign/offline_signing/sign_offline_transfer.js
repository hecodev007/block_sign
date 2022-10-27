const steem = require('@steemit/steem-js');
//const readline = require('readline');
//import steem from './steem-js/src';
//const steem = require('./steem-js/src');

let refBlockNum = 0;
let prefix = 0;
let fromAccount = '';
let toAccount = '';
let amountToSend = '';

let privateKey = '';

let transaction = {
    ref_block_num: 0,
    ref_block_prefix: 0,
    expiration: '',
    operations: [],
    extensions: []
}

let time = new Date();
time.setMinutes(time.getMinutes() + 1);
let expirationTime = time.toISOString().split('.')[0];
//expirationTime = '2022-06-06T00:10:31';
//let expirationTime;
async function handleArgs() {
    if (process.argv.length != 8)
        handleError('Usage: node sign_offline_transfer.js <ref block num> <prefix> <from account> <to account> <amount>');
    refBlockNum = process.argv[2];
    prefix = process.argv[3];
    fromAccount = process.argv[4];
    toAccount = process.argv[5];
    amountToSend = process.argv[6];
    privateKey = process.argv[7];
    //  expirationTime = process.argv[8];
    //  console.log("handleArgs->expirationTime:",expirationTime);
    // console.log("expirationTime1:",expirationTime1);
}

async function buildTransaction() {
    transaction.ref_block_num = parseInt(refBlockNum);
    //console.log("buildTransaction->ref_block_num:",transaction.ref_block_num);
    transaction.ref_block_prefix = parseInt(prefix);
    transaction.expiration = expirationTime;
    transaction.operations = [[
        'transfer',
        {
            from: fromAccount,
            to: toAccount,
            amount: amountToSend + ' STEEM',
            memo: ''
        }
    ]];
}

async function signTransaction() {
    // console.log('------Signed transaction-----: \n');
    // console.log("signed:",JSON.stringify(steem.auth.signTransaction(transaction, [privateKey])))
    let r = steem.auth.signTransaction(transaction, [privateKey]);
    // let s = r.signatures
    // console.log(s[0])
    // console.log(r)
    let ret = JSON.stringify(r);
    console.log(ret)
    // console.log("signTransaction->",ret.signatures);
    // console.log('\n');
    //   process.exit(1);
    return r.signatures;
}


function handleError(err) {
    console.log(`Error: ${err}`);
    process.exit(1);
}

handleArgs()
    // .then(getWifFromTerminal)
    .then(buildTransaction)
    .then(signTransaction);