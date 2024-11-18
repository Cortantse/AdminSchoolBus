### `exception.go`：定义常见的错误类型

`exception.go` 文件主要用于定义一些常见的错误类型。这些错误类型在应用中可以统一引用，用于表示特定的错误情境。开发者不需要修改这个文件，只需要使用这些预定义的错误类型来处理常见的错误场景。

#### 1. **常见错误类型定义**

```go
// exception.go - 定义常见的错误类型

var ErrCodeUnfounded = errors.New("没有找到账户代码")
var UnmatchedRoleAndCode = errors.New("账户代码和角色不匹配")

var TokenNotFound = errors.New("token not found")
var TokenRevoked = errors.New("token revoked")

var UserIDNotFound = errors.New("user_id not found")
```

这些错误类型包括：

- `ErrCodeUnfounded`: 没有找到账户代码。
- `UnmatchedRoleAndCode`: 账户代码和角色不匹配。
- `TokenNotFound`: 未找到 token。
- `TokenRevoked`: token 已被撤销。
- `UserIDNotFound`: 未找到用户 ID。

#### 2. **如何使用**

这些错误类型通常直接用来作为函数返回值中的错误部分。例如：

```go
if err != nil {
    return "", exception.ErrCodeUnfounded
}
```

---

### `exception_functions.go`：记录错误与警告

`exception_functions.go` 文件定义了两个**重要**的函数，分别用于打印错误和警告。开发者在代码中处理错误时，**应该**使用这两个函数来记录错误和警告信息。它们会帮助开发者在调试和运维过程中清晰地追踪问题。

#### 1. **`PrintError`：打印错误信息**

`PrintError` 用于记录程序中的错误，并打印错误的详细信息，包括错误发生的函数位置和具体的错误信息。

```go
// exception_functions.go - 打印错误的函数

func PrintError(fn interface{}, err error) {
    // 定义颜色 ANSI 转义序列
    red := "\033[31m"  // 红色字体
    bold := "\033[1m"  // 加粗
    reset := "\033[0m" // 重置样式

    // 获取函数名
    pc := runtime.FuncForPC(reflect.ValueOf(fn).Pointer())
    if pc != nil {
        log.Printf("%s%sError occurs in %s: %s%s\n", bold, red, pc.Name(), err.Error(), reset)
    } else {
        log.Printf("%s%sError occurs in unknown function: %s%s\n", bold, red, err.Error(), reset)
    }
}
```

- **功能**：该函数会记录一个错误，并在日志中输出详细信息，包括错误发生的函数名称、错误消息等，帮助开发人员快速定位问题。

**使用方法**：

当发生错误时，调用 `PrintError` 函数记录日志：

```go
signer, err := jwt.NewES256(pubKey, priKey)
if err != nil {
    exception.PrintError(generateToken, err)
    return Token{}, err
}
```

- `fn` 参数：传入当前函数（通常是出错的函数）名，用于定位错误源。
- `err` 参数：传入错误信息，通常是通过 `fmt.Errorf` 获取的错误。

#### 2. **`PrintWarning`：打印警告信息**

`PrintWarning` 用于记录程序中的警告信息。警告信息不一定会导致程序终止，但仍然值得关注。

```go
// exception_functions.go - 打印警告的函数

func PrintWarning(fn interface{}, err error) {
    // 定义颜色 ANSI 转义序列
    yellow := "\033[33m" // 黄色字体
    bold := "\033[1m"    // 加粗
    reset := "\033[0m"   // 重置样式

    // 获取函数名
    pc := runtime.FuncForPC(reflect.ValueOf(fn).Pointer())
    if pc != nil {
        log.Printf("%s%sWarning in %s: %s%s\n", bold, yellow, pc.Name(), err.Error(), reset)
    } else {
        log.Printf("%s%sWarning in unknown function: %s%s\n", bold, yellow, err.Error(), reset)
    }
}
```

- **功能**：该函数会记录一个警告，并在日志中输出详细信息，包括警告发生的函数名称、警告消息等。

**使用方法**：

当发生警告时，调用 `PrintWarning` 函数记录日志：

```go
if tokens[0].TokenRevoked {
    exception.PrintWarning(VerifyAToken, fmt.Errorf("token is revoked"))
    return "", config.Unknown, exception.TokenRevoked
}
```

- `fn` 参数：传入当前函数（通常是发出警告的函数）名，用于定位警告源。
- `err` 参数：传入警告信息，通常是通过 `fmt.Errorf` 获取的警告。

---

### **如何正确使用 `exception_functions.go` 和 `exception.go`**

#### **1. 错误的处理：**

- 当程序遇到需要记录的错误时，使用 `exception.PrintError` 打印错误，并传入相关的错误信息。通过这种方式，错误可以在日志中进行详细记录，帮助开发人员快速追踪问题。
- 错误类型（如 `exception.ErrCodeUnfounded`）可以作为函数返回的错误类型返回。

**示例：**

```go
signer, err := jwt.NewES256(pubKey, priKey)
if err != nil {
    exception.PrintError(generateToken, err)  // 打印错误
    return Token{}, err  // 返回错误
}
```

#### **2. 警告的处理：**

- 如果程序遇到的错误是警告性错误（即不是致命错误，可以继续执行），使用 `exception.PrintWarning` 打印警告。警告信息有助于开发人员识别潜在问题并进行监控，但不会影响程序的继续运行。

**示例：**

```go
if tokens[0].TokenRevoked {
    exception.PrintWarning(VerifyAToken, fmt.Errorf("token is revoked"))  // 打印警告
    return "", config.Unknown, exception.TokenRevoked  // 返回警告
}
```

---

### **总结**

- **`exception.go`** 定义了错误类型，这些错误类型应该用于标识特定的错误场景，帮助开发人员识别问题。
- **`exception_functions.go`** 提供了两个函数 `PrintError` 和 `PrintWarning`，分别用于记录错误和警告。开发人员在出现错误或警告时，应该调用这两个函数来记录日志，确保能够追踪和定位问题。

开发人员在开发过程中，务必遵循以下规则：
- 对于错误，使用 `exception.PrintError` 记录并返回。
- 对于警告，使用 `exception.PrintWarning` 记录并继续执行。
