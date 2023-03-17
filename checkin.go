package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

const (
	Germany      Country = "Germany"
	UnitedStates Country = "United States"
	China        Country = "China"
)

var (
	_url            = "https://zhiyou.smzdm.com/user/checkin/jsonp_checkin"
	qmsgurl         = "https://qmsg.zendee.cn/send/"
	smzdm_cookie    = ""
	qmsgkey         = ""
	default_headers = map[string]string{
		"Accept":  "*/*",
		"Host":    "zhiyou.smzdm.com",
		"Referer": "https://www.smzdm.com/",
		// "User-Agent": "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/108.0.0.0 Safari/537.36",
		// "User-Agent": "smzdm_android_V8.7.8 rv:456 (Nexus 5;Android6.0.1;zh)smzdmapp",
		"User-Agent": "smzdm/134.2 CFNetwork/1206 Darwin/20.1.0",
	}

	timez = map[Country]string{
		Germany:      "Europe/Berlin",
		UnitedStates: "America/Los_Angeles",
		China:        "Asia/Shanghai",
	}
)

type Country string

func (c Country) TimeZoneID() string {
	if id, ok := timez[c]; ok {
		return id
	}
	return timez[China]
}

type checkinType struct {
	ErrorCode int    `json:"error_code,omitempty"`
	ErrorMsg  string `json:"error_msg,omitempty"`
	Data      struct {
		AddPoint                  int    `json:"add_point,omitempty"`
		CheckinNum                string `json:"checkin_num,omitempty"`
		Point                     int    `json:"point,omitempty"`
		Exp                       int    `json:"exp,omitempty"`
		Gold                      int    `json:"gold,omitempty"`
		Prestige                  int    `json:"prestige,omitempty"`
		Rank                      int    `json:"rank,omitempty"`
		Slogan                    string `json:"slogan,omitempty"`
		Cards                     string `json:"cards,omitempty"`
		CanContract               int    `json:"can_contract,omitempty"`
		ContinueCheckinDays       int    `json:"continue_checkin_days,omitempty"`
		ContinueCheckinRewardShow bool   `json:"continue_checkin_reward_show,omitempty"`
	} `json:"data,omitempty"`
}

func initCheck() {
	if os.Getenv("SMZDM_COOKIE") == "" {
		panic("SMZDM_COOKIE 为空")
	}
	if os.Getenv("QMSGKEY") == "" {
		fmt.Println("QMSGKEY 未设置，无失败通知")
	} else {
		qmsgkey = os.Getenv("QMSGKEY")
	}
}
func main() {
	initCheck()
	req, _ := http.NewRequest("GET", _url, nil)
	for k, v := range default_headers {
		req.Header.Set(k, v)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (iPhone; CPU iPhone OS 13_2_3 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/13.0.3 Mobile/15E148 Safari/604.1 Edg/108.0.0.0")
	req.Header.Set("Cookie", os.Getenv("SMZDM_COOKIE"))
	// req.Header.Set("Cookie", test_cookie)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Println("clirnt.Do()", err)
		return
	}
	// byteresp, _ := ioutil.ReadAll(resp.Body)
	var ct checkinType
	var errs error
	if errs = json.NewDecoder(resp.Body).Decode(&ct); errs != nil && errs != io.EOF {
		switch et := err.(type) {
		case *json.UnmarshalTypeError:
			log.Printf("UnmarshalTypeError: Value[%s] Type[%v]\n", et.Value, et.Type)
		case *json.InvalidUnmarshalError:
			log.Printf("InvalidUnmarshalError: Type[%v]\n", et.Type)
		default:
			log.Println(errs)
		}
	}

	switch ct.ErrorCode {
	case 0:
		ltime, _ := time.LoadLocation(China.TimeZoneID())
		fmt.Println("东八区时间:", time.Now().Local().In(ltime).Format("2006-01-02 15:04:05"))
		log.Println("张大妈签到完毕!", ct.ErrorCode, ct.Data.Slogan)
		msg := fmt.Sprintf("\n⭐⭐⭐签到成功%s天⭐⭐⭐\n🏅🏅🏅金币[%d]\n🏅🏅🏅积分[%d]\n🏅🏅🏅经验[%d]\n🏅🏅🏅等级[%d]\n🏅🏅补签卡[%s]",
			ct.Data.CheckinNum, ct.Data.Gold, ct.Data.Point, ct.Data.Exp, ct.Data.Rank, ct.Data.Cards)
		log.Println(msg)
	default:
		s := fmt.Sprintf("张大妈签到失败 %s ErrCode:%d,ErrMsg:%s", time.Now().Format("2006-01-02"), ct.ErrorCode, ct.ErrorMsg)
		log.Println(s)
		Send("签到失败,请从浏览器手动签到一次,并更新cookies")
	}

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
