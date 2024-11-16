# `auth` Package 使用说明

## 介绍

`auth` 包提供了处理用户认证相关功能的 API，主要包括生成和验证 token 的功能。它包括以下几个主要功能：

- **生成 token**：根据用户的角色生成一个相应的 token。
- **验证 token**：验证用户提供的 token 是否有效，并返回与之对应的用户 ID 和角色。
  
此包处理与数据库的交互，存储和获取 token 信息，确保 token 的有效性，防止 token 被篡改、过期或撤销。

---

## 主要类型

### `Token`
`Token` 类型表示一个用户的认证 token 数据。它的字段包括：

- `TokenID`：token 的唯一标识符。
- `TokenHash`：token 的哈希值，用于存储和比对。
- `TokenRevoked`：标识 token 是否被撤销。
- `TokenExpiry`：token 的过期时间。
- `UserID`：与 token 关联的用户 ID。

### `TokenDetail`
`TokenDetail` 类型存储了与 token 相关的详细信息，包括：

- `TokenID`：token 的 ID。
- `TokenCreatedAt`：token 创建时间。
- `ClientInfo`：客户端信息（可选，用于标识请求来源）。

### `UserPass`
`UserPass` 类型存储了用户的账户信息，包括：

- `UserID`：用户的 ID。
- `UserPassword`：用户密码的哈希值。
- `Role`：用户角色（如管理员、普通用户等）。
- `UserStatus`：用户账户状态。

---

## 主要功能

### `GiveAToken`
根据给定的角色生成一个 token。

#### 参数：
- `role`：用户的角色，类型为 `config.Role`。
- `userId`：用户的 ID，类型为 `string`。
- `clientInfo`：客户端信息，类型为 `string`，用于标识请求的来源（可选，但建议传入）。

#### 返回：
- `token`：生成的 token，以 `string` 形式返回。
- `error`：可能的错误类型。用户可能需要关注的错误包括：
  - `exception.ErrCodeUnfounded`：没有找到对应的 `user_id`。
  - `exception.UnmatchedRoleAndCode`：角色与 `user_id` 不匹配。

#### 示例：
```go
token, err := auth.GiveAToken(config.Admin, "user123", "clientA")
if err != nil {
    fmt.Println("Error generating token:", err)
} else {
    fmt.Println("Generated token:", token)
}
```

### `VerifyAToken`
验证用户提供的 token 是否有效，检查 token 是否存在、是否被篡改、是否过期以及是否被撤销。

#### 参数：
- `token`：需要验证的 token，类型为 `string`。

#### 返回：
- `user_id`：与 token 关联的用户 ID，类型为 `string`。
- `role`：用户的角色，类型为 `config.Role`。
- `error`：可能的错误类型。用户可能需要关注的错误包括：
  - `exception.TokenNotFound`：没有找到对应的 token。
  - `exception.TokenRevoked`：token 已被撤销。
  - `jwt.ErrTokenExpired`：token 已过期。
  - `jwt.ErrInvalidSignature`：token 无效，签名无法验证。

#### 示例：
```go
userId, role, err := auth.VerifyAToken("some_token")
if err != nil {
    fmt.Println("Error verifying token:", err)
} else {
    fmt.Println("User ID:", userId)
    fmt.Println("Role:", role)
}
```

---

## 错误处理

在使用该包时，可能会遇到以下几类错误：

- **token 未找到** (`exception.TokenNotFound`)：表示在数据库中未找到与提供的 token 匹配的数据。
- **token 被撤销** (`exception.TokenRevoked`)：表示该 token 已被撤销，无法再使用。
- **token 超时** (`jwt.ErrTokenExpired`)：表示 token 已过期。
- **无效签名** (`jwt.ErrInvalidSignature`)：表示 token 的签名无法验证，通常是恶意篡改导致的错误。

对于一些错误，如 `exception.TokenRevoked` 和 `jwt.ErrTokenExpired`，虽然它们表示 token 无效，但在某些情况下，这些错误可能是正常的（例如，token 已过期）。因此，错误返回时，通常会通过警告日志来提示用户，而不是直接中断程序执行。

---

## 数据库交互

该包会与数据库交互，以下是涉及的数据库表：

- **usersPass**：用于存储用户的账户信息，包括 `user_id`、密码的哈希值、角色、账户状态等。
- **tokens**：用于存储生成的 token 信息，包括 `token_id`、`token_hash`、`token_revoked` 和 `token_expiry`。
- **tokensDetails**：用于存储 token 的详细信息，包括 `token_id`、`token_created_at` 和 `client_info`。

在生成 token 时，首先会查询 `usersPass` 表验证用户身份，然后将生成的 token 信息插入 `tokens` 和 `tokensDetails` 表中。

---

## 注意事项

1. **客户端信息**：在调用 `GiveAToken` 时，建议提供客户端信息（如用户的设备信息或请求来源），以增强系统的可扩展性，尤其是在未来需要根据客户端信息来管理 token 时。
2. **token 有效期**：该包对 token 的过期时间进行了处理，确保过期的 token 无法继续使用。
3. **错误处理**：请确保在调用 API 时处理好可能的错误，并根据错误类型采取相应的操作。

---

## 结论

`auth` 包提供了一个简洁的方式来生成和验证用户 token，它通过与数据库的交互，确保 token 的安全性和有效性。通过合理的错误处理机制和日志记录，能够帮助开发者在实现用户认证时保持较高的系统安全性。
