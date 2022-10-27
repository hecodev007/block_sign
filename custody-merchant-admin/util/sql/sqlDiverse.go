package sql

import "bytes"

type SqlAndVars struct {
	Sql  string
	V    interface{}
	Flag bool
}

type SqlBuilder struct {
	Builder bytes.Buffer
	SqlVar  []SqlAndVars
}

// SqlAdd
// 添加没有参数的sql
func (build *SqlBuilder) SqlAdd(sql string) *SqlBuilder {
	build.Builder.WriteString(sql)
	return build
}

// SqlAddParse
// 添加有参数的sql
func (build *SqlBuilder) SqlAddParse(sql string, v ...interface{}) *SqlBuilder {

	build.Builder.WriteString(SqlParse(sql, v))
	return build
}

// SqlAddParseExc
// 添加有参数的sql
func (build *SqlBuilder) SqlAddParseExc(sql string, b bool, v ...interface{}) *SqlBuilder {
	if b {
		build.Builder.WriteString(SqlParse(sql, v))
	}
	return build
}

// SqlEnd
// 给 sql 后面添加带参数的sql
func (build *SqlBuilder) SqlEnd(sql string, vars interface{}, ex bool) *SqlBuilder {
	if ex {
		build.Builder.WriteString(SqlParse(sql, []interface{}{vars}))
	}
	return build
}

// SqlWhere
// 给 sql 添加 where,
// sql: " id = ? and name = ? "
func (build *SqlBuilder) SqlWhere(sql string, vars interface{}, ex bool) *SqlBuilder {
	build.Builder.WriteString(" where ")

	if ex {
		build.Builder.WriteString(SqlParse(sql, []interface{}{vars}))
		build.Builder.WriteString(" ")
	}
	return build
}

// SqlWhereVars
// 给 sql 添加 where 后的参数...
func (build *SqlBuilder) SqlWhereVars(sql string, vars interface{}, ex bool) *SqlBuilder {
	if ex {
		build.Builder.WriteString(SqlParse(sql, []interface{}{vars}))
	}
	return build
}

// SqlWhereMap
// 给 sql 添加 where 后的参数...
func (build *SqlBuilder) SqlWhereMap(mp map[string]interface{}) *SqlBuilder {
	i := 0
	for key, v := range mp {
		var (
			bf   bytes.Buffer
			vars []interface{}
		)
		vars = append(vars, v)
		bf.WriteString(" ")
		if i == 0 {
			bf.WriteString("and ")
		}
		bf.WriteString(key)
		bf.WriteString(" = ? ")
		build.Builder.WriteString(SqlParse(bf.String(), vars))
		i++
	}
	return build
}

// SqlList
// sql + 参数...
func (build *SqlBuilder) SqlList() *SqlBuilder {

	for _, v := range build.SqlVar {
		if v.V != nil {
			var bf bytes.Buffer
			bf.WriteString(" ")
			bf.WriteString(v.Sql)
			build.Builder.WriteString(SqlParse(bf.String(), []interface{}{v.V}))
		} else {
			build.Builder.WriteString(v.Sql)
		}
	}
	return build
}

func (build *SqlBuilder) ToSqlString() string {

	return build.Builder.String()
}
