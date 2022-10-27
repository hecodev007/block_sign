package domain

type DictListInfo struct {
	DictName  string `json:"dict_name"`
	DictValue int    `json:"dict_value"`
}

type DictTag struct {
	DictType string `json:"dict_typ"`
}

type DictInfo struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}
