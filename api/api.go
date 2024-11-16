package api


// @Summary Create a new user
// @Description Create a new user and return the created user
// @Tags users
// @Accept json
// @Produce json
// @Param user body User true "User data"
// @Success 201 {object} User
// @Failure 400 {object} ErrorResponse
// @Router /users [post]

//swag 注释的常用选项
// @Summary: 用于简要描述接口的功能。
// @Description: 详细描述接口的行为和用途。
// @Tags: 给接口打标签，便于分类。
// @Accept: 指定请求的内容类型，通常是 json 或 xml。
// @Produce: 指定响应的内容类型，通常是 json。
// @Param: 描述请求参数。支持 query, body, path, header 等类型。
// @Success: 描述成功响应的状态码和响应体。
// @Failure: 描述失败响应的状态码和响应体。
// @Router: 描述路由和 HTTP 方法（如 GET, POST, PUT 等）。
// func createUser(c *gin.Context) {
//     var user User
//     if err := c.ShouldBindJSON(&user); err != nil {
//         c.JSON(400, ErrorResponse{Message: "Invalid input"})
//         return
//     }
//     // 假设创建用户成功
//     c.JSON(201, user)
// }