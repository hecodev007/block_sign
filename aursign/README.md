# aur-sign

###发送交易
curl --location --request POST '127.0.0.1:14025/v1/aur/transfer' \
--header 'Content-Type: application/json' \
--data-raw 
'{"from_address": "0xC2Bba9Ae55bc6A7d13797e3FD393C406702aAF60",
"to_address": "0xB6306d3a97B895EDf857375673a5e839106cdd0c",
"amount": "1.02",
"fee":1
}'


