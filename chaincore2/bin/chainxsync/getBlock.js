const { ApiBase, HttpProvider, WsProvider } = require('chainx.js');
const { Observable, timer } = require('rxjs');
const { timeout, retryWhen, delayWhen, take } = require('rxjs/operators');
var http = require('http');    
var url = require('url');


(async () => {
  // 使用 http 连接
  const api = new ApiBase(new WsProvider('wss://w1.chainx.org.cn/ws'));
  // 使用 websocket 连接
  //const api = new ApiBase(new WsProvider('ws://3.113.247.189:8087/'))

  await api.isReady;

  async function getTransfers(blockNumber) {
    const blockHash = await api.rpc.chain.getBlockHash(blockNumber);
    const block = await api.rpc.chain.getBlock(blockHash);

    const estrinsics = block.block.extrinsics;
    const transfers = [];

    for (let i = 0; i < estrinsics.length; i++) {
      const e = estrinsics[i];

      const allEvents = await api.query.system.events.at(blockHash);
      events = allEvents
        .filter(({ phase }) => phase.type === 'ApplyExtrinsic' && phase.value.eqn(i))
        .map(event => {
          const o = event.toJSON();
          o.method = event.event.data.method;
          return o;
        });
      result = events[events.length - 1].method;

      transfers.push({
        index: i,
        blockHash: blockHash.toHex(),
        blockNumber: blockNumber,
        result,
        tx: {
          signature: e.signature.toJSON(),
          method: e.method.toJSON(),
        },
        events: events,
        txHash: e.hash.toHex(),
      });
    }

    return transfers;
  }
  
    // 错误重试
    const getTransfersWithRetry = async function getTransfersWithRetry(blockNumber) {
      return new Observable(async subscriber => {
        try {
          const result = await getTransfers(blockNumber);
          subscriber.next(result);
          subscriber.complete();
        } catch (error) {
          if (!subscriber.closed) {
            console.log(error);
            subscriber.error(error);
          }
        }
      })
        .pipe(
          timeout(3000),
          retryWhen(errors => {
            console.log('发生了一个错误，等待重试');
            return errors.pipe(delayWhen(val => timer(3000)));
          })
        )
        .toPromise();
    };

    // 根据指定高度获取数据
    async function getheight(res, height) {
      try {
        const result = await getTransfersWithRetry(height)
        //console.log(JSON.stringify(result))
        res.end(JSON.stringify({"result":result}));
      } catch (error) {
        console.log(error)
        res.end({"result":null});
      }
    };

    async function queryTimeStamp(res, hash) {
      try {
        const time = await api.query.timestamp.now.at(hash);
        res.end(JSON.stringify({"result":{"hash":hash, "timestamp":time.toNumber()}}));
      } catch (error) {
        console.log(error)
        res.end({"result":null});
      }
    }  
    
    http.createServer(function(req, res) { 
      var pathobj = url.parse(req.url, true)
      switch(pathobj.pathname){
      case '/':
      case '/parseheight':
        var height = pathobj.query.height;
        getheight(res, height);
        break;
      case '/timestamp':
          var hash = pathobj.query.hash;
          queryTimeStamp(res, hash);
          break;
      }
    }).listen(8090)

  //console.log(JSON.stringify(await getTransfersWithRetry(3203767)));
  //console.log('over');
})();



