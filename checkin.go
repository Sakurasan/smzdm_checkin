package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/Sakurasan/to"
	"github.com/duke-git/lancet/cryptor"
	"github.com/duke-git/lancet/v2/random"
	"github.com/tidwall/gjson"
)

const (
	Germany      Country = "Germany"
	UnitedStates Country = "United States"
	China        Country = "China"
)

var (
	timez = map[Country]string{
		Germany:      "Europe/Berlin",
		UnitedStates: "America/Los_Angeles",
		China:        "Asia/Shanghai",
	}
	qmsgurl          = "https://qmsg.zendee.cn/send/"
	qmsgkey          = ""
	requestUrl       = "https://user-api.smzdm.com/checkin"
	Android_SIGN_KEY = "apr1$AwP!wRRT$gJ/q.X24poeBInlUJC"
	IOS_SIGN_KEY     = "zok5JtAq3$QixaA%mncn*jGWlEpSL3E1"
	cookie           = ""
	default_data     = map[string]string{
		"weixin":  "1",
		"time":    strconv.FormatInt(time.Now().Unix(), 10),
		"basic_v": "0",
		"f":       "android",
		"v":       "10.4.26",
	}
)

type Country string

func (c Country) TimeZoneID() string {
	if id, ok := timez[c]; ok {
		return id
	}
	return timez[China]
}

type SmzdmBot struct {
	Cookies string
	Sk      string
	Token   string
	sess    string
}

func (bot *SmzdmBot) cookiesToDict() map[string]string {
	re := regexp.MustCompile(`(.*?)=(.*?);`)
	cookies := re.FindAllStringSubmatch(bot.Cookies, -1)
	cookiesDict := make(map[string]string)
	for _, v := range cookies {
		cookiesDict[v[1]] = v[2]
	}
	return cookiesDict
}

func (bot *SmzdmBot) userAgent() string {
	cookiesDict := bot.cookiesToDict()
	switch cookiesDict["device_smzdm"] {
	case "iphone":
		return fmt.Sprintf("smzdm %s rv:%s (%s; iOS %s; zh_CN)/iphone_smzdmapp/%s",
			cookiesDict["device_smzdm_version"],
			cookiesDict["device_smzdm_version_code"],
			cookiesDict["device_name"],
			cookiesDict["device_system_version"],
			cookiesDict["device_smzdm_version"],
		)
	case "android":
		return fmt.Sprintf("smzdm_%s_V%s rv:%s (%s;%s;zh)smzdmapp",
			cookiesDict["device_smzdm"],
			cookiesDict["device_smzdm_version"],
			cookiesDict["device_smzdm_version_code"],
			cookiesDict["device_type"],
			strings.Title(cookiesDict["device_smzdm"])+cookiesDict["device_system_version"],
		)
	default:
		return "smzdm_android_V10.4.26 rv:866 (Redmi Note 3;Android10;zh)smzdmapp"
	}
}

func (bot *SmzdmBot) headers() map[string]string {
	headers := map[string]string{
		"User-Agent":   bot.userAgent(),
		"Content-Type": "application/x-www-form-urlencoded; charset=UTF-8",
		"Request_Key":  random.RandNumeral(8) + to.String(time.Now().Unix()),
		"Cookie":       bot.Cookies,
	}
	return headers
}

func (bot *SmzdmBot) webHeaders() map[string]string {
	headers := map[string]string{
		"Accept":          "*/*",
		"Accept-Language": "en-US,en;q=0.9",
		"Connection":      "keep-alive",
		"Cookie":          bot.Cookies,
		"Referer":         "https://m.smzdm.com/",
		"User-Agent":      "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/112.0.0.0 Safari/537.36 Edg/112.0.1722.48",
	}
	return headers
}

func (bot *SmzdmBot) signData(data map[string]string) map[string]string {
	var signStr string
	keys := make([]string, len(data))
	for k, _ := range data {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		v := data[k]
		if v != "" {
			signStr += k + "=" + v + "&"
		}
	}
	if data["f"] == "iphone" {
		signStr += "key=" + IOS_SIGN_KEY
	} else {
		signStr += "key=" + Android_SIGN_KEY
	}

	data["sign"] = strings.ToUpper(cryptor.Md5String(signStr))
	return data
}

func (bot *SmzdmBot) Data(extraData map[string]string) map[string]string {
	data := map[string]string{
		"weixin":           "1",
		"captcha":          "",
		"basic_v":          bot.cookiesToDict()["basic_v"],
		"f":                bot.cookiesToDict()["device_smzdm"],
		"v":                bot.cookiesToDict()["device_smzdm_version"],
		"touchstone_event": "",
		"time":             strconv.FormatInt(time.Now().Unix(), 10),
		"token":            bot.Token,
		"sk":               bot.Sk,
	}
	if bot.Sk != "" {
		data["sk"] = bot.Sk
	}
	if extraData != nil {
		for k, v := range extraData {
			data[k] = v
		}
	}
	return bot.signData(data)
}

func (bot *SmzdmBot) Request(method, url string, params url.Values, extraData map[string]string) (resp *http.Response, err error) {
	client := &http.Client{}
	data := bot.Data(extraData)
	keys := make([]string, len(data))
	for k, _ := range data {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var bodystr string
	for _, k := range keys {
		if data[k] != "" {
			bodystr += "&" + k + "=" + data[k]
		}

	}
	req, err := http.NewRequest(method, url, strings.NewReader(bodystr[1:]))
	header := bot.headers()
	for k, v := range header {
		req.Header.Set(k, v)
	}
	if err != nil {
		return nil, err
	}
	if params != nil {
		q := req.URL.Query()
		for k, v := range params {
			q.Add(k, v[0])
		}
		req.URL.RawQuery = q.Encode()
	}
	return client.Do(req)
}

func (bot *SmzdmBot) Checkin() {
	log.Println("========== Checkin ==========")
	resp, err := bot.Request(http.MethodPost, requestUrl, nil, nil)
	if err != nil {
		Send("签到失败:" + err.Error())
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		Send("签到失败:" + err.Error())
	}
	respstr := DecodeUnicode(body)
	errcode := gjson.Get(respstr, "error_code").String()
	errmsg := gjson.Get(respstr, "error_msg").String()
	if errcode != "0" {
		Send("签到失败:" + errmsg)
	}
	log.Println(respstr)
}
func (bot *SmzdmBot) all_reward() {
	log.Println("========== all_reward ==========")
	resp, err := bot.Request(http.MethodPost, "https://user-api.smzdm.com/checkin/all_reward", nil, nil)
	if err != nil {
		log.Println(err)
		return
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		return
	}
	respstr := DecodeUnicode(body)
	errcode := gjson.Get(respstr, "error_code").String()
	errmsg := gjson.Get(respstr, "error_msg").String()
	if errcode != "0" {
		log.Println("没有奖励", errmsg, string(body))
		return
	}
	log.Println(respstr)
}

func (bot *SmzdmBot) extra_reward() {
	log.Println("========== extra_reward ==========")
	resp, err := bot.Request(http.MethodPost, "https://user-api.smzdm.com/checkin/extra_reward", nil, nil)
	if err != nil {
		log.Println(err)
		return
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		return
	}
	respstr := DecodeUnicode(body)
	errcode := gjson.Get(respstr, "error_code").String()
	errmsg := gjson.Get(respstr, "error_msg").String()
	if errcode != "0" {
		log.Println(errmsg)
		return
	}
	log.Println(respstr)
}

func initCheck() {
	if os.Getenv("SMZDM_COOKIE") == "" {
		panic("SMZDM_COOKIE 为空")
	}
	cookie = os.Getenv("SMZDM_COOKIE")
	if os.Getenv("QMSGKEY") == "" {
		fmt.Println("QMSGKEY 未设置，无失败通知")
	} else {
		qmsgkey = os.Getenv("QMSGKEY")
	}
}
func main() {
	initCheck()
	ltime, _ := time.LoadLocation(China.TimeZoneID())
	fmt.Println("东八区时间:", time.Now().Local().In(ltime).Format("2006-01-02 15:04:05"))
	bot := SmzdmBot{
		Cookies: cookie,
	}
	bot.Checkin()
	bot.all_reward()
	bot.extra_reward()

}

func DecodeUnicode(raw []byte) string {
	str, _ := strconv.Unquote(strings.Replace(strconv.Quote(string(raw)), `\\u`, `\u`, -1))
	return str
}

func Send(msg string) {
	if len(qmsgkey) < 5 {
		log.Println("未设置Qmsg key，不发送通知")
		return
	}
	v := url.Values{}
	v.Add("msg", msg)
	req, _ := http.NewRequest(http.MethodPost, qmsgurl+qmsgkey, strings.NewReader(v.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	_, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Println(err)
	}

}
