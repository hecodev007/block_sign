#测试curl

curl --location --request POST 'http://127.0.0.1:8042/sign' \
--header 'Content-Type: application/json' \
--data-raw '{
"account": "eosio.token",
"actor": "xutonghua111",
"chain_id": "d5a3d18fbb3c084e3b1f3fa98c21014b5f3db536cc15d08f9f6479517c6a3d86",
"coinName": "eos",
"data": "a592d5618c8b8dc7ce07000000000100a6823403ea3055000000572d3ccdcd01104230bab149b3ee00000000a8ed323228104230bab149b3ee90558c8653da303d40420f000000000004454f5300000000073437393639383600",
"eos_code": "eosio.token",
"expiration": "2022-01-05 20:44:21 +0800 CST",
"hash": "ed3de7f0568659f1b94a2a77ef873ca2",
"mchId": "hoo",
"orderId": "HsgwHy8+UvfSepHNISPmcVxRVzjJJRa7bWmab9UYNEw+0WbwUQ==_1620237362",
"public_key": "EOS6VeUZo93nzcmhK3HfQaXBsiw9tsd6hPfU2QwS2adpYQqM9G2Rt",
"ref_block_num": 35724,
"ref_block_prefix": 130992013
}'