package config

// 导出角色常量，使用大写字母
type Role int

// ** 请在使用角色的时候使用iota模拟枚举类型 **
const (
	RoleAdmin Role = iota
	RolePassenger
	RoleDriver
	Unknown
)

func (r Role) String() string {
	switch r {
	case RoleAdmin:
		return "Admin"
	case RolePassenger:
		return "Passenger"
	case RoleDriver:
		return "Driver"
	default:
		return "Unknown"
	}
}

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

type Server struct {
	Port string `yaml:"port"`
}

type Jwt struct {
	ExpirationHoursPass   int `yaml:"expiration_hours_passenger"`
	ExpirationHoursAdmin  int `yaml:"expiration_hours_admin"`
	ExpirationHoursDriver int `yaml:"expiration_hours_driver"`
}

type Config struct {
	Database DatabaseConfig `yaml:"database_connection"`
	Server   Server         `yaml:"server"`
	DBNames  DatabaseNames  `yaml:"database_names"`
	Jwt      Jwt            `yaml:"jwt"`
}
