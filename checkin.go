package main

import (
	"encoding/json"
	"fmt"
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
		"Accept":     "*/*",
		"Host":       "zhiyou.smzdm.com",
		"Referer":    "https://www.smzdm.com/",
		"User-Agent": "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/88.0.4324.192 Safari/537.36",
	}
	test_cookie = "__ckguid=2bj5ig7YcEpCorJFUJgrs35; r_sort_type=score; __jsluid_s=bf898ad6c1a7449815696c170abe0648; _ga=GA1.2.153444063.1616936101; Hm_lvt_9b7ac3d38f30fe89ff0b8a0546904e58=1616935583,1617097704; smzdm_user_source=395EEC66AFF744A150CC1C17FB8E4A52; _gid=GA1.2.1961060084.1617511787; zdm_qd=%7B%22referrer%22%3A%22https%3A%2F%2Fwww.baidu.com%2Flink%3Furl%3D8CViaLDpZd4J6wpyJkRdZ81tb6SR-hjTHdFH6C1qmNVkvsev1LUL97UY5Vb-KuDe%26wd%3D%26eqid%3Da02dc37b000619ad00000006606947e5%22%7D; __jsluid_h=1c6d1f221e778204c6965364df0252e3; device_id=3018514116175175973324076c4074304f0c881e30408109b603db44; sess=NmNhZjV8MTYyMTQwNTU5N3w3MjU4MTAwNjQ5fGRkOGRkOWQ4MjY4YjViMjFhZGJiZjFmMGQyOTU3OTA3fDMwMTg1MTQxMTYxNzUxNzU5NzMzMjQwNzZjNDA3NDMwNGYwYzg4MWUzMDQwODEwOWI2MDNkYjQ0fHdlYg%3D%3D; user=user%3A7258100649%7C7258100649; smzdm_id=7258100649; sensorsdata2015jssdkcross=%7B%22distinct_id%22%3A%227258100649%22%2C%22first_id%22%3A%2217878e6ddb5bb7-0b9a91d36a0156-33697709-1296000-17878e6ddb66c7%22%2C%22props%22%3A%7B%22%24latest_traffic_source_type%22%3A%22%E7%9B%B4%E6%8E%A5%E6%B5%81%E9%87%8F%22%2C%22%24latest_search_keyword%22%3A%22%E6%9C%AA%E5%8F%96%E5%88%B0%E5%80%BC_%E7%9B%B4%E6%8E%A5%E6%89%93%E5%BC%80%22%2C%22%24latest_referrer%22%3A%22%22%2C%22%24latest_landing_page%22%3A%22https%3A%2F%2Fzhiyou.smzdm.com%2Fuser%22%7D%2C%22%24device_id%22%3A%2217878e6ddb5bb7-0b9a91d36a0156-33697709-1296000-17878e6ddb66c7%22%7D; homepage_sug=a; smzdm_user_view=BE9735DE27AF0E2BDD8D30F938A17893; _zdmA.uid=ZDMA.kCE6MoRLz.1617523939.2419200; Hm_lpvt_9b7ac3d38f30fe89ff0b8a0546904e58=1617524002"
	timez       = map[Country]string{
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
	if errs = json.NewDecoder(resp.Body).Decode(&ct); errs != nil {
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
	default:
		s := fmt.Sprintf("张大妈签到失败 %s ErrCode:%d,ErrMsg:%s", time.Now().Format("2006-01-02"), ct.ErrorCode, ct.ErrorMsg)
		log.Println(s)
		Send(s)
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
