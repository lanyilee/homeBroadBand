package core

import (
	"bytes"
	"crypto/cipher"
	"crypto/des"
	"os/exec"
)

type CbcDesEncrypt struct {
}

//填充,刚好8的倍数，就8个8填充
func (cbc *CbcDesEncrypt) Padding(src []byte, blockSize int) []byte {
	padNum := blockSize - len(src)%blockSize
	pad := bytes.Repeat([]byte{byte(padNum)}, padNum)
	return append(src, pad...)
}

//去掉填充，根据最后一位的数值，来决定去掉多少个相应的字节
func (cbc *CbcDesEncrypt) UnPadding(src []byte) []byte {
	length := len(src)
	padNum := int(src[length-1])
	return src[:length-padNum]
}

func (cbc *CbcDesEncrypt) Encrypt3DES(src []byte, key []byte) []byte {
	block, _ := des.NewTripleDESCipher(key)
	src = cbc.Padding(src, block.BlockSize())
	blockmode := cipher.NewCBCEncrypter(block, key[:block.BlockSize()])
	blockmode.CryptBlocks(src, src)
	return src
}

func (cbc *CbcDesEncrypt) Decrypt3DES(src []byte, key []byte) []byte {
	block, _ := des.NewTripleDESCipher(key)
	blockmode := cipher.NewCBCDecrypter(block, key[:block.BlockSize()])
	blockmode.CryptBlocks(src, src)
	src = cbc.UnPadding(src)
	return src
}
