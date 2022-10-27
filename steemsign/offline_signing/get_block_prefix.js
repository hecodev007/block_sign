const steem = require('@steemit/steem-js');

steem.api.setOptions({
    url: "https://api.steemit.com",
    retry: false,
    useAppbaseApi: true,
});

function getProperties() {
    return new Promise((resolve, reject) => {
      steem.api.getDynamicGlobalProperties(function(err, result) {
        if(!err) {
          resolve(result);
        }
        else
          console.log(err);
      });
    });
}

async function getBlockAndPrefix(properties) {
    let refBlockNum = (properties.last_irreversible_block_num - 1) & 0xFFFF;
    steem.api.getBlockHeaderAsync(properties.last_irreversible_block_num).then((block) => {
        let headBlockId = block ? block.previous : '0000000000000000000000000000000000000000';
        console.log('ref_block_num: ' + refBlockNum);
        console.log('ref_block_prefix: ' + new Buffer(headBlockId, 'hex').readUInt32LE(4));
    });
}

getProperties()
.then(getBlockAndPrefix);