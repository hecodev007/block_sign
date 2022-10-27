package labs

import "math/big"

type Transaction struct{
	DeployHash string
	From string
	To string
	Source string
	Target string
	Amounb big.Int
	Gas big.Int
	Id uint64
}
func (tx *Transaction) Tobytes()([]byte,error){
	return  nil,nil
}