package main

import (
	"archive/zip"
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

	//启动先扫描一次
	Timerwork()
	timer := time.NewTicker(time.Hour * 10)
	for {
		select {
		case <-timer.C:
			Timerwork()
		}
	}

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
	//filePath := "./files/JKGD20181109.txt"
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
		//JKApi(data, jkData, quit)
	} else {
		interval := len(data) / concurrencyNum //每个并发线程所处理的数据量
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
	formatFilePath := "./formatFiles/" + dateStr + ".txt"
	formatFile, err := os.OpenFile(formatFilePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0777)
	for _, data := range *jkData {
		fileContent := core.FormatJKText(&data)
		buf := []byte(fileContent)
		formatFile.Write(buf)
	}
	formatFile.Close()
	//压缩文件
	baseZipPath := dateStr + ".zip"
	logzip, _ := os.Create("./formatFiles/" + baseZipPath)
	defer logzip.Close()
	w := zip.NewWriter(logzip)
	defer w.Close()
	formatFile, _ = os.OpenFile(formatFilePath, os.O_RDWR, 0777)
	core.CompressSingle(formatFile, "", w)
	core.Logger("压缩文件成功")
	//加密文件
	desPath, err := core.Encrypt3DESByOpenssl(config.DesKey, baseZipPath)
	if err != nil {
		core.Logger("加密文件出错")
		return
	}
	core.Logger("加密文件成功:" + desPath)

	//上传文件
	err = core.FtpPutFile(&config, baseZipPath+".des")
	if err != nil {
		core.Logger("上传文件出错")
		return
	}
	core.Logger("上传文件成功")
	//调用通知接口
	baseDesPath := baseZipPath + ".des"
	notice := &core.ZDNotice{}
	notice.FilePath = "./formatFiles/" + baseDesPath
	//notice.PhoneSum = strconv.Itoa(len(*jkData))
	notice.PhoneSum = "8000"
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
func JKApi(data []string, jkData *([]core.KdcheckResult), quit chan int) {
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
				//mutex.Lock() //上锁，上锁后，被锁定的内容不会被两个或者多个线程同时竞争
				*jkData = append(*jkData, *result)
				//mutex.Unlock()
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
