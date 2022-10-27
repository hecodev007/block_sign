package service

//部分币种出账依赖节点注册 扫描utxo

type RegisterService interface {
	RegisterToNode(addrs []string) ([]byte, error)
}
