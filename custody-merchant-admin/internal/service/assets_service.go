package service

import (
	"custody-merchant-admin/global"
	"custody-merchant-admin/internal/domain"
	"custody-merchant-admin/internal/service/deals"
	"strings"
	"time"
)

func FindAssetsListService(as *domain.AssetsSelect, id int64) ([]domain.AssetsInfo, int64, error) {
	return deals.FindAssetsList(as, id)
}

func FindServiceAssetsList(as *domain.AssetsSelect, id int64) ([]map[string]interface{}, int64, error) {
	return deals.FindServiceAssetsList(as, id)
}

func FindAssetsRingService(id int64) (domain.RingObj, error) {
	return deals.FindAssetsRing(id)
}

func FindAssetsByTagService(as *domain.AssetsByTag) ([]domain.AssetsTimeInfo, error) {

	var (
		timeList []domain.AssetsTimeInfo
		err      error
	)
	if as.Tag == "" {
		as.Tag = global.DAY
	}

	if as.StartTime == "" {
		next := time.Now().Add(time.Hour * (-24 * 11))
		next = time.Date(next.Year(), next.Month(), next.Day(), 0, 0, 0, 0, next.Location())
		as.StartTime = next.Format(global.YyyyMmDd)
	}
	as.Tag = strings.ToUpper(as.Tag)
	switch as.Tag {
	case global.DAY:
		if as.StartTime != as.EndTime {
			timeList, err = deals.FindAssetsByDay(as.StartTime, as.EndTime, as.UserId)
		} else {
			timeList, err = deals.FindAssetsByHour(as.StartTime, as.EndTime, as.UserId)
		}
		break
	case global.WEEK:
		timeList, err = deals.FindAssetsByWeek(as.StartTime, as.EndTime, as.UserId)
		//createTime = util.WeekByDate(nt)
		break
	case global.MONTH:
		month := strings.Split(as.StartTime, "-")
		as.StartTime = month[0] + "-" + month[1] + "-01"
		timeList, err = deals.FindAssetsByMonth(as.StartTime, as.EndTime, as.UserId)
		//createTime = fmt.Sprintf("%d-%d", nt.Year(), nt.Month())
	case global.YEAR:
		month := strings.Split(as.StartTime, "-")
		as.StartTime = month[0] + "-01" + "-01"
		timeList, err = deals.FindAssetsByYears(as.StartTime, as.EndTime, as.UserId)
		//createTime = fmt.Sprintf("%d", nt.Year())
		break
	}
	if err != nil {
		return timeList, err
	}
	return timeList, nil
}

// FindFinanceAssetsList
// 给财务的资产查询
func FindFinanceAssetsList(as *domain.AssetsSelect) (map[string]interface{}, error) {
	return deals.FindFinanceAssetsList(as)
}
