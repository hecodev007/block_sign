package domain

type Menu struct {
	Sort      int    `json:"sort"` // 排序值
	Id        int64  `json:"id"`
	Pid       int64  `json:"pid"`       // 父级menu
	Label     string `json:"label"`     // menu名称
	Path      string `json:"path"`      // 跳转路由
	Icon      string `json:"icon"`      // 图标
	Component string `json:"component"` // 组件路径
}

type TreeList struct {
	Sort       int         `json:"sort"`
	Id         int64       `json:"id"`
	Pid        int64       `json:"pid"`
	Label      string      `json:"label"`
	Path       string      `json:"path"`
	Component  string      `json:"component"`
	Icon       string      `json:"icon"`
	Children   []*TreeList `json:"children"`
	ActiveMenu string      `json:"activeMenu"`
	Hidden     bool        `json:"hidden"`
}

type DataRes struct {
	Data []*TreeList `json:"data"`
}
