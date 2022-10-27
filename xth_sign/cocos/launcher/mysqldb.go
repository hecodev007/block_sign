package launcher

import (
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

var MysqlDB *gorm.DB

func InitDB() {
	var err error
	MysqlDB, err = gorm.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8&parseTime=True&loc=Local",
		viper.GetString("mysql.user"),
		viper.GetString("mysql.password"),
		viper.GetString("mysql.host"),
		viper.GetString("mysql.dbname")))
	if err != nil {
		logrus.Fatalf("model.Setup err: %v", err)
	}
	//表名前缀
	//gorm.DefaultTableNameHandler = func(db *gorm.DB, defaultTableName string) string {
	//	return "prefix" + defaultTableName
	//}
	// 全局禁用表名复数 默认情况下type User struct {}  默认表名是`users`
	// 如果设置为true,`User`的默认表名为`user`,使用`TableName`设置的表名不受影响
	MysqlDB.SingularTable(true)

	//闲置的连接数
	MysqlDB.DB().SetMaxIdleConns(10)

	//最大打开的连接数
	MysqlDB.DB().SetMaxOpenConns(100)

	// 启用Logger，显示详细日志
	MysqlDB.LogMode(true)

}
