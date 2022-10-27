package domain

type SaveUserInfo struct {
	Sex                int                `json:"sex"`
	Id                 int64              `json:"id"`
	Pid                int64              `json:"pid"`
	Name               string             `json:"name" `
	Email              string             `json:"email"`
	Phone              string             `json:"phone"`
	PhoneCode          string             `json:"phone_code"`
	Password           string             `json:"password"`
	Remark             string             `json:"remark"`
	Passport           string             `json:"passport"`
	Identity           string             `json:"identity"`
	Role               int                `json:"role"`
	Menus              []int              `json:"menus"`
	Services           []int              `json:"services" description:"业务线"`                      // 业务线和角色的结构
	ServiceAuditLevels []ServiceAuditRole `json:"service_audit_level" description:"业务线和角色审核等级的结构"` // 业务线和角色的结构
}

type UserInfo struct {
	Name  string `json:"name"  description:"用户名"`
	Email string `json:"email" description:"邮箱"`
	Phone string `json:"phone" description:"手机号"`
}

type LoginInfo struct {
	Account  string `json:"account" example:"18707876666,123@qq.com" description:"账号"`
	Password string `json:"password" example:"xxxx" description:"密码"`
}

type NewPwd struct {
	Password   string `json:"password" description:"密码"`
	RePassword string `json:"re_password" description:"确认密码"`
}

type RePwd struct {
	Account    string `json:"account" description:"账号"`
	Code       string `json:"code" description:"验证码"`
	Password   string `json:"password" description:"密码"`
	RePassword string `json:"re_password" description:"确认密码"`
}

type AccountInfo struct {
	Account   string `json:"account" description:"账号:输入手机号/邮箱，手机短信请输入手机号"`
	PhoneCode string `json:"phone_code,omitempty" description:"手机号地区：中国(+86),邮箱不用输入"`
}

type UserAccount struct {
	Id             int64         `json:"id"  description:"用户Id"`
	ServiceName    string        `json:"service_name" description:"业务线拼接的名称"`
	Remark         string        `json:"remark"  description:"备注"`
	AuthIds        []int         `json:"auth_ids"  description:"管理员权限ID的数组"`
	ServiceRoleIds []ServiceRole `json:"service_role_ids" description:"业务线和角色的结构"` // 业务线和角色的结构
	Name           string        `json:"name" `
	Email          string        `json:"email"`
	Phone          string        `json:"phone"`
	PhoneCode      string        `json:"phone_code"`
	Password       string        `json:"password"`
}

type UserAccountErr struct {
	Id         int64 `json:"id"  description:"用户Id"`
	MerchantId int64 `json:"merchant_id"  description:"商户Id"`
	ClearErr   []int `json:"clear_err"`
}

type SelectUserList struct {
	Serial             int                `json:"serial"  description:"序号"`
	Id                 int64              `json:"id"  description:"用户Id"`
	Pid                int64              `json:"pid"`
	Name               string             `json:"name"`
	Sex                int                `json:"sex"`
	SexName            string             `json:"sex_name"`
	Email              string             `json:"email"`
	Phone              string             `json:"phone"`
	PhoneCode          string             `json:"phone_code"`
	RoleAndAudit       string             `json:"role_and_audit"`
	RoleAndService     string             `json:"role_and_service"`
	ServiceAuditLevels []ServiceAuditRole `json:"service_audit_level" description:"业务线和角色审核等级的结构"` // 业务线和角色的结构
	ServiceLevelName   string             `json:"service_level_name" description:"业务线-审核等级拼接的名称"`
	Role               int                `json:"role"`
	RoleName           string             `json:"role_name"`
	ServiceName        string             `json:"service_name"`
	Services           []int              `json:"services"`
	Remark             string             `json:"remark"  description:"备注"`
	Reason             string             `json:"reason"`
	State              int                `json:"state"`
	Show               int                `json:"show"`
	StateName          string             `json:"state_name"`
	Passport           string             `json:"passport"`
	Identity           string             `json:"identity"`
	PwdErr             int                `json:"pwd_err" gorm:"column:pwd_err"`
	IsTest             int                `json:"is_test" gorm:"column:is_test"`
	IsTestName         string             `json:"is_test_name" gorm:"column:is_test_name"`
	PhoneCodeErr       int                `json:"phone_code_err" gorm:"column:phone_code_err"`
	EmailCodeErr       int                `json:"email_code_err" gorm:"column:email_code_err"`
	AccountErr         string             `json:"account_err"`
	CreateTime         string             `json:"create_time"`
	LoginTime          string             `json:"login_time"`
	Menus              []int              `json:"menus"`
	IsErr              string             `json:"is_err"`
}

type SelectAdminUserList struct {
	Serial       int    `json:"serial"  description:"序号"`
	Id           int64  `json:"id"  description:"用户Id"`
	Name         string `json:"name"`
	Sex          int    `json:"sex"`
	SexName      string `json:"sex_name"`
	Email        string `json:"email"`
	Phone        string `json:"phone"`
	PhoneCode    string `json:"phone_code"`
	Role         int    `json:"role"`
	RoleName     string `json:"role_name"`
	Remark       string `json:"remark"  description:"备注"`
	Reason       string `json:"reason"`
	State        int    `json:"state"`
	Show         int    `json:"show"`
	IsMerchant   int    `json:"is_merchant"`
	StateName    string `json:"state_name"`
	Passport     string `json:"passport"`
	Identity     string `json:"identity"`
	CreateTime   string `json:"create_time"`
	LoginTime    string `json:"login_time"`
	Menus        []int  `json:"menus"`
	PwdErr       int    `json:"pwd_err" `
	PhoneCodeErr int    `json:"phone_code_err"`
	EmailCodeErr int    `json:"email_code_err"`
}

type ServiceRole struct {
	RoleId    int `json:"role_id" description:"角色Id"`
	ServiceId int `json:"service_id"  description:"业务线Id"`
}
type ServiceAuditRole struct {
	AuditLevel int `json:"audit_level" description:"审核角色等级Id"`
	ServiceId  int `json:"service_id"  description:"业务线Id"`
}

type UserIsService struct {
	ServiceId   int    `json:"service_id" description:"业务线Id"`
	ServiceName string `json:"service_name" description:"业务线名"`
}

type UserSelect struct {
	Limit       int    `json:"limit" description:"查询的列表数量"`
	Offset      int    `json:"offset" description:"查询的列表起始值"`
	UserId      int64  `json:"user_id" description:"用户Id"`
	Remark      string `json:"remark"  description:"备注"`
	ServiceName string `json:"service_name"  description:"业务线名"`
}

type SelectUserInfo struct {
	Limit      int    `json:"limit" description:"查询的列表数量"`
	Offset     int    `json:"offset" description:"查询的列表起始值"`
	Name       string `json:"name" description:"名称查询"`
	Phone      string `json:"phone" description:"手机号"`
	Account    string `json:"account"`
	MerchantId int64  `json:"merchant_id" description:"用户Id"`
	UserId     int64  `json:"user_id" description:"用户Id"`
	RoleId     int    `json:"role_id"  description:"角色Id"`
	State      int    `json:"state"  description:"状态"`
	Aid        int    `json:"aid"  description:"审核等级"`
	Sid        int    `json:"sid"  description:"业务线Id"`
}

type UserAccountLogin struct {
	Account  string `json:"account" description:"账号"`
	Password string `json:"password" description:"密码"`
	Code     string `json:"code" description:"验证码"`
}

type UserAddrSelect struct {
	CoinId  int      `json:"coin_id" description:"币种Id"`
	Title   []string `json:"title"`
	Limit   int      `json:"limit" description:"查询的列表数量"`
	Offset  int      `json:"offset" description:"查询的列表起始值"`
	Address string   `json:"address"  description:"地址"`
}
type UserAddrInfo struct {
	CoinId    int    `json:"coin_id" description:"币种Id"`
	ChainId   int    `json:"chain_id" description:"链Id"`
	UserId    int64  `json:"user_id" description:"链Id"`
	Address   string `json:"address"  description:"地址"`
	ServiceId int    `json:"service_id"  description:"业务线Id"`
}

type UserAddrList struct {
	CoinId    int    `json:"coin_id" gorm:"column:coin_id" description:"币种Id"`
	ChainId   int    `json:"chain_id" gorm:"column:chain_id" description:"链Id"`
	Id        int64  `json:"id" gorm:"column:id;"`
	CoinName  string `json:"coin_Name" gorm:"column:coin_name" description:"币种名"`
	ChainName string `json:"chain_Name" gorm:"column:chain_name" description:"链名"`
	Address   string `json:"address"  description:"地址"`
}

type UserPersonal struct {
	Id        int64  `json:"id"`
	Sex       int    `json:"sex"`
	Pid       int64  `json:"pid"`
	Name      string `json:"name" `
	Email     string `json:"email"`
	Phone     string `json:"phone"`
	PhoneCode string `json:"phone_code"`
	Role      int    `json:"role"`
	Passport  string `json:"passport"`
	Identity  string `json:"identity"`
	LoginTime string `json:"login_time"`
	Account   string `json:"account"`
	RoleName  string `json:"role_name"`
}

type RedisKeyInfo struct {
	Key string `json:"key"`
}
