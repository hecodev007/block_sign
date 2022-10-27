package dip

import (
	"github.com/group-coldwallet/scanning-service/common"
)

func (d *DipService) httpGet(path string) ([]byte, error) {
	url := d.nodeCfg.Url + path
	req := common.HttpGet(url)
	return req.Bytes()
}
