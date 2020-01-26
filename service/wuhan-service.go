package service

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"sync"
	"time"
)

var result string = ""

var once sync.Once

func Province(provinceName string) map[string]interface{} {
	urlStr := "http://lab.isaaclin.cn/nCoV/api/area"
	if result == "" {
		once.Do(func() {
			result = Get(urlStr)
		})
	}
	data := Response{}

	json.Unmarshal([]byte(result), &data)

	res := map[string]Result{}
	for _, r := range data.Results {
		res[r.ProvinceName] = r
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

	//for _, r := range res {
	//	names = append(names, r.ProvinceName)
	//	confirmed = append(confirmed, r.ConfirmedCount)
	//	dead = append(dead, r.DeadCount)
	//	cured = append(cured, r.CuredCount)
	//	suspected = append(suspected, r.SuspectedCount)
	//
	//	for _, city := range r.Cities {
	//		names = append(names, city.CityName)
	//		confirmed = append(confirmed, city.ConfirmedCount)
	//		dead = append(dead, city.DeadCount)
	//		cured = append(cured, city.CuredCount)
	//		suspected = append(suspected, city.SuspectedCount)
	//	}
	//}





	dataMap := map[string]interface{}{}
	dataMap["names"] = names
	dataMap["confirmed"] = confirmed
	dataMap["dead"] = dead
	dataMap["cured"] = cured
	dataMap["suspected"] = suspected

	return dataMap
}

func Trend(provinceName string) map[string]interface{} {
	if provinceName == "" {
		provinceName = "湖北省"
	}
	urlStr := "http://lab.isaaclin.cn/nCoV/api/area"
	if result == "" {
		once.Do(func() {
			result = Get(urlStr)
		})
	}
	data := Response{}

	json.Unmarshal([]byte(result), &data)

	cacheResult := []Result{}
	for _, r := range data.Results {
		if r.ProvinceName == provinceName {
			cacheResult = append(cacheResult, r)
		}
	}

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

	return dataMap
}

/*时间戳->字符串*/
func Stamp2Str(stamp int64) string{
	timeLayout := "2006-01-02 15:04:05"
	str:=time.Unix(stamp/1000, 0).Format(timeLayout)
	return str
}

func GetAllData() map[string]Result {
	urlStr := "http://lab.isaaclin.cn/nCoV/api/area"
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
	client := &http.Client{Timeout: 180 * time.Second}
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
