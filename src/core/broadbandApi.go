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
	Mobileno string
}
type KdcheckResult struct {
	KdAccount string
	UserStatus string
	IsManualHandle string
	IsBookRenewals string
	IsYearPackAge string
	DateTouse string
	LastDate string
	BroadSpeed string

}

type ResultData struct {
	Result bool
	ResultMsg string
	ResultContent json.RawMessage  `json:"ResultContent"`
	ResultCode string
	ResultCount int
}

func (data *MobileData)BroadbandTypeApi(url string) *TypeResult{
	jsonObj := make(map[string] interface{})
	jsonObj["Mobile"] = data.Mobile
	bytesData,err:=json.Marshal(jsonObj)
	if err!=nil{
		Logger(err.Error())
		return nil
	}
	reader:=bytes.NewReader(bytesData)
	request,err:=http.NewRequest("POST",url,reader)
	if err != nil {
		Logger("BroadbandTypeApi:request:"+err.Error())
		return nil
	}
	client:=http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		Logger("BroadbandTypeApi:client:"+err.Error())
		return nil
	}
	respBytes,err:=ioutil.ReadAll(resp.Body)
	if err!=nil{
		Logger("BroadbandTypeApi:readall:"+err.Error())
		return nil
	}
	respObj:=&ResultData{}
	json.Unmarshal(respBytes,respObj)
	if respObj.Result==true{
		//二次解析
		typeResult := &TypeResult{}
		json.Unmarshal(respObj.ResultContent,typeResult)
		Logger("宽带类型查询成功")
		return typeResult
	}else{
		respStr := (string)(respBytes)
		Logger(respStr)
	}
	return nil
}

func (data *TypeResult)KdcheckrenewalsApi(url string) *KdcheckResult{
	jsonObj := make(map[string] interface{})
	jsonObj["Mobile"] = data.Mobileno
	jsonObj["Kdtype"] = data.Accounttype
	bytesData,err:=json.Marshal(jsonObj)
	if err!=nil{
		Logger(err.Error())
		return nil
	}
	reader:=bytes.NewReader(bytesData)
	request,err:=http.NewRequest("POST",url,reader)
	if err!=nil{
		Logger(err.Error())
		return nil
	}
	client:=http.Client{}
	resp, err := client.Do(request)
	if err!=nil{
		Logger(err.Error())
		return nil
	}
	respBytes,err:=ioutil.ReadAll(resp.Body)
	if err!=nil{
		Logger(err.Error())
		return nil
	}
	result:=&ResultData{}
	json.Unmarshal(respBytes,result)
	if result.Result==true{
		//二次
		KCheck := &KdcheckResult{}
		json.Unmarshal(result.ResultContent,KCheck)
		Logger("宽带续费资格校验成功")
		return KCheck
	}else{
		respStr := (string)(respBytes)
		Logger(respStr)
	}
	return nil
}
