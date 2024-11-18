### 配置项填写提醒

**请注意**：在 `config.yaml` 中，`database_names` 下的 **`passenger_db`** 和 **`driver_db`** 配置项需要相应的模块人员根据实际情况填写对应的数据库名称。请确保填写正确，避免使用占位符（如 `??????` 或 `?????`）。

- **`passenger_db`**：请填写乘客数据库的名称。
- **`driver_db`**：请填写司机数据库的名称。

不要修改 `admin_db` 和 `driver_db` 的原始配置项，只有 `passenger_db` 和 `driver_db` 的数据库名称需要根据实际情况填写。

---

### `config` 模块说明

在本项目中，`config` 模块负责加载和管理全局配置，配置项存储在 `config.yaml` 文件中。该模块在程序启动时自动加载配置，开发人员无需手动调用配置加载函数。

### `config.go` 主要内容

`config.go` 中定义了用于解析和存储配置的结构体，并且包含了一个全局变量 `AppConfig`，它保存了从配置文件中解析出来的所有配置项。

#### 1. **全局配置变量**
`AppConfig` 是一个全局变量，加载配置后存储在该变量中，供项目中其他模块访问。

```go
// AppConfig 静态全局变量载入
var AppConfig Config
```

#### 2. **配置结构体**
配置结构体映射了 `config.yaml` 中的配置项。常见的配置项包括数据库连接信息、JWT 配置等。

```go
// Config 是从 config.yaml 中提取的配置结构体
type Config struct {
    Database DatabaseConfig `yaml:"database_connection"`
    Server   Server         `yaml:"server"`
    DBNames  DatabaseNames  `yaml:"database_names"`
    Jwt      Jwt            `yaml:"jwt"`
}
```

- **DatabaseConfig**：包含数据库连接相关配置，如数据库主机、端口、用户名、密码等。
- **DatabaseNames**：指定不同数据库的名称。
- **Server**：包含服务器相关配置，如服务器端口。
- **Jwt**：JWT 配置项，包括不同角色的过期时间。

#### 3. **配置结构体的具体定义**

```go
type DatabaseConfig struct {
    Host     string `yaml:"host"`
    Port     int    `yaml:"port"`
    User     string `yaml:"user"`
    Password string `yaml:"password"`
}

type DatabaseNames struct {
    AdminDB     string `yaml:"admin_db"`
    PassengerDB string `yaml:"passenger_db"`  // 请按实际填充
    DriverDB    string `yaml:"driver_db"`     // 请按实际填充
}

type Server struct {
    Port string `yaml:"port"`
}

type Jwt struct {
    ExpirationHoursPass   int `yaml:"expiration_hours_passenger"`
    ExpirationHoursAdmin  int `yaml:"expiration_hours_admin"`
    ExpirationHoursDriver int `yaml:"expiration_hours_driver"`
}
```

### 如何添加新的全局配置项

如果需要在 `config.yaml` 中添加新的配置项，按以下步骤操作：

1. **修改结构体**：在 `config.go` 中的 `Config` 结构体中添加新的字段。例如，添加一个 `AppName` 配置项：

    ```go
    type Config struct {
        Database DatabaseConfig `yaml:"database_connection"`
        Server   Server         `yaml:"server"`
        DBNames  DatabaseNames  `yaml:"database_names"`
        Jwt      Jwt            `yaml:"jwt"`
        AppName  string         `yaml:"app_name"`  // 新增字段
    }
    ```

2. **更新 `config.yaml`**：在 `config.yaml` 中添加对应的配置项：

    ```yaml
    app_name: "MyApp"  # 新增的配置项
    ```

3. **重新加载配置**：配置加载操作已经在程序启动时自动完成，因此无需手动调用。

### 示例 `config.yaml`

以下是 `config.yaml` 的示例，展示了如何配置数据库连接、服务器端口、JWT 配置以及新增的 `app_name` 配置项。

```yaml
database_connection:
  host: "127.0.0.1"
  port: 3306
  user: "root"
  password: "123456"
  name: "schoolBus"

server:
  port: ":8888"

database_names:
  admin_db: "schoolbus"  # 请勿修改此项
  passenger_db: "??????"  # ******* 请填写此数据库名称 ********
  driver_db: "?????"  # ******* 请填写此数据库名称 ********

jwt:
  expiration_hours_passenger: 1
  expiration_hours_admin: 9999
  expiration_hours_driver: 24

app_name: "MyApp"  # 新增的配置项
```
