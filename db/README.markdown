### db 模块简介

`db` 模块是一个面向结构化数据库交互的工具包，提供了一组安全且灵活的数据库操作方法，支持多种 SQL 功能，包括插入、更新、查询和删除操作。该模块对常见的操作进行了封装，简化了开发过程，同时支持复杂的 SQL 查询需求。

---

### 快速上手

`db` 模块是一个封装了常见数据库操作的工具包，旨在简化与数据库交互的代码编写。它提供了一系列高效、专业且实用的功能，满足不同层次的数据库操作需求。以下是模块功能的简要指引：
你可以找到和数据库操作的例子在exmaple模块中
- 1. [简单的例子](https://github.com/Cortantse/AdminSchoolBus/blob/main/example/example_simple.go)
- 2. [复杂的例子](https://github.com/Cortantse/AdminSchoolBus/blob/main/example/example_complicated.go)

#### 你想完成什么操作？

1. **快速初始化数据库连接**：  
   - 调用 `InitDB`，即可根据角色选择合适的数据库连接。
   - **场景**：初始化数据库连接时使用。

2. **执行常见 SQL 语句（INSERT/UPDATE/DELETE/SELECT）**：  
   - 使用 `ExecuteSQL`，编写通用 SQL 语句并传入参数即可。
   - **场景**：需要精准控制 SQL 语句时，比如插入新记录或删除某条数据。

3. **插入数据（单条或批量）**：  
   - 使用 `Insert`，将结构体或切片传递进去，无需编写复杂 SQL。
   - **场景**：快速插入大量数据，减少手动 SQL 编写的麻烦。

4. **构造复杂查询（支持动态条件、分页、排序等）**：  
   - 调用 `SelectEasy`，传入查询条件和参数即可。
   - **场景**：需要安全、灵活的查询功能，避免直接书写 SQL。

5. **处理高风险 SQL 操作（不推荐频繁使用）**：  
   - 使用 `UnSafeExecuteSQL`，直接执行 SQL 语句。
   - **场景**：需要执行不常见的操作（如表结构修改），但必须确保 SQL 来源安全。

**推荐用法**：
- **简单插入/更新/删除**：使用 `ExecuteSQL`。
- **条件查询**：使用 `ExecuteSQL` 或 `SelectEasy`。
- **数据批量插入**：使用 `Insert`。
- **复杂 SQL（事务、视图等）**：在保证安全的前提下使用 `UnSafeExecuteSQL`。


#### 为什么选择这个模块？

- **灵活性**：封装了通用和高级数据库操作，适应不同需求。
- **安全性**：对于大多数场景，提供了参数化查询和自动检查，避免 SQL 注入风险。
- **简化代码**：封装复杂逻辑，让开发者专注于核心业务逻辑。

---

### 公有函数介绍

#### 1. `InitDB`

**功能**：初始化数据库连接，这个函数已在main中被调用，你不需要调用。

**参数**：
- `chooseDB`：枚举类型，指定使用的数据库（如 `RoleAdmin`, `RolePassenger` 等）。




---

#### 2. `ExecuteSQL`

**功能**：执行通用 SQL 语句（支持至少四大语句 SELECT, INSERT, UPDATE, DELETE），并提供安全保护和审计。

**参数**：
- `role`：数据库角色。
- `sqlStatement`：带占位符的 SQL 语句。
- `args`：SQL 参数。

**示例**：插入数据
```go
sql := "INSERT INTO users (name, age) VALUES (?, ?)"
_, err := db.ExecuteSQL(config.RoleAdmin, sql, "Alice", 30)
if err != nil {
    fmt.Println("插入失败：", err)
}
```

---

#### 3. `Insert`

**功能**：通用插入函数，支持单条和批量插入。

**参数**：
- `role`：数据库角色。
- `tableName`：目标表名。
- `records`：结构体或结构体切片。

**示例**：批量插入
```go
users := []User{
    {Name: "Alice", Age: 25},
    {Name: "Bob", Age: 30},
}
_, err := db.Insert(config.RoleAdmin, "users", users)
if err != nil {
    fmt.Println("批量插入失败：", err)
}
```

---

#### 4. `SelectEasy`

**功能**：封装的查询函数，支持动态查询条件、排序、分页等功能。

**参数**：
- `role`：数据库角色。
- `tableName`：目标表名。
- `dest`：存储查询结果的结构体数组指针。
- `conditionFields`：条件字段（如 `["age > ?", "name = ?"]`）。
- `params`：查询参数。

**示例**：条件查询
```go
var users []User
err := db.SelectEasy(config.RoleAdmin, "users", &users, true, nil, []string{"age > ?"}, []interface{}{20}, "age ASC", 10, 0, "", "")
if err != nil {
    fmt.Println("查询失败：", err)
}
```

---

#### 5. `UnSafeExecuteSQL`

**功能**：直接执行 SQL 语句，支持所有合法的 SQL 操作，但不提供安全保护。

**参数**：
- `role`：数据库角色。
- `sqlStatement`：完整的 SQL 语句。

**示例**：清空表数据
```go
sql := "TRUNCATE TABLE users"
_, err := db.UnSafeExecuteSQL(config.RoleAdmin, sql)
if err != nil {
    fmt.Println("操作失败：", err)
}
```

---

### 总结

`db` 模块主要功能包括：

1. **数据库初始化**：通过 `InitDB` 方法连接指定数据库。
2. **SQL 执行**：
   - 通用 SQL：使用 `ExecuteSQL` 快速执行常见操作（INSERT、UPDATE、DELETE）。
   - 安全查询：通过 `SelectEasy` 进行条件、分页等复杂查询。
   - 插入操作：使用 `Insert` 处理单条或批量数据插入。
   - 高风险操作：`UnSafeExecuteSQL` 用于不带安全校验的复杂 SQL。

根据项目需求选择适当的函数，既能提升开发效率，又能确保安全性。
