package transpush

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/group-coldwallet/blockchains-go/dao"
	"github.com/group-coldwallet/blockchains-go/entity"
	"github.com/group-coldwallet/blockchains-go/log"
	"github.com/group-coldwallet/blockchains-go/pkg/redis"
	"github.com/group-coldwallet/blockchains-go/pkg/util"
	"github.com/group-coldwallet/blockchains-go/runtime/global"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// 商户推送
type ExecPustList struct {
	// nothing
	Index     int64
	RetryList map[int64]*RetryContent
}

type RetryContent struct {
	num  int64  // 重试次数
	time int64  // 重试时间
	data []byte // 内容
}

func (r *ExecPustList) Run(reqexit <-chan bool) {
	r.Index = 0
	if r.RetryList == nil {
		r.RetryList = make(map[int64]*RetryContent, 0)
	}
	redisHelper, err := util.AllocRedisClient()
	if err != nil {
		log.Error(err)
		WaitGroupTransPush.Done()
		return
	}
	defer redisHelper.Close()

	var repushtime = []int64{10, 30, 60, 180, 720, 1800, 7200}
	var onceTick time.Duration = 10
	timer := time.NewTicker(onceTick * time.Second) // timer
	defer timer.Stop()

	log.Infof("Run ExecPushList")
	run := true
	for run {
		select {
		case s := <-reqexit:
			log.Error("ExecPushList exit", s)
			run = false
			break

		case <-timer.C:
			for k, v := range r.RetryList {
				v.time += int64(onceTick)
				if v.num >= int64(len(repushtime)) {
					delete(r.RetryList, k)
					continue
				}
				if v.time >= repushtime[v.num] {
					v.time = 0
					v.num += 1
					if r.retry_notice(v.data) {
						delete(r.RetryList, k)
					}
				}
			}

		default:
			item, err := redisHelper.Rpop("notice_list_new")
			//log.Infof("Rpop notice_list_new %s", string(item))
			if err != nil {
				if strings.Contains(err.Error(), "nil") {
					time.Sleep(time.Second * 1)
					break
				}
				log.Error(err)
				time.Sleep(time.Second * 1)
				break
			}
			log.Infof("2. pop notice_list_new item: %s", string(item))
			if item != nil {
				log.Infof("PushList item非空 %s", string(item))
				if err := r.dispossPushList(item, redisHelper); err != nil {
					log.Debug("暂时取消掉重复推送")
					//redisHelper.LeftPush("notice_list_new", string(item))
				}
			}
		}
	}
	WaitGroupTransPush.Done()
}

func (r *ExecPustList) dispossPushList(postdata []byte, redisHelper *util.RedisClient) error {
	if redisHelper == nil {
		return errors.New("mq not nil")
	}
	// 这里先用php的推送
	//redisHelper.LeftPush("notice_list", string(postdata))
	if !r.send_notice(postdata) {
		r.RetryList[r.Index] = &RetryContent{
			num:  0,
			time: 0,
			data: postdata,
		}
		r.Index++
	}

	return nil
}

func (r *ExecPustList) isForce(txId string) bool {
	if txId == "" {
		return false
	}
	value, err := redis.Client.Get(redis.CacheKeyForceRePush + strings.ToLower(txId))
	if err != nil {
		log.Errorf("%s 推送数据超过7天，判断是否强制推送失败 %v", txId, err)
		return false
	}
	log.Infof("%s 推送数据超过7天，判断是否强制推送 从缓存获取值为 %s", txId, value)
	return value != ""
}

func (r *ExecPustList) send_notice(item []byte) bool {
	log.Infof("start send notice 1")
	var data map[string]interface{}
	d := json.NewDecoder(bytes.NewReader(item))
	d.UseNumber()
	d.Decode(&data)

	app_id := ""
	if data["user_sub_id"] != nil {
		app_id = data["user_sub_id"].(json.Number).String()
	} else {
		if data["app_id"] != nil {
			app_id = data["app_id"].(json.Number).String()
		}
	}
	log.Infof("start send notice 2")
	tx_id := ""
	if data["transaction_id"] != nil {
		tx_id = data["transaction_id"].(string)
	}
	log.Infof("推送-01 %s", tx_id)

	time_stamp, _ := data["timestamp"].(json.Number).Int64()
	limit_time := time.Now().Unix() - (14 * 24 * 3600)
	task_name := "push_list"
	if time_stamp < limit_time {
		if !r.isForce(tx_id) {
			log.Infof("txid:%s,已经超过14天，暂停推送", tx_id)
			return true
		}
	}
	log.Infof("start send notice 4")
	log.Debug(tx_id, task_name)
	info := &entity.FcMch{}
	log.Infof("PushList dao.TransPushGet %s", tx_id)
	isfind, err := dao.TransPushGet(info, "select api_key, api_secret, platform from fc_mch where id = ?", app_id)
	log.Infof("PushList dao.TransPushGet 完成 %s", tx_id)
	if err != nil {
		log.Debug(err)
		return true
	}
	if !isfind {
		log.Debug(string(item), "商户信息缺失")
		return true
	}

	if info.ApiKey == "" || info.ApiSecret == "" {
		log.Debug(string(item), "密钥格式有误")
		return true
	}
	log.Infof("推送-02 %s", tx_id)

	log.Infof("start send notice 5")

	api_coin_type := ""
	if data["coin"] != nil {
		api_coin_type = strings.ToLower(data["coin"].(string))
	} else {
		api_coin_type = strings.ToLower(data["coin_type"].(string))
	}
	log.Infof("start send notice 6")
	get_token_func := func(pid int) []interface{} {
		tokens := []interface{}{}
		for _, v := range global.CoinDecimal {
			if v.Pid == pid {
				tokens = append(tokens, v.Name)
			}
		}
		return tokens
	}
	log.Infof("推送-03 %s", tx_id)

	//fix 怎么是写死的代码。测试环境跟线上ID又不一致。如果不小心把测试环境ID编译上去就会容易无法找到回调地址。
	eos_token := get_token_func(8)
	eth_token := get_token_func(5)
	qtum_token := get_token_func(34)
	rub_token := get_token_func(326)
	ont_token := get_token_func(332)
	hc_token := get_token_func(11)
	nas_token := get_token_func(346)
	bnb_token := get_token_func(349)
	if in_array(api_coin_type, eos_token) {
		api_coin_type = "eos"
	}
	if in_array(api_coin_type, eth_token) {
		api_coin_type = "eth"
	}
	if in_array(api_coin_type, qtum_token) {
		api_coin_type = "qtum"
	}
	if in_array(api_coin_type, rub_token) {
		api_coin_type = "rub"
	}
	if in_array(api_coin_type, ont_token) {
		api_coin_type = "ont"
	}
	if in_array(api_coin_type, hc_token) {
		api_coin_type = "hc"
	}
	if in_array(api_coin_type, nas_token) {
		api_coin_type = "nas"
	}
	if in_array(api_coin_type, bnb_token) {
		api_coin_type = "bnb"
	}
	log.Infof("start send notice 7")
	power := &entity.FcApiPower{}
	log.Infof("PushList L218 dao.TransPushGet %s", tx_id)
	isfind, err = dao.TransPushGet(power, "select url from fc_api_power where user_id = ? and api_id = ? and coin_name = ? and status = 2 and user_del = 1 and admin_del = 1", app_id, 8, api_coin_type)
	log.Infof("PushList L218 dao.TransPushGet 完成 %s", tx_id)

	if err != nil {
		log.Debug(err)
		return true
	}
	if !isfind {
		log.Debug(string(item), "没有接口权限")
		return true
	}
	log.Infof("start send notice 8")
	url := power.Url
	//send_data, err := sign_data(data, info.ApiKey, info.ApiSecret)
	//if err != nil || send_data == nil {
	//	log.Error(err)
	//	return false
	//}
	log.Debug(app_id, power.Url, data)
	//auth := util.NoticeSignV(data, public_key, app_id)
	//log.Debug(app_id, power.Url, auth)
	tap := map[string]interface{}{}
	tap["app_id"] = app_id
	tap["msg"] = string(item)
	tap["coin_type"] = data["coin_type"].(string)
	tap["coin"] = ""
	if data["coin"] != nil {
		tap["coin"] = data["coin"].(string)
	}
	tap["url"] = url
	tap["tx_id"] = data["transaction_id"]
	tap["confirmations"] = data["confirmations"]
	tap["add_time"] = time.Now().Unix()
	log.Infof("PushList TransPushGetDBEnginGroup dao.TransPushGet %s", tx_id)
	sqlRes, err := dao.TransPushGetDBEnginGroup().Exec("insert into fc_push_record2(app_id,msg,coin_type,coin,url,tx_id,confirmations,add_time) values(?,?,?,?,?,?,?,?)",
		tap["app_id"], tap["msg"], tap["coin_type"], tap["coin"], tap["url"], tap["tx_id"], tap["confirmations"], tap["add_time"])
	log.Infof("PushList TransPushGetDBEnginGroup dao.TransPushGet 完成 %s", tx_id)

	if err != nil {
		log.Error(err)
		return true
	}
	log.Infof("start send notice 9")
	insertId, err := sqlRes.LastInsertId()
	if err != nil {
		log.Error(err, tx_id)
		return true
	}
	log.Infof("推送-06 %s", tx_id)

	//reg, err := post3(url, send_data)
	reg, err := util.PostMapForCallBack(url, data, info.ApiKey, info.ApiSecret)
	if err != nil || reg == nil {
		log.Error(err, tx_id)
		return false
	}
	log.Infof("推送-07 %s", tx_id)

	log.Infof("数据服务推送回调结果：%s", string(reg))
	var result map[string]interface{}
	json.Unmarshal(reg, &result)
	if result["code"] != nil {
		var code int64 = 0
		switch result["code"].(type) {
		case string:
			code = strToInt64(result["code"].(string))
		case float64:
			code = int64(result["code"].(float64))
		}

		switch code {
		case 0:
			if insertId > 0 {
				log.Infof("PushList TransPushUpdate %s", tx_id)
				dao.TransPushUpdate("update fc_push_record2 set status = 1 where id = ?", insertId)
				log.Infof("PushList TransPushUpdate 完成 %s", tx_id)
			}
			log.Debug("推送成功")
		case 4:
			log.Debug("地址不存在，不处理")
		case 5:
			log.Debug("金额过低，不处理")
		default:
			log.Debug("推送失败", string(reg))
			return false
		}
		//if code == 0 {
		//	if insertId > 0 {
		//		dao.TransPushUpdate("update fc_push_record2 set status = 1 where id = ?", insertId)
		//	}
		//	log.Debug("推送成功")
		//} else {
		//	log.Debug("推送失败", string(reg))
		//	return false
		//}
	} else {
		log.Debug("推送失败", string(reg))
		return false
	}
	log.Infof("推送-08 %s", tx_id)
	return true
}

func (r *ExecPustList) retry_notice(item []byte) bool {
	var data map[string]interface{}
	d := json.NewDecoder(bytes.NewReader(item))
	d.UseNumber()
	d.Decode(&data)

	app_id := ""
	if data["user_sub_id"] != nil {
		app_id = data["user_sub_id"].(json.Number).String()
	} else {
		if data["app_id"] != nil {
			app_id = data["app_id"].(json.Number).String()
		}
	}

	tx_id := ""
	if data["transaction_id"] != nil {
		tx_id = data["transaction_id"].(string)
	}

	time_stamp, _ := data["timestamp"].(json.Number).Int64()
	limit_time := time.Now().Unix() - (9 * 24 * 3600)
	task_name := "push_list"
	if time_stamp < limit_time {
		log.Debug(string(item), "time_over")
		return true
	}

	log.Debug(tx_id, task_name)
	info := &entity.FcMch{}
	isfind, err := dao.TransPushGet(info, "select api_key, api_secret, platform from fc_mch where id = ?", app_id)
	if err != nil {
		log.Debug(err)
		return true
	}
	if !isfind {
		log.Debug(string(item), "商户信息缺失")
		return true
	}

	if info.ApiKey == "" || info.ApiSecret == "" {
		log.Debug(string(item), "密钥格式有误")
		return true
	}

	api_coin_type := ""
	if data["coin"] != nil {
		api_coin_type = strings.ToLower(data["coin"].(string))
	} else {
		api_coin_type = strings.ToLower(data["coin_type"].(string))
	}

	get_token_func := func(pid int) []interface{} {
		tokens := []interface{}{}
		for _, v := range global.CoinDecimal {
			if v.Pid == pid {
				tokens = append(tokens, v.Name)
			}
		}
		return tokens
	}

	//fix 怎么是写死的代码。测试环境跟线上ID又不一致。如果不小心把测试环境ID编译上去就会容易无法找到回调地址。
	eos_token := get_token_func(8)
	eth_token := get_token_func(5)
	qtum_token := get_token_func(34)
	rub_token := get_token_func(326)
	ont_token := get_token_func(332)
	hc_token := get_token_func(11)
	nas_token := get_token_func(346)
	bnb_token := get_token_func(349)
	if in_array(api_coin_type, eos_token) {
		api_coin_type = "eos"
	}
	if in_array(api_coin_type, eth_token) {
		api_coin_type = "eth"
	}
	if in_array(api_coin_type, qtum_token) {
		api_coin_type = "qtum"
	}
	if in_array(api_coin_type, rub_token) {
		api_coin_type = "rub"
	}
	if in_array(api_coin_type, ont_token) {
		api_coin_type = "ont"
	}
	if in_array(api_coin_type, hc_token) {
		api_coin_type = "hc"
	}
	if in_array(api_coin_type, nas_token) {
		api_coin_type = "nas"
	}
	if in_array(api_coin_type, bnb_token) {
		api_coin_type = "bnb"
	}

	power := &entity.FcApiPower{}
	isfind, err = dao.TransPushGet(power, "select url from fc_api_power where user_id = ? and api_id = ? and coin_name = ? and status = 2 and user_del = 1 and admin_del = 1", app_id, 8, api_coin_type)
	if err != nil {
		log.Debug(err)
		return true
	}
	if !isfind {
		log.Debug(string(item), "没有接口权限")
		return true
	}

	//send_data, err := sign_data(data, info.ApiKey, info.ApiSecret)
	//if err != nil || send_data == nil {
	//	log.Error(err)
	//	return false
	//}
	//log.Debug(app_id, power.Url, send_data)
	//
	//reg, err := post3(power.Url, send_data)
	reg, err := util.PostMapForCallBack(power.Url, data, info.ApiKey, info.ApiSecret)
	if err != nil || reg == nil {
		log.Error(err)
		return false
	}

	var result map[string]interface{}
	json.Unmarshal(reg, &result)
	if result["code"] != nil {
		var code int64 = 0
		switch result["code"].(type) {
		case string:
			code = strToInt64(result["code"].(string))
		case float64:
			code = int64(result["code"].(float64))
		}

		if code == 0 {
			dao.TransPushUpdate("update fc_push_record2 set status = 1 where coin_type = ? and tx_id = ? and confirmations = ?", data["coin_type"], data["transaction_id"], data["confirmations"])
			log.Debug("重试推送成功")
		} else {
			log.Debug("重试推送失败", string(reg))
			return false
		}
	} else {
		log.Debug("重试推送失败", string(reg))
		return false
	}
	return true
}

//
//func sign_data(data map[string]interface{}, api_key, api_secret string) (map[string]interface{}, error) {
//	key := api_key
//	secret := api_secret
//
//	ts := time.Now().Unix()
//	nonce := createRandomString(6)
//	params := &sign.ApiSignParams{
//		ClientId: key,
//		Ts:       ts,
//		Nonce:    nonce,
//	}
//
//	sign := &util.ApiSign{
//		ApiKey:    params.ClientId,
//		ApiSecret: secret,
//		Ts:        fmt.Sprintf("%v", params.Ts),
//		Nonce:     fmt.Sprintf("%v", params.Nonce),
//	}
//	signResult, err := sign.GetSignParams()
//	if err != nil {
//		log.Error(err)
//		return nil, err
//	}
//
//	for k, v := range signResult {
//		data[k] = v
//	}
//	return data, nil
//
//	//query := map[string]string{}
//	//for k, v := range signResult {
//	//	query[k] = fmt.Sprintf("%v", v)
//	//}
//	//for k, v := range data {
//	//	query[k] = fmt.Sprintf("%v", v)
//	//}
//	//return encodeQueryString(query), nil
//}

func post(url string, postData []byte, auth string) ([]byte, error) {
	client := &http.Client{}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(postData))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	if auth != "" {
		req.Header.Add("auth", auth)
	}

	response, err := client.Do(req)
	if response != nil {
		defer response.Body.Close()
	}
	if err != nil {
		log.Error(err)
		return nil, err
	}

	content, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	return content, nil
}

// 发送表单请求
func post2(url string, postData []byte) ([]byte, error) {
	client := &http.Client{}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(postData))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	response, err := client.Do(req)
	if response != nil {
		defer response.Body.Close()
	}
	if err != nil {
		log.Error(err)
		return nil, err
	}

	content, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	return content, nil
}

//// 发送表单请求
//func post3(urladdr string, postData map[string]interface{}) ([]byte, error) {
//	urldata := make(url.Values)
//	for k, v := range postData {
//		urldata.Set(k, fmt.Sprintf("%v", v))
//	}
//	response, err := http.PostForm(urladdr, urldata)
//	if response != nil {
//		defer response.Body.Close()
//	}
//	if err != nil {
//		log.Error(err)
//		return nil, err
//	}
//
//	content, err := ioutil.ReadAll(response.Body)
//	if err != nil {
//		return nil, err
//	}
//	return content, nil
//}

// string to int64
func strToInt64(str string) int64 {
	val, _ := strconv.ParseInt(str, 10, 64)
	return val
}

//
////产生随机字符串，主要是测试使用
//func createRandomString(len int) string {
//	var container string
//	var str = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890"
//	b := bytes.NewBufferString(str)
//	length := b.Len()
//	bigInt := big.NewInt(int64(length))
//	for i := 0; i < len; i++ {
//		randomInt, _ := rand.Int(rand.Reader, bigInt)
//		container += string(str[randomInt.Int64()])
//	}
//	return container
//}
//
////签名
//func computeHmac256(data string, secret string) string {
//	key := []byte(secret)
//	h := hmac.New(sha256.New, key)
//	_, err := h.Write([]byte(data))
//	if err != nil {
//		return ""
//	}
//	return fmt.Sprintf("%x", h.Sum(nil))
//}
//
///// 拼接query字符串
//func encodeQueryString(query map[string]string) string {
//	keys := make([]string, 0)
//	for k := range query {
//		keys = append(keys, k)
//	}
//	sort.Strings(keys)
//	var len = len(keys)
//	var lines = make([]string, len)
//	for i := 0; i < len; i++ {
//		var k = keys[i]
//		lines[i] = url.QueryEscape(k) + "=" + url.QueryEscape(query[k])
//	}
//	return strings.Join(lines, "&")
//}
