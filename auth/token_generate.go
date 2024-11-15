package auth

import (
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"github.com/cristalhq/jwt"
	"login/config"
	"time"
)

var pubKey *ecdsa.PublicKey
var priKey *ecdsa.PrivateKey

// InitTokenService 初始化token服务，加载公私钥
func InitTokenService() error {
	err := loadKeys()
	if err != nil {
		return err
	}
	return nil
}

// generateToken 生成用户的token，内部函数，禁止使用
// 负责生成根据身份信息生成的token，返回token
// Parameters:
//   - role  需要生成的对象
//
// Returns:
//   - string 生成的token的string形式，还未变为raw
//   - error 如果有错误
func generateToken(role config.Role, user_id string) (string, error) {
	signer, err := jwt.NewES256(pubKey, priKey)
	if err != nil {
		return "", err
	}
	builder := jwt.NewTokenBuilder(signer)

	// 根据用户的角色，设置token的过期时间和转换literal值
	var expiry int
	var roleLiteral string
	switch role {
	case config.RoleAdmin:
		expiry = config.AppConfig.Jwt.ExpirationHoursAdmin
		roleLiteral = "admin"
		break
	case config.RolePassenger:
		expiry = config.AppConfig.Jwt.ExpirationHoursPass
		roleLiteral = "passenger"
		break
	case config.RoleDriver:
		expiry = config.AppConfig.Jwt.ExpirationHoursDriver
		roleLiteral = "driver"
		break
	default:
		return "", fmt.Errorf("role choosen is not valid in generateToken")
	}

	// claim仅包含发放对象和截止日期
	claims := &jwt.StandardClaims{
		Audience:  []string{roleLiteral}, // 获得token的对象
		ExpiresAt: jwt.Timestamp((time.Now().Add(time.Hour * time.Duration(expiry))).Unix()),
		Subject:   user_id, //存取获得这个令牌的用户id
	}
	token, _ := builder.Build(claims)

	// 注意需要返回raw作为string！！！！！！！！！
	return string(token.Raw()), nil
}

// 解析和验证 JWT，返回对应错误，验证是否过期，并返回请求者身份
func verifyToken(tokenString string) (config.Role, string, error) {

	// 创建 signer
	signer, err := jwt.NewES256(pubKey, priKey)
	if err != nil {
		return config.Unknown, "", fmt.Errorf("error creating signer: %v", err)
	}

	// 解析并验证 JWT 字符串
	token, err := jwt.ParseAndVerifyString(tokenString, signer)
	if err != nil {
		return config.Unknown, "", jwt.ErrInvalidSignature
	}

	// 提取原始声明数据（rawClaims）
	rawClaims := token.RawClaims()

	// 解码 rawClaims 为标准声明（StandardClaims）
	claims := &jwt.StandardClaims{}
	if err := json.Unmarshal(rawClaims, claims); err != nil {
		return config.Unknown, "", jwt.Error("error decoding claims")
	}

	// 提取角色（audience - aud）字段
	roleLiteral := claims.Audience[0]
	var role config.Role
	switch roleLiteral {
	case "admin":
		role = config.RoleAdmin
		break
	case "passenger":
		role = config.RolePassenger
		break
	case "driver":
		role = config.RoleDriver
		break
	default:
		return config.Unknown, "", fmt.Errorf("unknown role in token: %s", roleLiteral)
	}

	// 判断是否超时
	if claims.IsExpired(time.Now()) {
		return role, roleLiteral, jwt.ErrTokenExpired
	}
	// 获取id
	subject := claims.Subject

	// 返回解析到的角色
	return role, subject, nil
}

func Test() {
	token, err1 := generateToken(config.RoleDriver, "123")

	role, id, err2 := verifyToken(token)
	fmt.Println(role, id, err1, err2)
}
