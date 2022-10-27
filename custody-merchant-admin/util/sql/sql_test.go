package sql

import (
	"fmt"
	"strings"
	"testing"
)

func TestSql(t *testing.T) {

	var vars []interface{}
	vars = append(vars, 1, "2")
	sql := SqlParse("select * from service where id= ? and ss= ?", vars)
	fmt.Println(sql)
	sqls := SqlBuilder{}
	sqls.SqlAdd("select orders.*, coin_info.name as coin_name, chain_info.name as chain_name,service.name as service_name,service.audit_type as audit_type,order_audit.audit_result as audit_result from orders")
	sqls.SqlVar = append(sqls.SqlVar, SqlAndVars{Sql: "left join order_audit on orders.id = order_audit.order_id", V: nil})
	sqls.SqlList()
	sqls.SqlAddParse("?,?,?", 1, 1, 1)
	fmt.Println(sqls.Builder.String())

	fmt.Println(strings.ToLower("Wss"))

}
