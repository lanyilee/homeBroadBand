package main

import (
	"core"
	"fmt"
	"strconv"
	"time"
)

func main() {
	//configPath:="./config.conf"
	//config,_:=core.ReadConfig(configPath)
	//mobileData:=&core.MobileData{"18320272979",""}
	//typeResult := mobileData.BroadbandTypeApi(config.QueryBroadbandTypeUrl)
	//result:= typeResult.KdcheckrenewalsApi(config.QueryKdcheckrenewalsUrl)
	//fmt.Println(result.BroadSpeed)

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

	timer := time.NewTicker(time.Minute * 1)
	for {
		select {
		case <-timer.C:
			Timerwork()
		}
	}

}

func Timerwork() {
	for i := 1; i < 6; i++ {
		go func(i int) {
			defer func() {
				err := recover()
				if err != nil {
					fmt.Println("error ")
				}
			}()
			//先检查当前日期是否已经处理过业务
			csvUtil := &core.CsvUtil{}
			dateStr := time.Now().Format("20060102") + strconv.Itoa(i)
			b, err := csvUtil.IsExist(dateStr)
			if err != nil {
				core.Logger("csv error")
				return
			}
			if b {
				return
			}
			//获取配置信息
			configPath := "./config.conf"
			config, _ := core.ReadConfig(configPath)
			//下载文件
			filePath := core.FtpGetFile(&config, dateStr)
			fmt.Println(filePath)
			//解析文件
			data, err := core.AnalysisText(filePath)
			if err != nil {
				core.Logger("analysis dataFile error ")
				return
			}
			//API
			for _, number := range data {
				fmt.Println(number)
			}
			//压缩文件
			//加密文件
			//上传文件

			//最后所有操作成功后将文件日期名记录
			csvUtil.Put(dateStr)
			//调用通知接口
		}(i)
	}

}
