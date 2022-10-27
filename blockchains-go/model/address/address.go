package address

//分配地址状态
type AddressStatus int

const (
	AddressStatusDel    AddressStatus = 0 //已删除
	AddressStatusNormal AddressStatus = 1 //正常
	AddressStatusAlloc  AddressStatus = 2 //已分配
)

// 分配地址类型
type AddressType int

func (a AddressType) ToInt() int {
	return int(a)
}

const (
	AddressTypeNil  AddressType = 0 //未知地址类型
	AddressTypeCold AddressType = 1 //冷地址
	AddressTypeUser AddressType = 2 //用户地址
	AddressTypeFee  AddressType = 3 //手续费地址
	AddressTypeHot  AddressType = 4 //热地址
	AddressTypeBAL  AddressType = 5 //商户余额地址
)
