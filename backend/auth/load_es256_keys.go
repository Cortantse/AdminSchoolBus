package auth

import (
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
)

func loadPrivateKey(filename string) (*ecdsa.PrivateKey, error) {
	// 读取文件内容
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			fmt.Println("Failed to close file:", err)
		}
	}(file)

	// 读取文件内容到字节数组
	fileInfo, err := file.Stat()
	if err != nil {
		return nil, err
	}
	fileSize := fileInfo.Size()
	pemBytes := make([]byte, fileSize)
	_, err = file.Read(pemBytes)
	if err != nil {
		return nil, err
	}

	// 解析 PEM 格式的私钥
	block, _ := pem.Decode(pemBytes)
	if block == nil {
		return nil, fmt.Errorf("failed to parse PEM block containing the private key")
	}

	// 解析 DER 格式的私钥
	privateKey, err := x509.ParseECPrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	return privateKey, nil
}

func loadPublicKey(filename string) (*ecdsa.PublicKey, error) {
	// 读取文件内容
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			fmt.Println("Failed to close file:", err)
		}
	}(file)

	// 读取文件内容到字节数组
	fileInfo, err := file.Stat()
	if err != nil {
		return nil, err
	}
	fileSize := fileInfo.Size()
	pemBytes := make([]byte, fileSize)
	_, err = file.Read(pemBytes)
	if err != nil {
		return nil, err
	}

	// 解析 PEM 格式的公钥
	block, _ := pem.Decode(pemBytes)
	if block == nil {
		return nil, fmt.Errorf("failed to parse PEM block containing the public key")
	}

	// 解析 DER 格式的公钥
	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	// 类型断言，确保解析出来的是 ECDSA 公钥
	publicKey, ok := pub.(*ecdsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("not an ECDSA public key")
	}

	return publicKey, nil
}

// 载入keys
func loadKeys() error {
	// 读取私钥文件
	var err error
	priKey, err = loadPrivateKey("private.pem")
	if err != nil {
		fmt.Println("Failed to load private key:", err)
		return err
	}

	// 读取公钥文件
	pubKey, err = loadPublicKey("public.pem")
	if err != nil {
		fmt.Println("Failed to load public key:", err)
		return err
	}

	return nil
}
