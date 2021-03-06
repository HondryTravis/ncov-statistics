package service

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"sort"
	"time"
)

const second = int64(1000000000)

type CacheResult struct {
	Response Response
	ExpireTime time.Time
	HasInit bool
}

var cr CacheResult
var history CacheResult

func Province(provinceName string) map[string]interface{} {
	if provinceName == "" {
		provinceName = "湖北省"
	}

	initHistoryData()

	data := history.Response
	res := map[string]Result{}
	for _, r := range data.Results {
		if v, ok := res[r.ProvinceName]; !ok || v.UpdateTime <= r.UpdateTime {
			res[r.ProvinceName] = r
		}
	}

	names := []string{}
	confirmed := []int{}
	dead := []int{}
	cured := []int{}
	suspected := []int{}

	r, ok := res[provinceName]
	if ok == false {
		r = res["湖北省"]
	}
	if provinceName == "国外" {
		res0 := map[string]Result{}
		for _, r := range history.Response.Results {
			if r.Country != "中国" {
				if v, ok := res0[r.ProvinceName]; !ok || v.UpdateTime <= r.UpdateTime {
					res0[r.ProvinceName] = r
				}
			}
		}
		for _, v := range res0 {
			names = append(names, v.ProvinceName)
			confirmed = append(confirmed, v.ConfirmedCount)
			dead = append(dead, v.DeadCount)
			cured = append(cured, v.CuredCount)
			suspected = append(suspected, v.SuspectedCount)
		}
		dataMap := map[string]interface{}{}
		dataMap["names"] = names
		dataMap["confirmed"] = confirmed
		dataMap["dead"] = dead
		dataMap["cured"] = cured
		dataMap["suspected"] = suspected

		return dataMap
	}

	names = append(names, r.ProvinceName)
	confirmed = append(confirmed, r.ConfirmedCount)
	dead = append(dead, r.DeadCount)
	cured = append(cured, r.CuredCount)
	suspected = append(suspected, r.SuspectedCount)
	for _, city := range r.Cities {
		names = append(names, city.CityName)
		confirmed = append(confirmed, city.ConfirmedCount)
		dead = append(dead, city.DeadCount)
		cured = append(cured, city.CuredCount)
		suspected = append(suspected, city.SuspectedCount)
	}

	dataMap := map[string]interface{}{}
	dataMap["names"] = names
	dataMap["confirmed"] = confirmed
	dataMap["dead"] = dead
	dataMap["cured"] = cured
	dataMap["suspected"] = suspected

	go refreshHistoryIfExpired()

	return dataMap
}

func Trend(provinceName string) map[string]interface{} {
	if provinceName == "" {
		provinceName = "湖北省"
	}
	initHistoryData()

	data := history.Response

	cacheResult := []Result{}
	for _, r := range data.Results {
		if r.ProvinceName == provinceName {
			cacheResult = append(cacheResult, r)
		}
	}

	// sort
	sort.Slice(cacheResult, func(i, j int) bool {
		if cacheResult[i].UpdateTime < cacheResult[j].UpdateTime {
			return true
		}

		return false
	})

	dates := []string{}
	confirmed := []int{}
	dead := []int{}
	cured := []int{}
	suspected := []int{}

	for _, v := range cacheResult {
		dates = append(dates, Stamp2Str(int64(v.UpdateTime)))
		confirmed = append(confirmed, v.ConfirmedCount)
		dead = append(dead, v.DeadCount)
		cured = append(cured, v.CuredCount)
		suspected = append(suspected, v.SuspectedCount)
	}

	dataMap := map[string]interface{}{}
	dataMap["dates"] = dates
	dataMap["confirmed"] = confirmed
	dataMap["dead"] = dead
	dataMap["cured"] = cured
	dataMap["suspected"] = suspected

	go refreshHistoryIfExpired()

	return dataMap
}

func Map(provinceName string) map[string]interface{} {
	if provinceName == "" {
		provinceName = "湖北省"
	}

	resp := map[string]interface{}{}

	initHistoryData()

	data := history.Response
	res := map[string]Result{}
	for _, r := range data.Results {
		if v, ok := res[r.ProvinceName]; !ok || v.UpdateTime <= r.UpdateTime {
			res[r.ProvinceName] = r
		}
	}

	file, _ := ioutil.ReadFile("./views/maps/" + provinceName + ".json")
	str := string(file)

	resp["map"] = str

	list := []NameValuePair{}
	province := res[provinceName]
	for _, city := range province.Cities {
		list = append(list, NameValuePair{
			Name:  city.CityName,
			Value: city.ConfirmedCount,
		})
	}
	resp["list"] = list

	go refreshHistoryIfExpired()

	return resp
}

func initData() {
	now := time.Now()

	if cr.HasInit == false {
		cr = CacheResult{
			Response:   GetAllAreaFromDXY(),
			ExpireTime: now.Add(600_000_000_000), //600s
			HasInit:    true,
		}
	}
}

func initHistoryData() {
	now := time.Now()

	if history.HasInit == false {
		history = CacheResult{
			Response:   GetHistoryAreaFromDXY(),
			ExpireTime: now.Add(600_000_000_000), //600s
			HasInit:    true,
		}
	}
}

func refreshIfExpired() {
	defer func() {
		err := recover()
		if err != nil {
			log.Println(err)
		}
	}()
	now := time.Now()
	if cr.HasInit && cr.ExpireTime.Before(now) {
		cr.Response = GetAllAreaFromDXY()
		cr.ExpireTime = now.Add(600_000_000_000)
	}
}

func refreshHistoryIfExpired() {
	defer func() {
		err := recover()
		if err != nil {
			log.Println(err)
		}
	}()
	now := time.Now()
	if history.HasInit && history.ExpireTime.Before(now) {
		history.Response = GetHistoryAreaFromDXY()
		history.ExpireTime = now.Add(600_000_000_000)
	}
}

/*时间戳->字符串*/
func Stamp2Str(stamp int64) string{
	timeLayout := "2006-01-02 15:04:05"
	str:=time.Unix(stamp/1000, 0).Format(timeLayout)
	return str
}

func GetAllData() map[string]Result {
	//https://lab.isaaclin.cn/nCoV/
	urlStr := "https://lab.isaaclin.cn/nCoV/api/area"
	result := Get(urlStr)
	data := Response{}

	json.Unmarshal([]byte(result), &data)

	dataMap := map[string]Result{}

	for _, r := range data.Results {
		dataMap[r.ProvinceName] = r
	}

	return dataMap
}

func Get(url string) string {

	// 超时时间：5秒
	client := &http.Client{
		Timeout: 180 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
	resp, err := client.Get(url)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	var buffer [512]byte
	result := bytes.NewBuffer(nil)
	for {
		n, err := resp.Body.Read(buffer[0:])
		result.Write(buffer[0:n])
		if err != nil && err == io.EOF {
			break
		} else if err != nil {
			panic(err)
		}
	}

	return result.String()
}

type Response struct {
	Results []Result `json:"results"`
	Success bool     `json:"success"`
}

type Result struct {
	Country string `json:"country"`
	Cities            []City `json:"cities"`
	Comment           string `json:"comment"`
	ConfirmedCount    int    `json:"confirmedCount"`
	CuredCount        int    `json:"curedCount"`
	DeadCount         int    `json:"deadCount"`
	ProvinceName      string `json:"provinceName"`
	ProvinceShortName string `json:"provinceShortName"`
	SuspectedCount    int    `json:"suspectedCount"`
	UpdateTime        int    `json:"updateTime"`
}

type City struct {
	CityName       string `json:"cityName"`
	ConfirmedCount int    `json:"confirmedCount"`
	CuredCount     int    `json:"curedCount"`
	DeadCount      int    `json:"deadCount"`
	SuspectedCount int    `json:"suspectedCount"`
}

type NameValuePair struct {
	Name string `json:"name"`
	Value int `json:"value"`
}