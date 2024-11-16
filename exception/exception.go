package exception

import "errors"

var ErrCodeUnfounded = errors.New("没有找到账户代码")
var UnmatchedRoleAndCode = errors.New("账户代码和角色不匹配")

var TokenNotFound = errors.New("token not found")
var TokenRevoked = errors.New("token revoked")

var UserIDNotFound = errors.New("user_id not found")
