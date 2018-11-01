package main

import (
	"archive/zip"
	"core"
	"os"
)

func main(){
	//configPath:="./url.conf"
	//config,_:=core.ReadConfig(configPath)
	//mobileData:=&core.MobileData{"18320272979",""}
	//typeResult := mobileData.BroadbandTypeApi(config.QueryBroadbandTypeUrl)
	//result:= typeResult.KdcheckrenewalsApi(config.QueryKdcheckrenewalsUrl)
	//fmt.Println(result.BroadSpeed)
	logzip,_:=os.Create("log/2018-11-01.zip")
	defer logzip.Close()
	w:=zip.NewWriter(logzip)
	defer w.Close()
	log,_:=os.OpenFile("log/2018-10-31.txt",os.O_RDWR,0777)
	core.CompressSingle(log,"",w)

}

