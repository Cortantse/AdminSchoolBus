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
