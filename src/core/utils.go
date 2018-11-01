package core

import (
	"gopkg.in/ini.v1"
	"os"
	"strings"
	"time"
)

type Config struct {   //配置文件要通过tag来指定配置文件中的名称
	QueryBroadbandTypeUrl string  `ini:"QueryBroadbandTypeUrl"`
	QueryKdcheckrenewalsUrl string  `ini:"QueryKdcheckrenewalsUrl"`
}

func Logger(strContent string){
	logPath:="./log/"+time.Now().Format("2006-01-02")+".txt"
	file,_:=os.OpenFile(logPath,os.O_RDWR|os.O_CREATE|os.O_APPEND,0777)
	fileTime:=time.Now().Format("2006-01-02 15:04:05")
	fileContent:=strings.Join([]string{"======",fileTime,"=====",strContent,"\n"},"")
	buf:=[]byte(fileContent)
	file.Write(buf)
	defer file.Close()
}


//读取配置文件并转成结构体
func ReadConfig(path string) (Config, error) {
	var config Config
	conf, err := ini.Load(path) //加载配置文件
	if err != nil {
		Logger("load config file fail!")
		return config, err
	}
	conf.BlockMode = false
	err = conf.MapTo(&config) //解析成结构体
	if err != nil {
		Logger("mapto config file fail!")
		return config, err
	}
	return config, nil
}

func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

//获取目录
func getDir(path string) string {
	return subString(path, 0, strings.LastIndex(path, "/"))
}

//截取字符串
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