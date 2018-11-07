// FTP Client for Google Go language
// cmd ftp命令
// upload or download has two connection,it has the same host and diferrent port
package core

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
)

type FTP struct {
	host         string
	port         int
	user         string
	passwd       string
	pasv         int
	cmd          string
	Code         int
	Message      string
	Debug        bool
	stream       []byte
	downloadpath string
	conn         net.Conn
	Error        error
}

func (ftp *FTP) debugInfo(s string) {
	if ftp.Debug {
		fmt.Println(s)
	}
}

func (ftp *FTP) Connect(host string, port int) {
	addr := fmt.Sprintf("%s:%d", host, port)
	ftp.conn, ftp.Error = net.Dial("tcp", addr)
	ftp.Response()
	ftp.host = host
	ftp.port = port
}

func (ftp *FTP) Login(user, passwd string) {
	ftp.Request("USER " + user)
	ftp.Request("PASS " + passwd)
	ftp.user = user
	ftp.passwd = passwd
}

func (ftp *FTP) Response() (code int, message string) {
	ret := make([]byte, 1024)
	n, _ := ftp.conn.Read(ret)
	msg := string(ret[:n])
	Logger(msg)
	code, _ = strconv.Atoi(msg[:3])
	message = msg[4 : len(msg)-2]
	ftp.debugInfo("<*cmd*> " + ftp.cmd)
	ftp.debugInfo(fmt.Sprintf("<*code*> %d", code))
	ftp.debugInfo("<*message*> " + message)
	return
}

func (ftp *FTP) Request(cmd string) {
	ftp.conn.Write([]byte(cmd + "\r\n"))
	ftp.cmd = cmd
	ftp.Code, ftp.Message = ftp.Response()
	if ftp.Code == 550 {
		Logger("the file can not be found :" + cmd)
		return
	}
	if ftp.Code == 221 || ftp.Code == 0 {
		Logger("Quit Ftp " + ftp.Message)
		return
	}
	if cmd == "PASV" {
		start, end := strings.Index(ftp.Message, "("), strings.Index(ftp.Message, ")")
		s := strings.Split(ftp.Message[start:end], ",")
		l1, _ := strconv.Atoi(s[len(s)-2])
		l2, _ := strconv.Atoi(s[len(s)-1])
		ftp.pasv = l1*256 + l2
	}
	if (cmd != "PASV") && (ftp.pasv > 0) {
		if strings.Contains(cmd, "RETR") {
			ftp.Message = newRequest(ftp.host, ftp.pasv, nil, ftp.downloadpath)
		} else {
			ftp.Message = newRequest(ftp.host, ftp.pasv, ftp.stream, "")
		}
		//ftp.Message = newRequest(ftp.host, ftp.pasv, ftp.stream)
		ftp.debugInfo("<*response*> " + ftp.Message)
		ftp.pasv = 0
		ftp.stream = nil
		ftp.Code, _ = ftp.Response()
	}
}

func (ftp *FTP) Pasv() {
	ftp.Request("PASV")
}

func (ftp *FTP) Port() {
	ftp.Request("PORT 192.168.22.122.14.178")
}

func (ftp *FTP) Pwd() {
	ftp.Request("PWD")
}

func (ftp *FTP) Cwd(path string) {
	ftp.Request("CWD " + path)
}

func (ftp *FTP) Help() {
	ftp.Request("HELP")
}

func (ftp *FTP) Mkd(path string) {
	ftp.Request("MKD " + path)
}

func (ftp *FTP) Size(path string) (size int) {
	ftp.Request("SIZE " + path)
	size, _ = strconv.Atoi(ftp.Message)
	return
}

func (ftp *FTP) List() {
	ftp.Pasv()
	ftp.Request("LIST")
}

//upload pasv
func (ftp *FTP) Stor(file string, data []byte) {
	ftp.Pasv()
	if data != nil {
		ftp.stream = data
	}
	ftp.Request("STOR " + file)
}

//download RETR pasv
func (ftp *FTP) RETR(localFile string, remoteFile string) {
	cmd := "RETR " + localFile + " " + remoteFile
	ftp.Request(cmd)
}

// upload PORT 主动模式,查了下help并没有get这个命令，不知道怎么解决
func (ftp *FTP) Put(remoteFile string, pathName string) {
	cmd := "PUT " + remoteFile + " " + pathName
	ftp.Request(cmd)
}

// download PORT 主动模式
func (ftp *FTP) Get(remotefile string, Pathname string) {
	cmd := "GET " + remotefile + " " + Pathname
	ftp.Request(cmd)
}

func (ftp *FTP) Quit() {
	ftp.Request("QUIT")
	ftp.conn.Close()
}

// new connect to FTP pasv port, return data
func newRequest(host string, port int, b []byte, downloadpath string) string {
	conn, _ := net.Dial("tcp", fmt.Sprintf("%s:%d", host, port))
	defer conn.Close()
	if b != nil {
		conn.Write(b)
		return "OK"
	}
	if downloadpath != "" {
		file, err := os.Create(downloadpath)
		defer file.Close()
		if err != nil {
			Logger("create file error:" + downloadpath)
			log.Panic(err)
		}
		io.Copy(file, conn)
		return "OK"
	}
	ret := make([]byte, 4096)
	n, _ := conn.Read(ret)
	return string(ret[:n])
}

// download file ,and return the filePath
//func FtpGetFile(config *Config, dateStr string) string {
//	//访问ftp服务器
//	ftp := new(FTP)
//	// debug default false
//	ftp.Debug = true
//	ftp.Connect(config.FromFtpHost, config.FromFtpPort)
//	// login
//	ftp.Login(config.FromFtpLoginUser, config.FromFtpLoginPassword)
//	if ftp.Code == 530 {
//		fmt.Println("error: login failure")
//		os.Exit(-1)
//	}
//	// download
//	remoteFile := "JKGD" + dateStr + ".txt"
//	downloadPath := "./files/" + remoteFile
//	ftp.RETR(remoteFile, downloadPath)
//	ftp.Quit()
//	return downloadPath
//}
