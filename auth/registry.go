package auth

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"login/config"
	"login/db"
	"login/exception"
	"login/utils"
	"net/http"
	"time"
)

const (
	// 你的 Google reCAPTCHA secret 密钥
	secretKey = "6Lexl4sqAAAAAOzkLKgxOgrg5dj7gu1_mKc51N6w"
)

type RecaptchaResponse struct {
	Success     bool     `json:"success"`
	ChallengeTS string   `json:"challenge_ts"`
	Hostname    string   `json:"hostname"`
	Score       float32  `json:"score"`
	ErrorCodes  []string `json:"error-codes,omitempty"`
}

// 解密密码
func decryptPassWord(encryptedPassword string) string {
	return encryptedPassword
}

// 注册
func HandleRegistry(w http.ResponseWriter, r *http.Request) {
	// 获取前端发送的 reCAPTCHA Token
	//_ = r.FormValue("recaptchaToken")

	//// 验证 reCAPTCHA Token
	//success := sendToGoogle(token, "register")
	//if !success {
	//	// 返回401 Unauthorized 状态码
	//	http.Error(w, "reCAPTCHA 验证失败", http.StatusUnauthorized)
	//	return
	//}
	exception.PrintWarning(HandleRegistry, fmt.Errorf("recaptcha is currently not implemented due to the connection issue with google, 我会尽快恢复"))

	// 如果验证通过，继续
	// 获得字段
	type User struct {
		Email    string `json:"email"`
		Phone    string `json:"phone"`
		UserName string `json:"username"`
		Password string `json:"password"`
	}

	user := User{}

	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		exception.PrintError(HandleRegistry, err)
		return
	}

	// 对user数据进行处理
	user.Phone = utils.TrimExtraSpaces(user.Phone)
	user.Email = utils.TrimExtraSpaces(user.Email)
	user.UserName = utils.TrimExtraSpaces(user.UserName)
	user.Password = utils.TrimExtraSpaces(user.Password)

	//插入数据库
	// 可登录凭据
	var alias []string
	if user.Email != "" {
		alias = append(alias, user.Email)
	}
	if user.Phone != "" {
		alias = append(alias, user.Phone)
	}
	alias = append(alias, user.UserName)

	// 查询是否alias撞库，如果是，当用户名撞库返回409 Conflict，其他撞库返回400 Bad Request
	for _, item := range alias {
		result, err := db.ExecuteSQL(config.RoleAdmin, "SELECT * FROM usersaliases WHERE user_name = ?", item)
		if err != nil {
			exception.PrintError(HandleRegistry, err)
			return
		}
		row := result.(*sql.Rows)
		if row.Next() {
			if item == user.UserName {
				http.Error(w, "用户名已被注册", http.StatusConflict)
				return
			} else {
				http.Error(w, "邮箱或手机号已被注册", http.StatusForbidden)
				return
			}
		}
	}
	// 注册信息没问题
	w.WriteHeader(http.StatusOK)

	// 要插入的表
	type UserPass struct {
		UserPasswordHash string `db:"user_password_hash"`
		UserType         int    `db:"user_type"`
		UserStatus       string `db:"user_status"`
	}

	type UserAlias struct {
		UserName string `db:"user_name"`
		UserID   int    `db:"user_id"`
	}

	type UserInfo struct {
		UserID           int    `db:"user_id"`
		UserRegistryDate string `db:"user_registry_date"`
	}
	// 插入信息
	userPass := UserPass{
		UserPasswordHash: decryptPassWord(user.Password),
		UserType:         config.RolePassenger.Int(),
		UserStatus:       "active",
	}

	result, err := db.ExecuteSQL(config.RoleAdmin, "INSERT INTO userspass (user_password_hash, user_type, user_status) VALUES (?, ?, ?)", userPass.UserPasswordHash, userPass.UserType, userPass.UserStatus)
	if err != nil {
		exception.PrintError(HandleRegistry, err)
		return
	}
	// 应该是第二个结果？ 先断言
	userID64 := result.(int64)
	userID := int(userID64)

	// 插入Alias
	for _, item := range alias {
		userAlias := UserAlias{
			UserName: item,
			UserID:   userID,
		}
		_, err := db.Insert(config.RoleAdmin, "usersaliases", userAlias)
		if err != nil {
			exception.PrintError(HandleRegistry, err)
			return
		}
	}

	// 插入时间
	dateTime := time.Now().String()
	dateTime, err = utils.RegularizeTimeForMySQL(dateTime)

	userInfo := UserInfo{
		UserID:           userID,
		UserRegistryDate: dateTime,
	}
	_, err = db.Insert(config.RoleAdmin, "usersinfo", userInfo)
	if err != nil {
		exception.PrintError(HandleRegistry, err)
		return
	}
}
