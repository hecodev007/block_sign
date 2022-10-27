package loadkeyservice

type BasicService interface {

	//加载指定目录的所有地址
	ReadNewFolder(folderpath string)

	//加载指定文件的地址
	ReadFile(fileAPath, fileBPath string)

	//加载历史遗留的旧文件(加密规则，顺序有所不同)
	ReadOleFolder(folderpath string)
}
