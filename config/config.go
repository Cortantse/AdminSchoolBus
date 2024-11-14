package config

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
)

// DatabaseConfig 定义从config.yaml中要提取的结构体
type DatabaseConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
}

type DatabaseNames struct {
	AdminDB     string `yaml:"admin_db"`
	PassengerDB string `yaml:"passenger_db"`
	DriverDB    string `yaml:"driver_db"`
}

type server struct {
	Port string `yaml:"port"`
}

type Config struct {
	Database DatabaseConfig `yaml:"database_connection"`
	Server   server         `yaml:"server"`
	DBNames  DatabaseNames  `yaml:"database_names"`
}

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
		return fmt.Errorf("error opening config file: %v", err)
	}
	// 关闭文件
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			fmt.Println("error closing config file: ", err)
		}
	}(file)

	decoder := yaml.NewDecoder(file)
	err = decoder.Decode(&AppConfig)
	if err != nil {
		return fmt.Errorf("error decoding config file: %v", err)
	}

	return nil
}
