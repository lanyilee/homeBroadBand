package core

import (
	"bufio"
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"github.com/jlaffaye/ftp"
	"gopkg.in/ini.v1"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	//配置文件要通过tag来指定配置文件中的名称
	//api
	QueryBroadbandTypeUrl   string `ini:"QueryBroadbandTypeUrl"`
	QueryKdcheckrenewalsUrl string `ini:"QueryKdcheckrenewalsUrl"`
	//from-ftp
	FromFtpHost          string `ini:"FromFtpHost"`
	FromFtpLoginUser     string `ini:"FromFtpLoginUser"`
	FromFtpLoginPassword string `ini:"FromFtpLoginPassword"`
	//to-ftp
	ToFtpHost          string `ini:"ToFtpHost"`
	ToFtpLoginUser     string `ini:"ToFtpLoginUser"`
	ToFtpLoginPassword string `ini:"ToFtpLoginPassword"`
	ToFtpPath          string `ini:"ToFtpPath"`
	//MD5
	Md5 string `ini:"Md5"`
	//des
	DesKey string `ini:"DesKey"`
	//fixed-time
	FixedTime string `ini:"FixedTime"`
	//zdapi
	ZdNoticeUrl  string `ini:"ZdNoticeUrl"`
	ZdNoticeUser string `ini:"ZdNoticeUser"`
	ZdNoticePass string `ini:"ZdNoticePass"`
	Comefrom     string `ini:"Comefrom"`
}

func Logger(strContent string) {
	logPath := "./log/" + time.Now().Format("2006-01-02") + ".txt"
	file, _ := os.OpenFile(logPath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0777)
	fileTime := time.Now().Format("2006-01-02 15:04:05")
	fileContent := strings.Join([]string{"===", fileTime, "===", strContent, "\n"}, "")
	buf := []byte(fileContent)
	file.Write(buf)
	defer file.Close()
}

//记录失败的号码
func LoggerFailNum(strContent string) {
	logPath := "./log/" + time.Now().Format("2006-01-02") + "-Num.txt"
	file, _ := os.OpenFile(logPath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0777)
	fileContent := strings.Join([]string{strContent, "\n"}, "")
	buf := []byte(fileContent)
	file.Write(buf)
	defer file.Close()
}

func SyncLoggerNum(strContent string) {
	go func(str string) {
		defer func() {
			recover()
		}()
		LoggerFailNum(str)
	}(strContent)
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

//生成32位MD5
func MD532(text string) string {
	ctx := md5.New()
	ctx.Write([]byte(text))
	return hex.EncodeToString(ctx.Sum(nil))
}

//用linux自带的openssl加密3DES-CBC,command的首参是openssl,不是平常的/bin/bash
func Encrypt3DESByOpenssl(key string, fileName string) (desPath string, err error) {
	filePath := "./formatFiles/" + fileName
	desPath = filePath + ".des"
	cmd := exec.Command("openssl", "enc", "-des-ede3-cbc", "-e", "-k", key, "-in", filePath, "-out", desPath)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		Logger("Error:can not obtain stdout pipe for command")
		return "", err
	}
	//执行命令
	if err := cmd.Start(); err != nil {
		Logger("Error:The command is err")
		return "", err
	}
	//读取所有输出
	_, err = ioutil.ReadAll(stdout)
	if err != nil {
		return "", err
	}
	if err := cmd.Wait(); err != nil {
		Logger("wait error")
		return "", err
	}
	Logger("encrypt success:")
	//fmt.Printf("stdout:\n\n %s", "")
	return desPath, nil
}

//解析文本
func AnalysisText(filePath string) (numbers []string, err error) {
	file, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0777)
	info, err := os.Stat(filePath)
	Logger(filePath + " file size is " + strconv.FormatInt(info.Size(), 10))
	defer file.Close()
	if err != nil {
		Logger("open file error :" + filePath)
		return nil, err
	}
	//buf := make([]byte, 12)
	bfrd := bufio.NewReader(file)
	for {
		line, err := bfrd.ReadBytes('\n')
		var number string
		length := len(line)
		if length == 12 {
			number = string(line[:len(line)-1]) //拿到的buf[:n]是"13432655213\n"这样的数据，所以要减-1，即11
			numbers = append(numbers, number)
		} else {
			number = string(line)
			Logger(number)
		}

		if err != nil { //遇到任何错误立即返回，并忽略 EOF 错误信息
			if err == io.EOF {
				Logger(filePath + " 文件号码总数：" + strconv.Itoa(len(numbers)))
				return numbers, nil
			}
			Logger("read file error:" + filePath)
			return nil, err
		}
	}
	Logger(filePath + " 文件号码总数：" + strconv.Itoa(len(numbers)))
	return numbers, nil
}

//重构文本-模板
func FormatJKText(kd *KdcheckResult) string {
	str := "START|" + kd.KdAccount + "\n" + "宽带属性|" + kd.KdAccount + "~家庭宽带~" + kd.UserStatus + "~" + kd.IsYearPackAge + "~" + kd.LastDate + "~" + kd.BroadSpeed + "|010000\nEND\n"
	return str
}

//取定时时间
func GetFixTime(config *Config) (fixTime time.Time, err error) {
	fixTimeStr := config.FixedTime
	//fixTime := time.Date(2018, 11, 06, 07, 52, 0, 0, time.Local)
	year, err := strconv.Atoi(subString(fixTimeStr, 0, 4))
	if err != nil {
		return fixTime, err
	}
	monthNum, _ := strconv.Atoi(subString(fixTimeStr, 4, 6))
	if err != nil {
		return fixTime, err
	}
	day, _ := strconv.Atoi(subString(fixTimeStr, 6, 8))
	if err != nil {
		return fixTime, err
	}
	hour, _ := strconv.Atoi(subString(fixTimeStr, 8, 10))
	if err != nil {
		return fixTime, err
	}
	min, _ := strconv.Atoi(subString(fixTimeStr, 10, 12))
	if err != nil {
		return fixTime, err
	}
	//这个month竟然还是个time.Month类型，奇葩
	month := time.Month(monthNum)
	fixTime = time.Date(year, month, day, hour, min, 0, 0, time.Local)
	return fixTime, nil
}

//FTP-Get操作
func FtpGetFile(config *Config, dateStr string) (path string, err error) {
	//访问ftp服务器
	entry, err := ftp.Connect(config.FromFtpHost)
	defer entry.Quit()
	if err != nil {
		Logger("connect to ftp server error :" + config.FromFtpHost)
		return "", err
	}
	Logger("connect to ftp server success :" + config.FromFtpHost)
	//login
	entry.Login(config.FromFtpLoginUser, config.FromFtpLoginPassword)
	if err != nil {
		Logger("ftp login error, user:" + config.FromFtpLoginUser + ";pass: " + config.FromFtpLoginPassword)
		fmt.Println(err)
		return "", err
	}
	Logger("ftp login success")
	//get
	remoteFile := "JKGD" + dateStr + ".txt"
	//remoteFile := "./logfile/10008105/201810/20181008_001.log"
	res, err := entry.Retr(remoteFile)
	if err != nil {
		Logger("get file error :" + remoteFile)
		Logger(err.Error())
		return "", err
	}
	Logger("get file success :" + remoteFile)
	downloadPath := "./files/" + remoteFile
	file, err := os.Create(downloadPath)
	defer file.Close()
	defer res.Close()
	//一次读取多少字节
	buf := make([]byte, 1024)
	for {
		n, err := res.Read(buf)
		file.Write(buf[:n]) //n是成功读取个数
		if err != nil {     //遇到任何错误立即返回，并忽略 EOF 错误信息
			if err == io.EOF {
				return downloadPath, nil
			}
			Logger(err.Error())
			return "", err
		}
	}
	return downloadPath, nil
}

//FTP-Put操作
func FtpPutFile(config *Config, fileName string) error {
	basePath := "./formatFiles/" + fileName
	toPath := config.ToFtpPath + fileName
	entry, err := ftp.Connect(config.ToFtpHost)
	defer entry.Quit()
	if err != nil {
		Logger("connect to ftp server error :" + config.ToFtpHost)
		return err
	}
	Logger("connect to ftp server success :" + config.ToFtpHost)
	//login
	entry.Login(config.ToFtpLoginUser, config.ToFtpLoginPassword)
	if err != nil {
		Logger("ftp login error, user:" + config.ToFtpLoginUser + ";pass: " + config.ToFtpLoginPassword)
		fmt.Println(err)
		return err
	}
	Logger("ftp login success")
	file, err := ioutil.ReadFile(basePath)
	buf := bytes.NewReader(file)
	err = entry.Stor(toPath, buf)
	if err != nil {
		Logger("upload file to ftp server error :" + basePath)
		return err
	}
	return nil
}

//SFTP-PUT 操作
func SFtpPutFile(config *Config, fileName string) error {
	basePath := "./formatFiles/" + fileName
	toPath := config.ToFtpPath + fileName
	//

	entry, err := ftp.Connect(config.ToFtpHost)
	defer entry.Quit()
	if err != nil {
		Logger("connect to ftp server error :" + config.ToFtpHost)
		return err
	}
	Logger("connect to ftp server success :" + config.ToFtpHost)
	//login
	entry.Login(config.ToFtpLoginUser, config.ToFtpLoginPassword)
	if err != nil {
		Logger("ftp login error, user:" + config.ToFtpLoginUser + ";pass: " + config.ToFtpLoginPassword)
		fmt.Println(err)
		return err
	}
	Logger("ftp login success")
	file, err := ioutil.ReadFile(basePath)
	buf := bytes.NewReader(file)
	err = entry.Stor(toPath, buf)
	if err != nil {
		Logger("upload file to ftp server error :" + basePath)
		return err
	}
	return nil
}
