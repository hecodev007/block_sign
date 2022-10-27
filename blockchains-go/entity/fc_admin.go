package entity

type FcAdmin struct {
	Id            int     `json:"id" xorm:"not null pk autoincr INT(11)"`
	Username      string  `json:"username" xorm:"not null comment('用户名') unique VARCHAR(25)"`
	Password      string  `json:"password" xorm:"not null VARCHAR(64)"`
	EcSalt        string  `json:"ec_salt" xorm:"not null VARCHAR(5)"`
	Sex           int     `json:"sex" xorm:"default 0 comment('性别') TINYINT(1)"`
	Addtime       int     `json:"addtime" xorm:"INT(11)"`
	Status        int     `json:"status" xorm:"default 1 comment('状态') INT(11)"`
	Nickname      string  `json:"nickname" xorm:"VARCHAR(25)"`
	GroupId       int     `json:"group_id" xorm:"default 1 comment('权限组') INT(11)"`
	LastLogin     int     `json:"last_login" xorm:"not null INT(11)"`
	LastIp        string  `json:"last_ip" xorm:"not null VARCHAR(15)"`
	Locktime      int     `json:"locktime" xorm:"INT(11)"`
	Email         string  `json:"email" xorm:"comment('邮箱') VARCHAR(30)"`
	Mobile        string  `json:"mobile" xorm:"not null unique VARCHAR(20)"`
	Areacode      string  `json:"areacode" xorm:"not null default '' comment('区号') VARCHAR(15)"`
	DeviceSn      string  `json:"device_sn" xorm:"default '' comment('设备号') VARCHAR(64)"`
	Lng           float64 `json:"lng" xorm:"DOUBLE(12,8)"`
	Lat           float64 `json:"lat" xorm:"DOUBLE(12,8)"`
	Address       string  `json:"address" xorm:"VARCHAR(255)"`
	AppLoginToken string  `json:"app_login_token" xorm:"comment('app登录token') VARCHAR(32)"`
	Ua            string  `json:"ua" xorm:"default 'browser' VARCHAR(200)"`
	FromCli       string  `json:"from_cli" xorm:"comment('客户端来源：ios, browser') VARCHAR(50)"`
	Idno          string  `json:"idno" xorm:"comment('身份证号') VARCHAR(30)"`
	Idcard        string  `json:"idcard" xorm:"comment('身份证正面图像路径') VARCHAR(200)"`
	Privilege     string  `json:"privilege" xorm:"comment('权限信息') TEXT"`
}
