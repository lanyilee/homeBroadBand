package main

import (
	"core"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

var config core.Config
var mutex sync.Mutex

// runtime.NumCPU() 逻辑CPU个数
// runtime.GOMAXPROCS设置机器能够参与执行的CPU的个数
// init()方法会在main函数之前执行
func init() {
	core.Logger("init()" + strconv.Itoa(runtime.NumCPU()))
	runtime.GOMAXPROCS(runtime.NumCPU())
}

func main() {

	//获取配置信息
	configPath := "./config.conf"
	config, _ = core.ReadConfig(configPath)

	//对于固定时间的定时器，可以用sleep，到了时间才启动
	fixTime, err := core.GetFixTime(&config)
	if err != nil {
		core.Logger("获取定时器固定时间出错！")
		return
	}
	for fixTime.After(time.Now()) {
		time.Sleep(time.Second * 1)
	}
	fmt.Println("start now" + time.Now().Format("2006-01-02 15:04:05"))

	//KdAccount:="18320272979"
	//phoneNum:=subString(KdAccount,0,3)+"****"+subString(KdAccount,7,11)
	//fmt.Println(phoneNum)

	//调用通知接口
	//baseGzPath:="JKGD20181215.txt.gz"
	//baseDesPath := baseGzPath + ".des"
	//notice := &core.ZDNotice{}
	////toftpPath := string([]byte(config.ToFtpPath)[1:])
	//notice.FilePath = "./formatFiles/" + baseDesPath
	//notice.FtpsFilePath = "ftps://" + config.ToFtpHost + "/" + baseDesPath
	//notice.PhoneSum ="12"
	//err = notice.ZDNoticeApi(&config)
	//if err != nil {
	//	fmt.Println(err)
	//	core.Logger("调用账单通知接口出错")
	//	return
	//}

	//启动先扫描一次
	Timerwork()
	timer := time.NewTicker(time.Minute*30)
	for {
		select {
		case <-timer.C:
			//设置时间
			local,_:=time.LoadLocation("Local")
			nowTime:=time.Now()
			toFixTimeStr:=nowTime.Format("2006")+"-"+nowTime.Format("01") +"-"+nowTime.Format("02")+" "+fixTime.Format("15")+":"+fixTime.Format("04")+":"+fixTime.Format("05")
			toFixTime,_:=time.ParseInLocation("2006-01-02 15:04:05",toFixTimeStr,local)
			for toFixTime.After(time.Now()){
				time.Sleep(time.Second * 1)
			}
			Timerwork()
		}
	}

}

func subString(str string, start, end int) string {
	rs := []rune(str)
	length := len(rs)
	if start < 0 || start > length {
		panic("start is wrong")
	}
	if end < start || end > length {
		panic("end is wrong")
	}
	return string(rs[start:end])
}

func Timerwork() {

	csvUtil := &core.CsvUtil{}
	dateStr := time.Now().Format("20060102")
	fmt.Println("dateStr" + dateStr)
	b, err := csvUtil.IsExist(dateStr)
	if err != nil {
		core.Logger("csv error")
		return
	}
	if b {
		return
	}

	//下载文件
	filePath, err := core.FtpGetFile(&config, dateStr)
	if err != nil {
		core.Logger("get ftp files error")
		return
	}
	core.Logger("download file success :" + filePath)

	//解析文件
	//filePath := "./files/xaa.txt"
	data, err := core.AnalysisText(filePath)
	if err != nil {
		core.Logger("analysis dataFile error ")
		return
	}
	core.Logger("analysis dataFile success :" + filePath)

	//API,并发8000个
	var quit chan int
	jkData := &([]core.KdcheckResult{})
	quit = make(chan int)
	concurrencyNum := 8000 //并发数
	if len(data) < concurrencyNum {
		for _, number := range data {
			mobileData := &core.MobileData{number, ""}
			typeResult, err := mobileData.BroadbandTypeApi(config.QueryBroadbandTypeUrl)
			if err != nil || typeResult == nil {
				core.SyncLoggerNum(mobileData.Mobile)
				continue
			} else {
				result, err := typeResult.KdcheckrenewalsApi(config.QueryKdcheckrenewalsUrl)
				if err != nil || result == nil {
					core.SyncLoggerNum(mobileData.Mobile)
					continue
				} else {
					*jkData = append(*jkData, *result)
				}
			}
		}
	} else {
		interval := len(data) / concurrencyNum //每个并发线程所处理的数据量
		lastNum := len(data) - interval*concurrencyNum
		//如果最后一条线程处理数量远大于前面线程平均的数量，就要根据平均数增加线程
		if lastNum >= interval {
			addNum := lastNum / interval
			concurrencyNum = concurrencyNum + addNum
		}
		fmt.Println("总并发数：" + strconv.Itoa(concurrencyNum))
		for i := 0; i < concurrencyNum; i++ {
			start := interval * i
			end := interval * (i + 1)
			if i == concurrencyNum-1 {
				start = interval * i
				end = len(data) - 1
			}
			data2 := data[start:end]
			go JKApi(data2, jkData, quit)
		}
		//信道出去
		for i := 0; i < concurrencyNum; i++ {
			<-quit
			core.Logger("第" + strconv.Itoa(i) + "信道out")
		}
	}
	core.Logger(filePath + " 为宽带用户且返回成功总数：" + strconv.Itoa(len(*jkData)))

	//模板并存储
	formatFilePath := "./formatFiles/JKGD" + dateStr + ".txt"
	formatFile, err := os.OpenFile(formatFilePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0777)
	for _, data := range *jkData {
		fileContent := core.FormatJKText(&data)
		buf := []byte(fileContent)
		formatFile.Write(buf)
	}
	formatFile.Close()

	//压缩文件
	err = core.CompressFile(formatFilePath)
	if err != nil {
		core.Logger("压缩文件出错")
		return
	}
	core.Logger("压缩文件成功")

	//加密文件
	baseGzPath := "JKGD" + dateStr + ".txt.gz"
	desPath, err := core.Encrypt3DESByOpenssl(config.DesKey, baseGzPath)
	if err != nil {
		core.Logger("加密文件出错")
		return
	}
	core.Logger("加密文件成功:" + desPath)

	//上传文件
	err = core.FtpsPutFile(&config, baseGzPath+".des")
	if err != nil {
		core.Logger("上传文件出错")
		return
	}
	core.Logger("上传文件成功")

	//调用通知接口
	baseDesPath := baseGzPath + ".des"
	notice := &core.ZDNotice{}
	//toftpPath := string([]byte(config.ToFtpPath)[1:])
	notice.FilePath = "./formatFiles/" + baseDesPath
	notice.FtpsFilePath = "ftps://" + config.ToFtpHost + "/" + baseDesPath
	notice.PhoneSum = strconv.Itoa(len(*jkData))
	err = notice.ZDNoticeApi(&config)
	if err != nil {
		fmt.Println(err)
		core.Logger("调用账单通知接口出错")
		return
	}
	core.Logger("调用账单通知接口成功")
	//最后所有操作成功后将文件日期名记录
	csvUtil.Put(dateStr)
}

//同步调用家宽api
func JKApi(data []string, jkData *[]core.KdcheckResult, quit chan int) {
	defer func() {
		quit <- 1 //finished
		recover()
	}()
	threadStr := "线程" + strconv.Itoa(GoID()) + ",总数" + strconv.Itoa(len(data)) + ";"
	//fmt.Println(threadStr)
	core.Logger(threadStr)
	for _, number := range data {
		mobileData := &core.MobileData{number, ""}
		typeResult, err := mobileData.BroadbandTypeApi(config.QueryBroadbandTypeUrl)
		if err != nil || typeResult == nil {
			core.SyncLoggerNum(mobileData.Mobile)
			continue
		} else {
			result, err := typeResult.KdcheckrenewalsApi(config.QueryKdcheckrenewalsUrl)
			if err != nil || result == nil {
				core.SyncLoggerNum(mobileData.Mobile)
				continue
			} else {
				//fmt.Println(number)
				mutex.Lock() //上锁，上锁后，被锁定的内容不会被两个或者多个线程同时竞争
				*jkData = append(*jkData, *result)
				mutex.Unlock()
			}
		}
	}
	threadStr2 := "线程：" + strconv.Itoa(GoID()) + " over;"
	//fmt.Println(threadStr2)
	core.Logger(threadStr2)
	//core.Logger("")

}

//获取线程ID
func GoID() int {
	var buf [64]byte
	n := runtime.Stack(buf[:], false)
	idField := strings.Fields(strings.TrimPrefix(string(buf[:n]), "goroutine "))[0]
	GoroutineId, err := strconv.Atoi(idField)
	if err != nil {
		panic(fmt.Sprintf("cannot get goroutine id: %v", err))
	}
	return GoroutineId
}
