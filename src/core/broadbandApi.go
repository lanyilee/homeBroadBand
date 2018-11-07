package core

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
)

type MobileData struct {
	Mobile string
	Kdtype string
}

type TypeResult struct {
	Accounttype string
	Mobileno    string
}
type KdcheckResult struct {
	KdAccount      string
	UserStatus     string
	IsManualHandle string
	IsBookRenewals string
	IsYearPackAge  string
	DateTouse      string
	LastDate       string
	BroadSpeed     string
}

type ResultData struct {
	Result        bool
	ResultMsg     string
	ResultContent json.RawMessage `json:"ResultContent"`
	ResultCode    string
	ResultCount   int
}

func (data *MobileData) BroadbandTypeApi(url string) (typeResult *TypeResult, err error) {
	jsonObj := make(map[string]interface{})
	jsonObj["Mobile"] = data.Mobile
	bytesData, err := json.Marshal(jsonObj)
	if err != nil {
		return nil, err
	}
	reader := bytes.NewReader(bytesData)
	request, err := http.NewRequest("POST", url, reader)
	if err != nil {
		return nil, err
	}
	client := http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	respBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	respObj := &ResultData{}
	json.Unmarshal(respBytes, respObj)
	if respObj.Result == true {
		//二次解析
		typeResult = &TypeResult{}
		json.Unmarshal(respObj.ResultContent, typeResult)
		return typeResult, nil
	} else {
		respStr := (string)(respBytes)
		//异步日志，优化速度
		//Logger("type api failure: "+respStr)
		return nil, err
	}
}

func (data *TypeResult) KdcheckrenewalsApi(url string) (KCheck *KdcheckResult, err error) {
	jsonObj := make(map[string]interface{})
	jsonObj["Mobile"] = data.Mobileno
	jsonObj["Kdtype"] = data.Accounttype
	bytesData, err := json.Marshal(jsonObj)
	if err != nil {
		return nil, err
	}
	reader := bytes.NewReader(bytesData)
	request, err := http.NewRequest("POST", url, reader)
	if err != nil {
		return nil, err
	}
	client := http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	respBytes, err := ioutil.ReadAll(resp.Body)
	respStr := (string)(respBytes)
	Logger(respStr)
	if err != nil {
		return nil, err
	}
	result := &ResultData{}
	json.Unmarshal(respBytes, result)
	if result.Result == true {
		//二次
		KCheck = &KdcheckResult{}
		//这里返回的是数组json[{}]结构，所以和上面那个api的处理不一样
		result.ResultContent = result.ResultContent[1 : len(result.ResultContent)-1]
		//respStr := (string)(result.ResultContent)
		//fmt.Println(respStr)
		json.Unmarshal(result.ResultContent, KCheck)
		return KCheck, nil
	} else {
		respStr := (string)(respBytes)
		//异步日志，优化速度，
		// Logger("check api failure: "+respStr)
		return nil, err
	}

}
