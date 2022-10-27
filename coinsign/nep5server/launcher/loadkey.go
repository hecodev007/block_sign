package launcher

import "github.com/group-coldwallet/nep5server/service/loadkeyservice"

func LoadKeys(folderpath string) {
	loadService := loadkeyservice.NewLoadService()
	loadService.ReadNewFolder(folderpath)
}
