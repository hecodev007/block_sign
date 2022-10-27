package daohc

import (
	"github.com/group-coldwallet/chaincore2/common"
	"github.com/group-coldwallet/common/log"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
)

type BlockInfo struct {
	Id             int64  `orm:"column(id);auto" description:"id"`
	Height         int64  `orm:"column(height)" description:"块高度"`
	Hash           string `orm:"column(hash)" description:"块hash"`
	FrontBlockHash string `orm:"column(previousblockhash)" description:"前一个块hash"`
	NextBlockHash  string `orm:"column(nextblockhash)" description:"后一个块hash"`
	Timestamp      int64  `orm:"column(time)" description:"出块时间"`
	Transactions   int    `orm:"column(transactions)" description:"交易总数"`
	Confirmations  int64  `orm:"column(confirmations)" description:"确认数"`
}

func NewBlockInfo() *BlockInfo {
	res := new(BlockInfo)
	return res
}

// 获取db存储最大区块高度
func GetMaxBlockIndex() (int64, error) {
	var maps []orm.Params
	o := orm.NewOrm()
	var count int64 = 0
	num, err := o.Raw("select max(height) as maxindex from block_info").Values(&maps)
	if err == nil && num > 0 {
		if maps[0]["maxindex"] == nil {
			count = beego.AppConfig.DefaultInt64("initheight", -1)
		} else {
			count = common.StrToInt64(maps[0]["maxindex"].(string))
		}
	}
	return count, nil
}

// 更新确认数
func UpdateConfirmations(height int64, confirmations int64, nextblockhash string) (int64, error) {
	o := orm.NewOrm()
	res, err := o.Raw("update block_info set confirmations = ?, nextblockhash = ? where height = ?", confirmations, nextblockhash, height).Exec()
	if err == nil {
		num, _ := res.RowsAffected()
		return num, nil
	}

	return 0, err
}

// 删除区块
func DeleteFromBlockInfo(height int64) (int64, error) {
	o := orm.NewOrm()
	res, err := o.Raw("delete from block_info where height >= ?", height).Exec()
	if err == nil {
		num, _ := res.RowsAffected()
		return num, nil
	}

	return 0, err
}

// 删除区块
func DeleteBlockInfo(height int64) (int64, error) {
	o := orm.NewOrm()
	res, err := o.Raw("delete from block_info where height = ?", height).Exec()
	if err == nil {
		num, _ := res.RowsAffected()
		return num, nil
	}

	return 0, err
}

// 根据 index 和 hash 获取块数据
func (b *BlockInfo) Print() {
	log.Debug(b)
}

// 查找指定高度索引数量
func (b *BlockInfo) GetBlockCountByIndex(index int64) int64 {
	o := orm.NewOrm()
	var maps []orm.Params
	nums, _ := o.Raw("select height from block_info where height = ?", index).Values(&maps)
	return nums
}

// 根据高度获取块数据
func (b *BlockInfo) GetBlockInfoByIndex(index int64) error {
	var maps []orm.Params
	o := orm.NewOrm()
	nums, err := o.Raw("select id, height, hash, confirmations, time from block_info where height = ?", index).Values(&maps)
	if err == nil && nums > 0 {
		b.Id = common.StrToInt64(maps[0]["id"].(string))
		b.Height = common.StrToInt64(maps[0]["height"].(string))
		b.Hash = maps[0]["hash"].(string)
		b.Confirmations = common.StrToInt64(maps[0]["confirmations"].(string))
		b.Timestamp = common.StrToTime(maps[0]["time"].(string))
	}
	return err
}

// 查找指定hash数量
func (b *BlockInfo) GetBlockCountByHash(hash string) int64 {
	o := orm.NewOrm()
	var maps []orm.Params
	nums, err := o.Raw("select height from block_info where hash = ?", hash).Values(&maps)
	if err != nil {
		log.Debug(err)
	}
	return nums
}

// 根据 index 和 hash 获取块数据
func (b *BlockInfo) GetBlockInfoByHash(hash string) error {
	o := orm.NewOrm()
	var maps []orm.Params
	nums, err := o.Raw("select id, height, hash from block_info where hash = ?", hash).Values(&maps)
	if err == nil && nums > 0 {
		b.Id = common.StrToInt64(maps[0]["id"].(string))
		b.Height = common.StrToInt64(maps[0]["height"].(string))
		b.Hash = maps[0]["hash"].(string)
	}
	return err
}

// 插入块数据
// return 影响行
func (b *BlockInfo) InsertBlockInfo() (int64, error) {
	o := orm.NewOrm()
	res, err := o.Raw("insert into block_info(height,hash,previousblockhash,nextblockhash,time,transactions,confirmations) values(?,?,?,?,?,?,?)",
		b.Height, b.Hash, b.FrontBlockHash, b.NextBlockHash, common.TimeToStr(b.Timestamp), b.Transactions, b.Confirmations).Exec()
	if err == nil {
		num, _ := res.RowsAffected()
		return num, nil
	}
	return 0, nil
}
