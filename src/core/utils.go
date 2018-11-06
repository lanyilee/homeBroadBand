package core

import (
	"fmt"
	"gopkg.in/ini.v1"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"time"
)

type Config struct { //配置文件要通过tag来指定配置文件中的名称
	//api
	QueryBroadbandTypeUrl   string `ini:"QueryBroadbandTypeUrl"`
	QueryKdcheckrenewalsUrl string `ini:"QueryKdcheckrenewalsUrl"`
	//from-ftp
	FromFtpHost          string `ini:"FromFtpHost"`
	FromFtpPort          int    `ini:"FromFtpPort"`
	FromFtpLoginUser     string `ini:"FromFtpLoginUser"`
	FromFtpLoginPassword string `ini:"FromFtpLoginPassword"`
}

func Logger(strContent string) {
	logPath := "./log/" + time.Now().Format("2006-01-02") + ".txt"
	file, _ := os.OpenFile(logPath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0777)
	fileTime := time.Now().Format("2006-01-02 15:04:05")
	fileContent := strings.Join([]string{"======", fileTime, "=====", strContent, "\n"}, "")
	buf := []byte(fileContent)
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

//截取字符串,截取的不包括第end位
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

//3DES加密
func Zip3DESEncrypt(zipPath string, key string, cbc *CbcDesEncrypt) error {
	logzip, _ := os.OpenFile(zipPath, os.O_RDWR, 0777)
	defer logzip.Close()
	buff, err := ioutil.ReadAll(logzip)
	if err != nil {
		Logger("jiami")
		return err
	}
	keyBytes := []byte(key)
	encryptBuff := cbc.Encrypt3DES(buff, keyBytes)
	logzipdes, err := os.Create(zipPath + ".des")
	if err != nil {
		return err
	}
	defer logzipdes.Close()
	_, err = logzipdes.Write(encryptBuff)
	if err != nil {
		return err
	}
	return nil
}

//3DES解密
func Zip3DESDEncrypt(zipDesPath string, key string, cbc *CbcDesEncrypt) error {
	logzipdes, _ := os.OpenFile(zipDesPath, os.O_RDWR, 0777)
	defer logzipdes.Close()
	buff, err := ioutil.ReadAll(logzipdes)
	if err != nil {
		return err
	}
	keyBytes := []byte(key)
	dencryptBuff := cbc.Decrypt3DES(buff, keyBytes)
	toPath := subString(zipDesPath, 0, len(zipDesPath)-8) + "2.zip"
	logzip2, err := os.Create(toPath)
	if err != nil {
		return err
	}
	defer logzip2.Close()
	_, err = logzip2.Write(dencryptBuff)
	if err != nil {
		return err
	}
	return nil
}

//用linux自带的openssl加密3DES-CBC,command的首参是openssl,不是平常的/bin/bash
func Encrypt3DESByOpenssl(key string, filePath string) error {
	toPath := filePath + ".des"
	cmd := exec.Command("openssl", "enc", "-des-ede3-cbc", "-e", "-k", key, "-in", filePath, "-out", toPath)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		Logger("Error:can not obtain stdout pipe for command")
		return err
	}
	//执行命令
	if err := cmd.Start(); err != nil {
		Logger("Error:The command is err")
		return err
	}
	//读取所有输出
	_, err = ioutil.ReadAll(stdout)
	if err != nil {
		return err
	}
	if err := cmd.Wait(); err != nil {
		fmt.Println("wait:", err.Error())
		return err
	}
	Logger("encrypt success:")
	//fmt.Printf("stdout:\n\n %s", "")
	return nil
}

//解析文本
func AnalysisText(filePath string) (numbers []string, err error) {
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		Logger("open file error :" + filePath)
		return nil, err
	}
	for _, line := range strings.Split(string(data), "\n") {
		numbers = append(numbers, line)
	}
	return numbers, nil
}

//并发调用api
