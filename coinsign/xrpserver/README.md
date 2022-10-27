# xrpserver

#### 初始化
go mod tidy

#### XRP签名
测试案例详见test类


1. 如果是原生短语密码,使用GenAddressFromSecret方法转换EcdsaKey私钥

2. 组装XprSignTpl进行签名，目前支持XRP，尚未测试其他代币，目前没有代币测试
