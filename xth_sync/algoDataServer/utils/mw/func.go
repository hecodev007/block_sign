package mw

func AccountidToAddr(aid string)string{
	addr := newAddress()
	ok :=addr.from_acc(aid)
	if !ok {
		return ""
	}
	return addr.toString()
}
