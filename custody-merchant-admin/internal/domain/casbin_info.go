package domain

type CasbinInfo struct {
	Path   string `json:"path" form:"path"`
	Method string `json:"method" form:"method"`
}
type CasbinCreateRequest struct {
	UserId      string       `json:"user_id" form:"user_id" description:"角色ID"`
	CasbinInfos []CasbinInfo `json:"casbin_infos" description:"权限模型列表"`
}

type CasbinUpdateRequest struct {
	OldPath    string     `json:"old_path"`
	OldMethod  string     `json:"old_method"`
	CasbinInfo CasbinInfo `json:"casbin_info" description:"权限模型列表"`
}

type CasbinListResponse struct {
	List []CasbinInfo `json:"list" form:"list"`
}

type CasbinListRequest struct {
	RoleID string `json:"role_id" form:"role_id"`
}
