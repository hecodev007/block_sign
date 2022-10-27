// Based on https://github.com/labstack/echo/blob/v1/binder.go
package binder

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"net/http"
	"reflect"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"
)

// New 创建新的绑定器
func New() *binder {
	return &binder{}
}

// 构建数据结构
type binder struct {
	maxMemory int64
}

// 默认最大内存：32MB
const (
	defaultMaxMemory = 32 << 20 // 32 MB
)

// SetMaxMemory SetMaxBodySize sets multipart forms max body size 设置多部分表单最大主体尺寸
func (b *binder) SetMaxMemory(size int64) {
	b.maxMemory = size
}

// MaxMemory MaxBodySize 返回多部分表单最大主体尺寸
func (b *binder) MaxMemory() int64 {
	return b.maxMemory
}

// Bind 绑定器函数实现
func (b *binder) Bind(i interface{}, c echo.Context) (err error) {
	rq := c.Request()
	ct := rq.Header.Get(echo.HeaderContentType)
	err = echo.ErrUnsupportedMediaType
	// strings.HasPrefix ct的开头是否包涵后面echo.MIMEApplicationJSON的字符
	if strings.HasPrefix(ct, echo.MIMEApplicationJSON) {
		// 进行json编码格式
		if err = json.NewDecoder(rq.Body).Decode(i); err != nil {
			err = echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}
	} else if strings.HasPrefix(ct, echo.MIMEApplicationXML) {
		// 进行xml编码格式
		if err = xml.NewDecoder(rq.Body).Decode(i); err != nil {
			err = echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}
	} else if strings.HasPrefix(ct, echo.MIMEApplicationForm) {
		r := c.Request()
		// 进行xml编码格式
		if err = b.bindForm(r, i); err != nil {
			err = echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}
	} else if strings.HasPrefix(ct, echo.MIMEMultipartForm) {
		r := c.Request()
		if err = b.bindMultiPartForm(r, i); err != nil {
			err = echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}
	}

	return
}

// 绑定表单
func (binder) bindForm(r *http.Request, i interface{}) error {
	// 解析请求的表单
	if err := r.ParseForm(); err != nil {
		return err
	}
	// 映射表单
	return mapForm(i, r.Form)
}

// 绑定多部分表格
func (b binder) bindMultiPartForm(r *http.Request, i interface{}) error {
	if b.maxMemory == 0 {
		b.maxMemory = defaultMaxMemory
	}
	if err := r.ParseMultipartForm(b.maxMemory); err != nil {
		return err
	}
	return mapForm(i, r.Form)
}

// 映射表单
func mapForm(ptr interface{}, form map[string][]string) error {
	// reflect 反射
	// Elem 返回一个类型的元素类型。

	// 如果类型的 Kind 不是 Array、Chan、Map、Ptr 或 Slice，它会发生恐慌。
	typ := reflect.TypeOf(ptr).Elem()
	// 如果是0值或者是nil发生恐慌
	val := reflect.ValueOf(ptr).Elem()
	// NumField typ字段数
	for i := 0; i < typ.NumField(); i++ {
		// 遍历获取
		typeField := typ.Field(i)
		structField := val.Field(i)
		// 报告是否可以更改 v 的值。
		if !structField.CanSet() {
			// 只有当值可寻址并且不是通过使用未导出的结构字段获得时，才可以更改值。
			// 如果 CanSet 返回 false，则调用 Set 或任何特定于类型的
			// setter（例如 SetBool、SetInt）将发生恐慌。
			continue // 下一个循环
		}
		// Kind 返回 v 的 Kind。
		// 如果 v 为零值（IsValid 返回 false），则 Kind 返回 Invalid。
		structFieldKind := structField.Kind()

		// Get 返回与标签字符串中的键关联的值。
		// 如果标签中没有这样的键，Get 返回空字符串。
		// 如果标签不具有常规格式，则 Get 返回的值
		// 未指定。要确定标签是否
		// 显式设置为空字符串，请使用 Lookup。
		inputFieldName := typeField.Tag.Get("form")
		// 如果字段为空则循环获取
		if inputFieldName == "" {
			// 获取字段名称
			inputFieldName = typeField.Name
			// 如果“form”标签为零，我们检查该字段是否为结构体。
			// 这对 JSON 解析没有意义，但它对表单有用
			// 因为数据是扁平化的
			if structFieldKind == reflect.Struct {
				// 给结构赋值，深度遍历
				err := mapForm(structField.Addr().Interface(), form)
				if err != nil {
					// 跳出递归
					return err
				}
				continue
			}
		}
		// map根据字段名取值
		inputValue, exists := form[inputFieldName]
		if !exists {
			continue
		}
		// 标签数量 = 值的长度
		numElems := len(inputValue)
		// 映射切片 == 结构字段类型 && 标签长度 > 0
		if structFieldKind == reflect.Slice && numElems > 0 {
			// 返回此类型的特定种类。
			sliceOf := structField.Type().Elem().Kind()
			// 构建引用切片
			slice := reflect.MakeSlice(structField.Type(), numElems, numElems)
			// 循环标签数
			for i := 0; i < numElems; i++ {
				// 传入数据类型赋值
				if err := setWithProperType(sliceOf, inputValue[i], slice.Index(i)); err != nil {
					return err
				}
			}
			// 给指定字段传入值
			val.Field(i).Set(slice)
		} else {
			// 传入参数
			if err := setWithProperType(typeField.Type.Kind(), inputValue[0], structField); err != nil {
				return err
			}
		}
	}
	return nil
}

// setWithProperType 传入数据类型
func setWithProperType(valueKind reflect.Kind, val string, structField reflect.Value) error {
	switch valueKind {
	case reflect.Int:
		return setIntField(val, 0, structField)
	case reflect.Int8:
		return setIntField(val, 8, structField)
	case reflect.Int16:
		return setIntField(val, 16, structField)
	case reflect.Int32:
		return setIntField(val, 32, structField)
	case reflect.Int64:
		return setIntField(val, 64, structField)
	case reflect.Uint:
		return setUintField(val, 0, structField)
	case reflect.Uint8:
		return setUintField(val, 8, structField)
	case reflect.Uint16:
		return setUintField(val, 16, structField)
	case reflect.Uint32:
		return setUintField(val, 32, structField)
	case reflect.Uint64:
		return setUintField(val, 64, structField)
	case reflect.Bool:
		return setBoolField(val, structField)
	case reflect.Float32:
		return setFloatField(val, 32, structField)
	case reflect.Float64:
		return setFloatField(val, 64, structField)
	case reflect.String:
		structField.SetString(val)
	default:
		return errors.New("Unknown type")
	}
	return nil
}

func setIntField(val string, bitSize int, field reflect.Value) error {
	if val == "" {
		val = "0"
	}
	intVal, err := strconv.ParseInt(val, 10, bitSize)
	if err == nil {
		field.SetInt(intVal)
	}
	return err
}

func setUintField(val string, bitSize int, field reflect.Value) error {
	if val == "" {
		val = "0"
	}
	uintVal, err := strconv.ParseUint(val, 10, bitSize)
	if err == nil {
		field.SetUint(uintVal)
	}
	return err
}

func setBoolField(val string, field reflect.Value) error {
	if val == "" {
		val = "false"
	}
	boolVal, err := strconv.ParseBool(val)
	if err == nil {
		field.SetBool(boolVal)
	}
	return err
}

func setFloatField(val string, bitSize int, field reflect.Value) error {
	if val == "" {
		val = "0.0"
	}
	floatVal, err := strconv.ParseFloat(val, bitSize)
	if err == nil {
		field.SetFloat(floatVal)
	}
	return err
}
