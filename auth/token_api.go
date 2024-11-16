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
	UserID       string `db:"user_id"`
	UserPassword string `db:"user_password_hash"`
	Role         int    `db:"user_type"`
	UserStatus   string `db:"user_status"`
}

// GiveAToken 根据role生成一个token，同时做检测是否user_id和role在数据库中是正确的
// Parameters:
//   - role: 需要生成的对象
//   - user_id: 改token对应的对象id
//   - clientInfo: 客户端信息，用于生成token，你可以不传，但建议你传，方便系统后期拓展性
//     大概的格式是这样： 冰箱来写
//
// Returns:
//   - token: 生成的token，以string形式
//   - error: 错误信息，如果有
//     使用者**可能**需要处理的错误类型有：****
//     1、exception.ErrCodeUnfounded没有找到对应的user_id
//     2、exception.UnmatchedRoleAndCode 传入的role和user_id不匹配
//     在正确使用函数的情况下，一般不会触发其它exceptions
func GiveAToken(role config.Role, userId string, clientInfo string) (string, error) {
	// 如果不传入clientInfo警告
	if clientInfo == "" {
		exception.PrintWarning(GiveAToken, fmt.Errorf("clientInfo is empty, you should parse data from client and pass it to this function"))
	}

	// 检测role和user_id是否存在，预检查
	if role == config.Unknown {
		exception.PrintError(GiveAToken, fmt.Errorf("error in GiveAToken: role is unknown"))
		return "", fmt.Errorf("error in GiveAToken: role is unknown")
	}

	// 创建数组，方便装结果
	var tems []UserPass

	params := []interface{}{userId}

	// 检测role和user_id是否匹配，正式检查
	err := db.Select(config.RoleAdmin, "usersPass", &tems, true,
		[]string{}, []string{"user_id = (?)"}, params, "user_id", 1, 0, "", "")
	if err != nil {
		exception.PrintError(GiveAToken, err)
		return "", err
	}

	if len(tems) == 0 {
		return "", exception.ErrCodeUnfounded
	}

	if tems[0].Role != int(role) {
		return "", exception.UnmatchedRoleAndCode
	}

	// 生成一个token结构体
	token, err := generateToken(role, userId)
	if err != nil {
		exception.PrintError(GiveAToken, err)
		return "", err
	}

	// 插入token前需要把token的expiry改一下，去掉时区，mysql的datetime不支持
	token.TokenExpiry, err = utils.RegularizeTimeForMySQL(token.TokenExpiry)
	if err != nil {
		exception.PrintError(GiveAToken, err)
		return "", err
	}

	// 存储token相关信息，到这一层如果还报错就属于意外错误，请您检查您的调用是否正确
	// 存储进入tokens表
	_, err = db.Insert(config.RoleAdmin, "tokens", token)
	if err != nil {
		exception.PrintError(GiveAToken, err)
		return "", err
	}

	// 获取数据库系统生成的tokenID
	var tokens []Token
	// 搜索相同token_hash和user_id的token，来获取刚插入的token
	err = db.Select(config.RoleAdmin, "tokens", &tokens, false,
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

	// 有了tokenID之后存储进入tokensDetails表
	tokenDetail := TokenDetail{
		TokenID:        tokenID,
		TokenCreatedAt: now,
		ClientInfo:     clientInfo,
	}
	_, err = db.Insert(config.RoleAdmin, "tokensDetails", tokenDetail)
	if err != nil {
		exception.PrintError(GiveAToken, err)
		return "", err
	}

	return token.TokenHash, nil
}

// token鉴权这里简化逻辑，不做权限控制，即只会检查用户提供的token是否存在且合法
// 并返回用户的id和用户的身份role，由调用者根据token原有主人user_id和身份role**自行处理身份问题**
// ！！！我们的三个JWT字段由ES256签名，理论上重复概率非常小

// VerifyAToken 鉴定用户提供的token是否合法，若不合法，抛出error，合法则返回用户id和用户身份role
//
// Parameters:
//   - token: 用户提供的token
//
// Returns:
//   - user_id: 用户id
//   - role: 用户身份
//   - error: 错误信息，如果有
//     使用者**可能**需要处理的错误类型有：****
//     1、exception.TokenNotFound 没有找到对应的token
//     2、exception.TokenRevoked 对应的token已经被撤销
//     3、jwt.ErrTokenExpired 对应的token已经过期
//     4、jwt.ErrInvalidSignature 对应的token无效，signature无法通过
//     为方便起见，您可以在error！=nil的情况下，直接使用role和user_id
//     当error = nil时拒绝请求即可
//     在正确使用函数的情况下，一般不会触发其它exceptions
func VerifyAToken(token string) (string, config.Role, error) {
	// 函数只验证token： 1、是否存在  2、是否被篡改  3、是否过期   4、是否被撤销

	// 1、是否存在
	var tokens []Token
	err := db.Select(config.RoleAdmin, "tokens", &tokens, false,
		[]string{"token_id"}, []string{"token_hash = ?"}, []interface{}{token}, "token_id", 2, 0, "", "")
	if len(tokens) == 0 {
		exception.PrintError(VerifyAToken, fmt.Errorf("token not found"))
		return "", config.Unknown, exception.TokenNotFound
	} else if err != nil {
		// select存在报错
		exception.PrintError(VerifyAToken, err)
		return "", config.Unknown, err
	} else if len(tokens) > 1 {
		// 理论上这不可能发生，但是为了安全起见，我们还是检查一下，如果此处报错，请联系我
		exception.PrintError(VerifyAToken, fmt.Errorf("token is duplicated, given the same token_hash have the same user_id"))
		panic("token is duplicated, given the same user_id have the same token_hash")
	}

	// 2&3、是否被篡改和是否超时
	role, userId, err := verifyToken(token)
	if err != nil {
		// 这是有可能发生的，发warning
		exception.PrintWarning(VerifyAToken, err)
		return "", config.Unknown, err
	}
	// 这里可能出现的异常有：jwt.ErrInvalidSignature, jwt.ErrTokenExpired

	//4、是否被撤销
	if tokens[0].TokenRevoked {
		// 这是有可能发生的，发warning
		exception.PrintWarning(VerifyAToken, fmt.Errorf("token is revoked"))
		return "", config.Unknown, exception.TokenRevoked
	}

	return userId, role, nil
}
