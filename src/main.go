package main

import (
	"core"
	"fmt"
	"runtime"
	"strconv"
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
	//logzip,_:=os.Create("log/123.zip")
	//defer logzip.Close()
	//w:=zip.NewWriter(logzip)
	//defer w.Close()
	//logtxt,_:=os.OpenFile("log/123.txt",os.O_RDWR,0777)
	//core.CompressSingle(logtxt,"",w)

	//加密
	//err := core.Encrypt3DESByOpenssl("12345678abcdefgh87654321", "log/123.zip")
	//if err != nil {
	//	log.Panic(err)
	//}

	//解密
	//err:=core.Zip3DESDEncrypt("log/test.zip.des","12345678abcdefgh87654321",&core.CbcDesEncrypt{})
	//if err!=nil{
	//	log.Panic(err)
	//}

	//获取配置信息
	configPath := "./config.conf"
	config, _ = core.ReadConfig(configPath)

	//dateStr := time.Now().Format("20060102") + strconv.Itoa(1)
	//core.FtpGetFile(&config,dateStr)
	//core.AnalysisText("./files/JKGD201811071.txt")

	//对于固定时间的定时器，可以用sleep，到了时间才启动
	//fixTime,err:= core.GetFixTime(&config)
	//if err!=nil{
	//	core.Logger("获取定时器固定时间出错！")
	//	return
	//}
	//for fixTime.After(time.Now()) {
	//	time.Sleep(time.Second * 1)
	//}
	//fmt.Println("start now" + time.Now().Format("2006-01-02 15:04:05"))

	//core.FtpGetFile(&config,"")

	//api测试
	//mobileData:=&core.MobileData{"18320272979",""}
	//typeResult,err:= mobileData.BroadbandTypeApi(config.QueryBroadbandTypeUrl)
	//result,err:= typeResult.KdcheckrenewalsApi(config.QueryKdcheckrenewalsUrl)
	//fmt.Println(result.KdAccount)

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
	for a := 1; a < 5; a++ {
		go func(a int) {
			defer func() {
				recover()
			}()
			//先检查当前日期是否已经处理过业务
			csvUtil := &core.CsvUtil{}
			dateStr := time.Now().Format("20060102") + strconv.Itoa(a)
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
			//filePath, err := core.FtpGetFile(&config, dateStr)
			//if err != nil {
			//	core.Logger("get ftp files error")
			//	return
			//}
			filePath := "./files/JKGD20181107" + strconv.Itoa(a) + ".txt"
			core.Logger("download file success :" + filePath)
			//解析文件
			data, err := core.AnalysisText(filePath)
			if err != nil {
				core.Logger("analysis dataFile error ")
				return
			}
			core.Logger("analysis dataFile success :" + filePath)
			//API,并发100个
			var quit chan int
			quit = make(chan int)
			loopNum := 1000
			if len(data) < 1000 {
				loopNum = len(data)
			}
			interval := len(data) / 1000
			fmt.Println("总数：" + strconv.Itoa(loopNum))
			jkData := &([]core.KdcheckResult{})
			//jkDataCh:=make(chan []core.KdcheckResult)
			for i := 0; i < loopNum; i++ {
				start := interval * i
				end := interval*(i+1) - 1
				if i == loopNum-1 {
					start = interval * i
					end = len(data) - 1
				}
				data2 := data[start:end]
				go JKApi(data2, jkData, quit)
			}
			//信道出去
			for i := 0; i < loopNum; i++ {
				<-quit
				core.Logger("第" + strconv.Itoa(i) + "信道out")
			}
			core.Logger(filePath + " 为宽带用户且返回成功总数：" + strconv.Itoa(len(*jkData)))
			for _, data := range *jkData {
				core.Logger(data.KdAccount + " is success")
				fmt.Println(data.KdAccount)
			}

			//压缩文件
			//加密文件
			//上传文件

			//最后所有操作成功后将文件日期名记录
			csvUtil.Put(dateStr)
			//调用通知接口
		}(a)
	}

}

//同步调用家宽api
func JKApi(data []string, jkData *([]core.KdcheckResult), quit chan int) {
	defer func() {
		quit <- 1 //finished
		recover()
	}()
	fmt.Println(len(data))
	for _, number := range data {
		//fmt.Println(number)
		mobileData := &core.MobileData{number, ""}
		typeResult, err := mobileData.BroadbandTypeApi(config.QueryBroadbandTypeUrl)
		if err != nil || typeResult == nil {
			//core.Logger("failure phone: " + mobileData.Mobile)
			continue
		} else {
			result, err := typeResult.KdcheckrenewalsApi(config.QueryKdcheckrenewalsUrl)
			if err != nil || result == nil {
				//core.Logger("failure phone: " + mobileData.Mobile)
				continue
			} else {
				mutex.Lock() //上锁，上锁后，被锁定的内容不会被两个或者多个线程同时竞争
				*jkData = append(*jkData, *result)
				//core.Logger(result.KdAccount+" success!")
				mutex.Unlock()
			}
		}
	}

	//core.Logger("")

}
