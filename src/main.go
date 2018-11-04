package main

import (
	"core"
	"fmt"
	"os"
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

	timer := time.NewTicker(time.Second * 10)
	for {
		select {
		case <-timer.C:
			Timerwork()
		}
	}

}

func Timerwork() {
	go func() {
		//访问ftp服务器
		ftp := new(core.FTP)
		// debug default false
		ftp.Debug = true
		ftp.Connect("192.168.1.100", 21)

		// login
		ftp.Login("Temp", "123")
		if ftp.Code == 530 {
			fmt.Println("error: login failure")
			os.Exit(-1)
		}

		//
		ftp.RETR("test2.txt", "./log/download.txt")
		fmt.Println("code:", ftp.Code, ", message:", ftp.Message)

		ftp.Quit()
	}()

}
