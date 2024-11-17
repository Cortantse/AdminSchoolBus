# API 文档 - 接口注释说明

本项目使用 [Swag](https://github.com/swaggo/swag) 来自动生成 API 文档，接口注释遵循 Swagger 标准。以下是常用的 Swag 注释选项及其作用说明。

## Swag 注释选项

### `@Summary`
简要描述接口的功能。

- **用途**：提供接口的简短描述，通常用于快速了解接口的目的。
- **示例**：`@Summary Create a new user`

### `@Description`
详细描述接口的行为和用途。

- **用途**：提供关于接口更多的细节，帮助用户了解接口的工作原理。
- **示例**：`@Description Create a new user and return the created user`

### `@Tags`
给接口打标签，便于分类和组织接口文档。

- **用途**：通过标签将接口分组，便于用户查找和分类。
- **示例**：`@Tags users`

### `@Accept`
指定请求的内容类型。

- **用途**：定义该接口接受的请求内容类型，通常为 `application/json` 或 `application/xml`。
- **示例**：`@Accept json`

### `@Produce`
指定响应的内容类型。

- **用途**：定义该接口返回的数据格式，通常为 `application/json` 或 `application/xml`。
- **示例**：`@Produce json`

### `@Param`
描述请求参数。

- **用途**：描述请求中的参数。可以是 `query`、`body`、`path` 或 `header` 等类型。
- **示例**：`@Param user body User true "User data"`

### `@Success`
描述成功响应的状态码和响应体。

- **用途**：说明接口成功时返回的 HTTP 状态码以及响应数据的结构。
- **示例**：`@Success 201 {object} User`

### `@Failure`
描述失败响应的状态码和响应体。

- **用途**：说明接口失败时返回的 HTTP 状态码以及错误响应数据的结构。
- **示例**：`@Failure 400 {object} ErrorResponse`

### `@Router`
描述路由和 HTTP 方法（如 GET、POST、PUT 等）。

- **用途**：定义接口的路径和 HTTP 请求方法（如 `GET`、`POST` 等）。
- **示例**：`@Router /users [post]`

---

## 示例

以下是一个使用 Swag 注释的接口示例：

```go
// @Summary Create a new user
// @Description Create a new user and return the created user
// @Tags users
// @Accept json
// @Produce json
// @Param user body User true "User data"
// @Success 201 {object} User
// @Failure 400 {object} ErrorResponse
// @Router /users [post]
func createUser(c *gin.Context) {
    var user User
    if err := c.ShouldBindJSON(&user); err != nil {
        c.JSON(400, ErrorResponse{Message: "Invalid input"})
        return
    }
    // 假设创建用户成功
    c.JSON(201, user)
}
