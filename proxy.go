package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"github.com/otokaze/go-kit/log"
)

var (
	proxyCli = http.Client{Timeout: 10 * time.Second}
	proxyIPs = map[string][]*IPInfo{}
)

const (
	_getIPAPI = "http://webapi.http.zhimacangku.com/getip?"
)

type IPInfo struct {
	IP         string `json:"ip"`
	Port       int16  `json:"port"`
	ExpireTime string `json:"expire_time"`
	City       string `json:"city"`
	Isp        string `json:"isp"`
}

func GetProxyIP(pack string) (ip *IPInfo, err error) {
	var params = url.Values{}
	params.Add("num", "1")
	params.Add("pro", "0")
	params.Add("yys", "0")
	params.Add("ts", "1")
	params.Add("ys", "1")
	params.Add("cs", "1")
	params.Add("lb", "1")
	params.Add("sb", "0")
	params.Add("pb", "45")
	params.Add("mr", "1")
	params.Add("city", "0")
	params.Add("port", "1")
	params.Add("type", "2")
	params.Add("pack", pack)
	var req *http.Request
	if req, err = http.NewRequest("GET", _getIPAPI+params.Encode(), nil); err != nil {
		log.Error("http.NewRequest(GET, %s) error(%v)", _getIPAPI, err)
		return
	}
	var resp *http.Response
	if resp, err = proxyCli.Do(req); err != nil {
		log.Error("proxyCli.Do(GET, %s) error(%v)", _getIPAPI, err)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("resp.StatusCode(%d) not 200(OK)", resp.StatusCode)
		log.Error("proxyCli.Do(GET, %s) error(%v)", _getIPAPI, err)
		return
	}
	var body []byte
	if body, err = ioutil.ReadAll(resp.Body); err != nil {
		log.Error("ioutil.ReadAll() error(%v)", err)
		return
	}
	var respBody struct {
		Code    int       `json:"code"`
		Data    []*IPInfo `json:"data"`
		Msg     string    `json:"msg"`
		Success bool      `json:"success"`
	}
	if err = json.Unmarshal(body, &respBody); err != nil {
		log.Error("json.Unmarshal() error(%v)", err)
		return
	}
	if respBody.Code != 0 {
		err = fmt.Errorf("respBody.Code(%d) not 0(OK)", respBody.Code)
		log.Error("proxyCli.Do(GET, %s) error(%v) msg(%s)", _getIPAPI, err, respBody.Msg)
		return
	}
	ip = respBody.Data[0]
	proxyIPs[pack] = append(proxyIPs[pack], ip)
	return

}
