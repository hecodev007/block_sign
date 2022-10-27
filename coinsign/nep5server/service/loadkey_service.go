package service

type LoadKeyService interface {
	//加载指定目录的所有地址
	ReadNewFolder(folderpath string)
}
