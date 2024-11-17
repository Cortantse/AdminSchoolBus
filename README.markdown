以下是重新设计后的项目总 Markdown，包含了模块概览、贡献者、文件修改权限等信息，格式简洁清晰：

---

# 项目总览

欢迎使用 **AdminSchoolBus** 项目！这是一个用于学校公交管理的系统，包含多个模块，每个模块负责系统中的不同功能。

## 模块一览

### 0. **[API 模块](https://github.com/Cortantse/AdminSchoolBus/blob/main/api/README.markdown)**  
用于存放符合 RESTful API 范式的函数，用于前后端交互。

- **贡献者**：  
  - **Cortantse**（模板提供）

- **可修改文件**：  
  - `api.go`（可增量添加/修改 API 函数）

- **禁止修改文件**：  
  - 无

---

### 1. **[Auth 模块](https://github.c`om/Cortantse/AdminSchoolBus/blob/main/auth/README.markdown)**  
负责用户认证与授权。包含用户登录、Token 生成与验证等功能。

- **贡献者**：  
  - **Cortantse**

- **可修改文件**：  
  - 无

- **禁止修改文件**：  
  - `auth/token.go`（禁止修改）  
  - `auth/token_api.go`（禁止修改，但提供 2 个公有函数供调用）  
  - `auth/load_es256_keys.go`（禁止修改）

---

### 2. **[Config 模块](https://github.com/Cortantse/AdminSchoolBus/blob/main/config/README.markdown)**  
管理项目的全局配置，包括数据库连接、JWT 配置等。

- **贡献者**：  
  - **Cortantse**

- **可修改文件**：  
  - `config/config.yaml`（可增量添加与修改，实际在项目目录中而非模块目录）  
  - `config/identities.go`（仅可增量添加，禁止修改）

- **禁止修改文件**：  
  - `config/config.go`（禁止修改）

---

### 3. **[DB 模块](https://github.com/Cortantse/AdminSchoolBus/blob/main/db/README.markdown)**  
负责数据库操作。支持数据库连接、查询与插入操作。

- **贡献者**：  
  - **Cortantse**

- **可修改文件**：  
  - 无

- **禁止修改文件**：  
  - `db/db.go`（禁止修改）  
  - `db/db_api.go`（禁止修改，但支持调用）

---

### 4. **[DriverShift 模块](https://github.com/Cortantse/AdminSchoolBus/blob/main/driverShift/README.markdown)**  
无

- **贡献者**：  
  - **shiganwen**  
  - **xuehaobing**

- **可修改文件**：  
  - 无

- **禁止修改文件**：  
  - 无

---

### 5. **[Exception 模块](https://github.com/Cortantse/AdminSchoolBus/blob/main/exception/README.markdown)**  
负责错误处理与日志记录。开发者可以自定义错误类型，并使用统一的错误记录函数。

- **贡献者**：  
  - **Cortantse**

- **可修改文件**：  
  - `exception/exception.go`（仅可增量添加自定义错误类型，禁止修改）  
  - `exception/exception_functions.go`（仅可增量添加处理错误的函数，禁止修改）

---

### 6. **[GPS 模块](https://github.com/Cortantse/AdminSchoolBus/blob/main/gps/README.markdown)**  
无

- **贡献者**：  
  - **shiganwen**  
  - **xuehaobing**

- **可修改文件**：  
  - 无

---

## 模块文件修改权限说明

| 模块               | 文件                                        | 说明                                            |
|--------------------|--------------------------------------------------|-------------------------------------------------|
| **[API 模块](https://github.com/Cortantse/AdminSchoolBus/blob/main/api/README.markdown)**   | `api.go`                                         | 可增量添加/修改 API 函数                        |
| **[Auth 模块](https://github.com/Cortantse/AdminSchoolBus/blob/main/auth/README.markdown)**  | 无                                               | **禁止修改**，仅调用公有函数                    |
|                    | `auth/token.go`，`auth/token_api.go`，`auth/load_es256_keys.go` | **禁止修改**                                    |
| **[Config 模块](https://github.com/Cortantse/AdminSchoolBus/blob/main/config/README.markdown)**| `config/config.yaml`，`config/identities.go`     | 可增量添加与修改（`identities.go` 仅可增量添加）|
|                    | `config/config.go`                               | **禁止修改**                                    |
| **[DB 模块](https://github.com/Cortantse/AdminSchoolBus/blob/main/db/README.markdown)**      | 无                                               | **禁止修改**，仅可调用相关接口                  |
|                    | `db/db.go`，`db/db_api.go`                       | **禁止修改**                                    |
| **[DriverShift 模块](https://github.com/Cortantse/AdminSchoolBus/blob/main/driverShift/README.markdown)** | 无                                               | 无修改权限                                      |
| **[Exception 模块](https://github.com/Cortantse/AdminSchoolBus/blob/main/exception/README.markdown)**  | `exception/exception.go`，`exception/exception_functions.go` | 可增量添加与修改自定义错误类型与错误处理函数   |
| **[GPS 模块](https://github.com/Cortantse/AdminSchoolBus/blob/main/gps/README.markdown)**     | 无                                               | 无修改权限                                      |

---

## 总结

该文档为开发人员提供了清晰的模块概览、文件修改权限和贡献者信息。请遵守各模块的文件修改规定，确保代码的一致性和模块的可维护性。如果需要进一步了解某个模块的细节，可以**点击模块链接**访问具体文档。
