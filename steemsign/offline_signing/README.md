# offline_signing

## Simple offline transfer transaction signing tool

Because Steem uses TaPoS and transaction expiration times that are built into the transactions, you must first get a recent block number and block prefix with a computer that is online and can connect to the network. Once you obtain these from a separate online computer they will be good for use for the next 64k blocks (or roughly 2.2 days).

The transaction generated must be broadcast within 1 hour after being created or it will expire.

## Building:

`
docker build -t="steemit/offline_signing:latest" .
`

(alternatively, you can just pull in a pre-built image with `docker pull steemit/offline_signing:latest`)

## Running:

1. On a computer that is online, to collect the necessary recent reference block number and prefix, run:

- `docker run -it steemit/offline_signing:latest`

- Collect the two numbers and save them for the next step.

2. From an offline computer that already has the docker image built and available, run:

- `docker run -it steemit/offline_signing:latest /bin/bash sign_offline_transfer.sh <ref block num> <prefix> <from account> <to account> <amount>`

- The software will ask for your private key WIF. This can be either your active or owner key private key for a transfer operation.

- Note: the <amount> must have a precision of 3 - example: 10 STEEM would be 10.000

- Example of a correct command:

`docker run -it steemit/offline_signing:latest /bin/bash sign_offline_transfer.sh 28468 851006437 alice bob 10.000`

3. Take the generated signed transaction and broadcast it from an online computer. You have 1 hour to complete this before the transaction is expired and is no longer valid. It can be brodcasted simply using `curl`

- Insert the transaction in this command where it says INSERT_TRANSACTION_HERE, leave the brackets around it:

`curl -s --data '{"jsonrpc":"2.0", "method":"condenser_api.broadcast_transaction", "params":[INSERT_TRANSACTION_HERE], "id":1}' https://api.steemit.com`

- Example of fully filled in command:

`curl -s --data '{"jsonrpc":"2.0", "method":"condenser_api.broadcast_transaction", "params":[{"ref_block_num":28468,"ref_block_prefix":851006437,"expiration":"2020-01-28T21:45:21","operations":[["transfer",{"from":"alice","to":"bob","amount":"10.000 STEEM","memo":""}]],"extensions":[],"signatures":["1f422e9f579951b9bf8333a9d419ad05dbdd9ac990b8aef5f9ac739ab698a7eef805143d9cdc43df620f15f4de0a06bc90cef3dd5ac07f77050a1c292b719d01f8"]}], "id":1}' https://api.steemit.com`

- A successful submission will return an empty json result, like this: `{"jsonrpc":"2.0","result":{},"id":1}`. If there is an error (such as exceeding the expiration time), you will instead receive that from the node.

4. Verify that the transaction was successful using either a block explorer (like steemd.com/@accountname), or on the Steemit Wallet site (steemitwallet.com/@accountname). 

## Alternative method without docker

If you don't have or want to use docker, you can also run this repo on any system that has Node 8.7+ and NPM installed.

To setup, run: `npm install`

Commands to run:

`node get_block_prefix.js`

`node sign_offline_transfer.js <ref block num> <prefix> <from account> <to account> <amount>`