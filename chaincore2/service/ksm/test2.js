/* eslint-disable @typescript-eslint/require-await */
/* eslint-disable @typescript-eslint/unbound-method */
/* eslint-disable @typescript-eslint/no-var-requires */
// Required imports
const { ApiPromise, WsProvider } = require('@polkadot/api');


async function main () {
  // Initialise the provider to connect to the local node
  const provider = new WsProvider('ws://ksm.rylink.io:30944');

  // Create the API and wait until ready
  const api = await ApiPromise.create({ provider });
  
  const height = 147209;
  const blockhash = await api.rpc.chain.getBlockHash(height);
  console.log(blockhash.toHex());	
  
  // Retrieve the current block header
  const lastHdr = await api.rpc.chain.getBlock(blockhash);
  //const lastHdr = await api.rpc.chain.getHeader();
  //lastHdr.hash, lastHdr.parentHash
  //console.log(lastHdr.hash.toHex());	
  //const momentPrev = await api.query.timestamp.now.at(lastHdr.hash);
  //console.log(momentPrev);
  
  // Make our basic chain state/storage queries, all in one go
  const [metaData] = await Promise.all([
    api.rpc.state.getMetadata(blockhash),
  ]);
  //console.log(metaData);
  
  const _meat = metaData.toJSON()
  //console.log(_meat);
  const modules = _meat.metadata.V9.modules;
  console.log(modules);
  console.log("------------------------------ 1");
  
  const ar = modules.find(function(elem){
	  //return elem.name==='Balances';
     return elem.name==='System';
  });
  console.log(ar);
  console.log("------------------------------ 2");
  console.log(ar.storage.items);
  
  
  const evetns = ar.storage.items.find(function(elem){
     return elem.name==='Events';
  });
  const events = ar.events;
  console.log(`\nReceived ${events.length} events:`);
	events.forEach(function(record){
		console.log(record);
	 });
 

// This will not work, because `name` is an instance of `Text`, not a string
// const system = modules.find(m => m.name === 'system');

// This will work, because `Text.eq()` can compare against a string
}

main().catch((error) => {
  console.error(error);
  process.exit(-1);
});