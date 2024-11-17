package config

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"login/exception"
	"os"
)

// AppConfig 静态全局变量载入
var AppConfig Config

// LoadConfig 载入yaml文件中的参数.
//
// Parameters:
//   - filename: 文件名或路径
//
// Returns:
//   - error: 如果出错，返回错误信息
func LoadConfig(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		exception.PrintError(LoadConfig, err)
		return fmt.Errorf("error opening config file: %v", err)
	}
	// 关闭文件
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			exception.PrintError(LoadConfig, err)
			fmt.Println("error closing config file: ", err)
		}
	}(file)

	decoder := yaml.NewDecoder(file)
	err = decoder.Decode(&AppConfig)
	if err != nil {
		exception.PrintError(LoadConfig, err)
		return fmt.Errorf("error decoding config file: %v", err)
	}

	return nil
}
