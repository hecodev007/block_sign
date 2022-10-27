package sql

import (
	"database/sql/driver"
	"fmt"
	"reflect"
	"regexp"
	"time"
	"unicode"
)

var (
	sqlRegexp                = regexp.MustCompile(`\?`)
	numericPlaceHolderRegexp = regexp.MustCompile(`\$\d+`)
)

// SqlParse 反射获取sql字段对应的值
func SqlParse(sql string, vars []interface{}) string {
	var (
		sqlStr          string
		formattedValues []string
	)

	for _, value := range vars {
		indirectValue := reflect.Indirect(reflect.ValueOf(value))
		if indirectValue.IsValid() {
			value = indirectValue.Interface()
			if t, ok := value.(time.Time); ok {
				formattedValues = append(formattedValues, fmt.Sprintf("'%v'", t.Format("2006-01-02 15:04:05")))
			} else if b, ok := value.([]byte); ok {
				if str := string(b); isPrintable(str) {
					formattedValues = append(formattedValues, fmt.Sprintf("'%v'", str))
				} else {
					formattedValues = append(formattedValues, "'<binary>'")
				}
			} else if r, ok := value.(driver.Valuer); ok {
				if value, err := r.Value(); err == nil && value != nil {
					formattedValues = append(formattedValues, fmt.Sprintf("'%v'", value))
				} else {
					formattedValues = append(formattedValues, "NULL")
				}
			} else {
				v := fmt.Sprintf("'%v'", value)
				if !FilteredSQLInject(v) {
					formattedValues = append(formattedValues, "NULL")
				} else {
					formattedValues = append(formattedValues, v)
				}
			}
		} else {
			formattedValues = append(formattedValues, "NULL")
		}
	}

	// differentiate between $n placeholders or else treat like ?
	if numericPlaceHolderRegexp.MatchString(sql) {
		sqlStr = sql
		for index, value := range formattedValues {
			placeholder := fmt.Sprintf(`\$%d`, index+1)
			sqlStr = regexp.MustCompile(placeholder).ReplaceAllString(sqlStr, value)
		}
	} else {
		formattedValuesLength := len(formattedValues)
		for index, value := range sqlRegexp.Split(sql, -1) {
			sqlStr += value
			if index < formattedValuesLength {
				sqlStr += formattedValues[index]
			}
		}

	}

	return sqlStr
}

func isPrintable(s string) bool {
	for _, r := range s {
		if !unicode.IsPrint(r) {
			return false
		}
	}
	return true
}

// FilteredSQLInject
// 正则过滤sql注入的方法
// 参数 : 要匹配的语句
func FilteredSQLInject(toMatchStr ...string) bool {
	//过滤 ‘
	//ORACLE 注解 --  /**/
	//关键字过滤 update ,delete
	// 正则的字符串, 不能用 " " 因为" "里面的内容会转义
	str := `(?:')|(?:--)|(/\\*(?:.|[\\n\\r])*?\\*/)|(\b(select|update|and|or|delete|insert|trancate|char|chr|into|substr|ascii|declare|exec|count|master|into|drop|execute)\b)`
	re, err := regexp.Compile(str)
	if err != nil {
		return false
	}
	for _, s := range toMatchStr {
		bl := re.MatchString(s)
		if bl {
			return true
		}
	}
	return false
}
