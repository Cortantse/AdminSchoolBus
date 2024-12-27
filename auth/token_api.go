package auth

import (
	"fmt"
	"login/config"
	"login/db"
	"login/exception"
	"login/utils"
	"time"
)

type Token struct {
	TokenID      string `db:"token_id"`
	TokenHash    string `db:"token_hash"`
	TokenRevoked bool   `db:"token_revoked"`
	TokenExpiry  string `db:"token_expiry"`
	UserID       string `db:"user_id"`
}

type TokenDetail struct {
	TokenID        string `db:"token_id"`
	TokenCreatedAt string `db:"token_created_at"`
	ClientInfo     string `db:"token_client"`
}

type UserPass struct {
	UserID       int    `db:"user_id"`
	UserPassword string `db:"user_password_hash"`
	Role         int    `db:"user_type"`
	UserStatus   string `db:"user_status"`
}

// GiveAToken 根据role生成一个对应的token
// Parameters:
//   - role: 需要生成的对象
//   - user_id: 该token对应的对象id，该函数为验证该字段对应的role是否正确
//   - clientInfo: 客户端信息，用于生成token，可选，但建议传，方便系统后期拓展性
//
// Returns:
//   - token: 生成的token，以string形式
//   - error: 错误信息，如果有
//     使用者**可能**需要关注的错误类型有：
//     1、exception.ErrCodeUnfounded没有找到对应的user_id **正确情况下不会发生
//     2、exception.UnmatchedRoleAndCode 传入的role和user_id不匹配  **正确情况下不会发生
//     在正确使用函数的情况下，一般不会触发其它exceptions
func GiveAToken(role config.Role, userId string, clientInfo string) (string, error) {
	// 如果不传入clientInfo会发出警告
	if clientInfo == "" {
		exception.PrintWarning(GiveAToken, fmt.Errorf("clientInfo is empty, you should parse data from client and pass it to this function"))
	}

	// 检测role是否正确
	if role == config.Unknown {
		exception.PrintError(GiveAToken, fmt.Errorf("error in GiveAToken: role is unknown"))
		return "", fmt.Errorf("error in GiveAToken: role is unknown")
	}

	// 创建数组，方便装结果
	var tems []UserPass
	// 创建参数表
	params := []interface{}{userId}

	// 查询数据库中是否有对应的user_id的账户
	err := db.SelectEasy(config.RoleAdmin, "userspass", &tems, true,
		[]string{}, []string{"user_id = (?)"}, params, "user_id", 1, 0, "", "")
	if err != nil {
		exception.PrintError(GiveAToken, err)
		return "", err
	}
	// 如果没有找到对应的user_id，则返回错误
	if len(tems) == 0 {
		return "", exception.ErrCodeUnfounded
	}
	// 如果找到了，但是role不匹配，则返回错误
	if tems[0].Role != int(role) {
		return "", exception.UnmatchedRoleAndCode
	}

	// 没问题，下面开始生成一个token结构体
	token, err := generateToken(role, userId)
	if err != nil {
		exception.PrintError(GiveAToken, err)
		return "", err
	}

	// 将token存入sql前需要把token的expiry改一下，去掉时区，mysql的datetime不支持
	token.TokenExpiry, err = utils.RegularizeTimeForMySQL(token.TokenExpiry)
	if err != nil {
		exception.PrintError(GiveAToken, err)
		return "", err
	}

	// 存储进入tokens表
	_, err = db.Insert(config.RoleAdmin, "tokens", token)
	if err != nil {
		exception.PrintError(GiveAToken, err)
		return "", err
	}

	// 获取数据库系统生成的tokenID
	var tokens []Token
	// 搜索相同token_hash和user_id的token，来获取刚插入的token
	err = db.SelectEasy(config.RoleAdmin, "tokens", &tokens, false,
		[]string{"token_id"}, []string{"token_hash = ? AND user_id = ?"}, []interface{}{token.TokenHash, userId}, "token_id", 1, 0, "", "")
	if err != nil {
		exception.PrintError(GiveAToken, err)
		return "", err
	}
	// 不可能没有，刚插入
	tokenID := tokens[0].TokenID

	// 获取现在的时间
	now, err := utils.RegularizeTimeForMySQL(time.Now().String())
	if err != nil {
		exception.PrintError(GiveAToken, err)
		return "", err
	}

	// 有了tokenID之后存储token更多详细的信息
	tokenDetail := TokenDetail{
		TokenID:        tokenID,
		TokenCreatedAt: now,
		ClientInfo:     clientInfo,
	}
	_, err = db.Insert(config.RoleAdmin, "tokensdetails", tokenDetail)
	if err != nil {
		exception.PrintError(GiveAToken, err)
		return "", err
	}

	return token.TokenHash, nil
}

// VerifyAToken 鉴定用户提供的token是否合法
// 则若不合法，抛出error；合法则返回token对应的用户id和用户身份role
// 这里只做token的合法性（是否被篡改、过期检查），不做权限控制
//
// Parameters:
//   - token: 用户提供的token
//
// Returns:
//   - user_id: 用户id
//   - role: 用户身份
//   - error: 错误信息，如果有
//     使用者**可能**需要关注的错误类型有：****
//     1、exception.TokenNotFound 没有找到对应的token **正确情况下不会发生
//     2、exception.TokenRevoked 对应的token已经被撤销 **可能发生，这是一个warning而非真的不可恢复的错误
//     3、jwt.ErrTokenExpired 对应的token已经过期 **可能发生，这是一个warning而非真的不可恢复的错误
//     4、jwt.ErrInvalidSignature 对应的token无效，signature无法通过 **恶意情况才发生
func VerifyAToken(token string) (string, config.Role, error) {
	// 函数只验证token： 1、是否存在  2、是否被篡改  3、是否过期   4、是否被撤销

	//-1、是否存在
	// 获取数据库中对应token的信息
	var tokens []Token
	err := db.SelectEasy(config.RoleAdmin, "tokens", &tokens, false,
		[]string{"token_id"}, []string{"token_hash = ?"}, []interface{}{token}, "token_id", 2, 0, "", "")
	// 没有找到该token
	if len(tokens) == 0 {
		// 内部日志，不要报错 // 后端请尽快实现日志！！！！！！！！！！！！
		return "", config.Unknown, exception.TokenNotFound
	} else if err != nil {
		// select函数存在报错
		exception.PrintError(VerifyAToken, err)
		return "", config.Unknown, err
	} else if len(tokens) > 1 {
		// 理论上这不可能发生，一般来说token几乎是唯一的很难有重复；如果此处报错，请联系我，这里直接panic
		exception.PrintError(VerifyAToken, fmt.Errorf("token is duplicated, given the same token_hash have the same user_id"))
		panic("token is duplicated, given the same user_id have the same token_hash")
	}

	// 2&3、是否被篡改和是否超时
	// 这里可能出现的异常有：jwt.ErrInvalidSignature, jwt.ErrTokenExpired
	role, userId, err := verifyToken(token)
	if err != nil {
		// 这是有可能发生的，发warning
		exception.PrintWarning(VerifyAToken, err)
		return "", config.Unknown, err
	}

	//4、是否被撤销
	if tokens[0].TokenRevoked {
		// 这是有可能发生的，发warning
		exception.PrintWarning(VerifyAToken, fmt.Errorf("token is revoked"))
		return "", config.Unknown, exception.TokenRevoked
	}

	return userId, role, nil
}

// ReturnUserIDFromToken 从token中获取用户id，可能会报错
func ReturnUserIDFromToken(token string) (string, error) {
	_, userId, err := verifyToken(token)
	if err != nil {
		exception.PrintWarning(ReturnUserIDFromToken, err)
		return "", err
	}
	return userId, nil
}
