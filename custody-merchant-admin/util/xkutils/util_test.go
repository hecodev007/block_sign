package xkutils

import (
	"crypto/hmac"
	"crypto/sha256"
	"custody-merchant-admin/internal/domain"
	"custody-merchant-admin/module/emails"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"reflect"
	"regexp"
	"testing"
	"time"
)

func TestUtil(t *testing.T) {
	TryCatch{}.Try(func() {
		println("do something buggy")
		panic(MyError{})
	}).Catch(MyError{}, func(err error) {
		println("catch MyError")
	}).CatchAll(func(err error) {
		println("catch error")
	}).Finally(func() {
		println("finally do something")
	})
	println("done")
}

func TestUtil1(t *testing.T) {
	var err error
	var name []string
	OptsToDo{}.NilHave(err == nil, func() {
		name = append(name, "2")
	}).NilHave(err == nil, func() {
		name = append(name, "3")
	})
	fmt.Println(name)
}

func TestUtil2(t *testing.T) {
	w := domain.AssetsSelect{}
	ty := reflect.TypeOf(w)
	for i := 0; i < ty.NumField(); i++ {
		fmt.Printf(ty.Field(i).Tag.Get("json"))
	}
}

func TestDeal(t *testing.T) {
	build := new(StringBuilder)
	s := build.StringBuild("1111").StringBuild("1111 %d", 1).ToString()
	fmt.Println(s)
}
func TestDeal2(t *testing.T) {
	location, err := time.ParseInLocation("2006-01-02", "2021-11-10", time.Local)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Println(location)

}

func TestDeal1(t *testing.T) {
	var build = new(StringBuilder)

	build.StringBuild("select sum(valuation) as valuation,sum(nums) as nums from assets").
		StringBuild(" where (select count(user_service_role_audit.service_id) from user_service_role_audit where user_service_role_audit.user_id = %d and user_service_role_audit.service_id = assets.service_id limit 1) > 0 ", 1)
	fmt.Println(build.ToString())
}

// FilteredSQLInject
// 正则过滤sql注入的方法
// 参数 : 要匹配的语句
func FilteredSQLInject(toMatchStr ...string) bool {
	// 过滤 ‘
	// ORACLE 注解 --  /**/
	// 关键字过滤 update ,delete
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

func TestUuId(t *testing.T) {

	var build = NewUUId("serial_no")
	fmt.Println(build)
	build = NewUUId("serial_no")
	fmt.Println(build)
}

func TestOrderNo(t *testing.T) {
	// key=value&....
	st := ComputeHmac256(
		"chain=TRX&client_id=6ee58a79189c4143b7013e116f9cd971903cc23ef27c4c2b96e6a99b6fed5da3&coin=USDT-TRC20&coin_nums=1&from_address=TTkyDdFatneCgA9tFXXRNKuXw2ByAjrJYL&memo=TRX&nonce=bc558f9c7f8a4db682d531c747a1cd9e&to_address=THRwvcx6PkE6Wy5WgXYZarJKqSUidbAeJY&ts=1646275272",
		"6204b691-7737-4eb9-b466-391ca6d3b9d4")
	fmt.Println(st)
	//ComputeHmacSha256("client_id=testtest&coin_name=BTC", "12345")
}

func ComputeHmacSha256(message string, secret string) string {
	key := []byte(secret)
	h := hmac.New(sha256.New, key)
	h.Write([]byte(message))
	fmt.Println(h.Sum(nil))
	sha := hex.EncodeToString(h.Sum(nil))
	fmt.Println(sha)
	base64Str := base64.StdEncoding.EncodeToString([]byte(sha))
	fmt.Println(base64Str)
	return sha
}

func ComputeHmac256(data string, secret string) string {
	key := []byte(secret)
	h := hmac.New(sha256.New, key)
	_, err := h.Write([]byte(data))
	if err != nil {
		return ""
	}
	return fmt.Sprintf("%x", h.Sum(nil))
}

func TestEmail(t *testing.T) {
	em := emails.EmailConfig{
		IamUserName:  "new-project@hoo.com",
		Recipient:    "moyunrz@163.com",
		SmtpUsername: "AKIA3J4EEKZXW6G5RQCZ",
		SmtpPassword: "BGTAH2cr4Ief0SdAip2G2PwP2NH8gjQvIvqp1lANTYfk",
		Host:         "email-smtp.ap-northeast-1.amazonaws.com",
		Port:         587,
		Title:        "laoqiu mail",
	}

	email, err := em.SendEmail("[hoo] code 000000")
	if err != nil {
		return
	}

	fmt.Printf("%v", email)
}
