package core

import (
	"bytes"
	"crypto/tls"
	"encoding/xml"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"
)

//账单通知接口

type ZDNotice struct {
	FilePath string //文件相对路径
	PhoneSum string //用户数
}

type EMAIL struct {
	Head HEAD `xml:"HEAD"`
	Body BODY `xml:"BODY"`
}

type HEAD struct {
	VERSION   string `xml:"VERSION"`   //版本，默认0200
	PROVINCE  string `xml:"PROVINCE"`  //gd
	COMEFROM  string `xml:"COMEFROM"`  //省份编码加渠道来源
	COMMANDID string `xml:"COMMANDID"` //CMD00003
	SKEY      string `xml:"SKEY"`      // SKEY=Md5(COMEFROM + COMMANDID + TIMESTAMP + BILLINGFILE + KEY) ，key为body对应账号的密码
	REQSN     string `xml:"REQSN"`     //客户端请求流水号,确保唯一性，用时间
	REQTIME   string `xml:"REQTIME"`   //接口请求时间，YYYYMMDD HH24:MI:SS
}
type BODY struct {
	BILLINGFILE string `xml:"BILLINGFILE"` //文件的相对路径,MD5校验码,解密密码,用户数
	BUSICODE    string `xml:"BUSICODE"`    //账号
}

type REEAMIL struct {
	head REHEAD `xml:"HEAD"`
	body string `xml:"BODY"`
}
type REHEAD struct {
	VERSION   string `xml:"VERSION"`   //版本，默认0200
	PROVINCE  string `xml:"PROVINCE"`  //gd
	COMEFROM  string `xml:"COMEFROM"`  //省份编码加渠道来源
	COMMANDID string `xml:"COMMANDID"` //CMD00003
	SKEY      string `xml:"SKEY"`      // SKEY=Md5(COMEFROM + COMMANDID + TIMESTAMP + BILLINGFILE + KEY) ，key为body对应账号的密码
	REQSN     string `xml:"REQSN"`     //客户端请求流水号,确保唯一性，用时间
	RSPSN     string `xml:"RSPSN"`     //服务端流水号
	REQTIME   string `xml:"REQTIME"`   //接口请求时间，YYYYMMDD HH24:MI:SS
	RSPTIME   string `xml:"RSPTIME"`   //接口应答时间，YYYYMMDD HH24:MI:SS
	RETCODE   string `xml:"RETCODE"`   //统一返回码
	RETDESC   string `xml:"RETDESC"`   //返回码对应描述
}

func (notice *ZDNotice) ZDNoticeApi(config *Config) error {
	xmlTop := `<?xml version="1.0" encoding="GBK"?>` + "\n"
	xmlTopData := []byte(xmlTop)
	email := EMAIL{}
	//head
	head := HEAD{}
	head.VERSION = "0200"
	head.PROVINCE = "gd"
	head.COMEFROM = config.Comefrom
	head.COMMANDID = "CMD00003"
	head.REQSN = time.Now().Format("20060102150405")
	timeSt := time.Now().Format("20060102 15:04:05")
	head.REQTIME = timeSt
	timeUnix := strconv.FormatInt(time.Now().Unix(), 10)
	//文件的相对路径,MD5校验码,解密密码,用户数
	billingFile := notice.FilePath + "," + config.Md5 + "," + config.DesKey + "," + notice.PhoneSum
	//SKEY=Md5(COMEFROM + COMMANDID + TIMESTAMP + BILLINGFILE + KEY) ，key为body对应账号的密码
	skeyStr := head.COMEFROM + head.COMMANDID + timeUnix + billingFile + config.ZdNoticePass
	//skeyData:=[]byte(skeyStr)
	skeyHas := MD532(skeyStr)
	head.SKEY = strings.ToUpper(string(skeyHas[:]))
	email.Head = head
	//body
	email.Body = BODY{BILLINGFILE: billingFile, BUSICODE: config.ZdNoticeUser}
	xmlData, err := xml.MarshalIndent(email, "  ", "    ")
	if err != nil {
		Logger("生成xml失败")
		return err
	}
	//http
	xmlData = append(xmlTopData, xmlData...)
	xmlDataLog := string(xmlData)
	reader := bytes.NewReader(xmlData)
	request, err := http.NewRequest("POST", config.ZdNoticeUrl, reader)
	if err != nil {
		Logger("通知接口请求失败")
		Logger("请求报文：" + xmlDataLog)
		return err
	}

	//省略校验https证书
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	clent := &http.Client{Transport: tr}
	clent.Timeout = time.Minute * 2
	resp, err := clent.Do(request)
	if err != nil {
		Logger("通知接口请求失败")
		Logger("请求报文：" + xmlDataLog)
		return err
	}

	respBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		Logger("通知接口请求失败")
		Logger("请求报文：" + xmlDataLog)
		return err
	}
	//读取xml流

	reEmail := &REEAMIL{}
	//gbk to uft，先把编码转成utf-8，还要把xml文本的gbk替换成UTF-8，xml.Unmarshal才不会出错
	decoStr := Decode(string(respBytes))
	decoStr = strings.Replace(decoStr, "gbk", "UTF-8", -1)
	respBytes = []byte(decoStr)
	err = xml.Unmarshal(respBytes, reEmail)
	if err != nil {
		Logger("通知接口xml数据转化出错")
		return err
	}
	if reEmail.head.RETCODE == "200" {
		Logger("调用通知接口成功")
		return nil
	} else {
		Logger("调用通知接口失败,返回码：" + reEmail.head.RETCODE + ",返回信息：" + reEmail.head.RETDESC)
		Logger("请求报文：" + xmlDataLog)
		Logger("返回报文：" + string(respBytes))
	}
	return nil
}
