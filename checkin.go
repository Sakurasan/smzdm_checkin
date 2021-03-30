package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/tidwall/gjson"
)

var (
	_url            = "https://zhiyou.smzdm.com/user/checkin/jsonp_checkin"
	smzdm_cookie    = ""
	sendkey         = ""
	default_headers = map[string]string{
		"Accept":     "*/*",
		"Host":       "zhiyou.smzdm.com",
		"Referer":    "https://www.smzdm.com/",
		"User-Agent": "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/88.0.4324.192 Safari/537.36",
	}
)

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
}
func main() {
	req, _ := http.NewRequest("GET", _url, nil)
	for k, v := range default_headers {
		req.Header.Set(k, v)
	}
	req.Header.Set("Cookie", go.Getenv("SMZDM_COOKIE"))
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Println("clirnt.Do()", err)
		return
	}
	byteresp, _ := ioutil.ReadAll(resp.Body)
	fmt.Println(string(byteresp))
	if gjson.GetBytes(byteresp, "error_code").Int() == 0 {
		log.Println("敲到成功")
	}

}
