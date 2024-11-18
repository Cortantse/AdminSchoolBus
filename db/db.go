package db

import (
	"fmt"
	"github.com/jmoiron/sqlx"
	"login/config"
	"login/exception"
	"reflect"
	"regexp"
	"strings"
)

// DBError 自定义数据库错误类型，用于封装 SQL 执行中的错误。
type DBError struct {
	FuncName    string        // 函数名
	Err         error         // 错误信息
	SQL         string        // SQL 语句
	QueryParams []interface{} // 查询参数
}

// Error 实现了 error 接口，返回自定义错误的字符串表示
func (e *DBError) Error() string {
	return fmt.Sprintf("[%s] 出现错误: %v\nSQL: %s\n参数: %v", e.FuncName, e.Err, e.SQL, e.QueryParams)
}

// getStructFields 获取结构体的字段名，支持通过 db 标签来映射数据库列名。
// 如果字段名没有 db 标签，自动转换为蛇形命名法，并检测其是否符合蛇形命名规则。
//
// 参数:
//   - t reflect.Type: 要获取字段名的结构体类型。
//
// 返回:
//   - []string: 结构体字段对应的数据库列名列表。
//   - error: 如果字段名不符合蛇形命名法，则返回错误。
func getStructFields(t reflect.Type) ([]string, error) {
	var fields []string

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		// 跳过匿名字段
		if field.Anonymous {
			continue
		}

		// 获取 db 标签的值
		dbTag := field.Tag.Get("db")
		if dbTag == "" {
			// 如果没有 db 标签，则使用结构体字段名，并将驼峰命名法转换为蛇形命名法
			dbTag = toSnakeCase(field.Name)

			// 检查转换后的蛇形命名是否符合规则
			if !isSnakeCase(dbTag) {
				exception.PrintError(getStructFields, fmt.Errorf("field name '%s' is not in snake_case format", field.Name))
				return nil, fmt.Errorf("field name '%s' is not in snake_case format", field.Name)
			}
		}

		// 跳过被标记为 "-" 的字段
		if dbTag == "-" {
			continue
		}

		fields = append(fields, dbTag)
	}

	if len(fields) == 0 {
		exception.PrintError(getStructFields, fmt.Errorf("no valid fields found in struct %s", t.Name()))
		return nil, fmt.Errorf("no valid fields found in struct %s", t.Name())
	}

	return fields, nil
}

// toSnakeCase 将驼峰命名法的字符串转换为蛇形命名法。
// 例如，"FirstName" -> "first_name"
func toSnakeCase(str string) string {
	var result []rune
	for i, c := range str {
		if i > 0 && 'A' <= c && c <= 'Z' {
			result = append(result, '_')
		}
		result = append(result, rune(strings.ToLower(string(c))[0]))
	}
	return string(result)
}

// buildInsertPlaceholdersAndValues 构建 SQL 插入语句的占位符和对应的值。
// 该函数根据传入的结构体字段，生成 SQL 插入语句中的占位符（`?`）和实际插入的值列表。
//
// 参数:
//   - v reflect.Value: 要获取字段值的结构体。
//
// 返回:
//   - string: SQL 插入语句中的占位符部分（例如：`?, ?, ?`）。
//   - []interface{}: 结构体字段的实际值列表，作为参数传入 SQL 执行。
func buildInsertPlaceholdersAndValues(v reflect.Value) (string, []interface{}) {
	var placeholders []string
	var values []interface{}
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		placeholders = append(placeholders, "?")
		values = append(values, field.Interface())
	}
	return strings.Join(placeholders, ","), values
}

// isSnakeCase 检查给定的字符串是否符合蛇形命名法。
// 蛇形命名法要求字符串全部小写，单词之间使用下划线分隔，且不能以下划线开头或结尾。
func isSnakeCase(str string) bool {
	// 使用正则表达式检查是否是小写字母和下划线的组合
	match, _ := regexp.MatchString(`^[a-z]+(_[a-z]+)*$`, str)
	return match
}

// buildQueryConditions 构建查询条件的字段和值列表。
// 如果传入的是结构体，会自动解析其字段及对应的值。
// 如果传入的是切片，会遍历切片中的每个条件并构建对应的字段和值列表。
func buildQueryConditions(conditions interface{}) ([]string, []interface{}, error) {
	var conditionFields []string
	var conditionValues []interface{}

	// 获取条件的类型
	rv := reflect.ValueOf(conditions)
	switch rv.Kind() {
	case reflect.Struct:
		// 结构体形式的条件
		fields, values, err := getStructFieldsAndValues(rv)
		if err != nil {
			exception.PrintError(buildQueryConditions, err)
			return nil, nil, err
		}
		conditionFields = fields
		conditionValues = values

	case reflect.Slice:
		// 切片形式的条件
		for i := 0; i < rv.Len(); i++ {
			// 获取切片中的每个条件项
			cond := rv.Index(i)
			fields, values, err := getStructFieldsAndValues(cond)
			if err != nil {
				exception.PrintError(buildQueryConditions, err)
				return nil, nil, err
			}
			conditionFields = append(conditionFields, fields...)
			conditionValues = append(conditionValues, values...)
		}

	default:
		exception.PrintError(buildQueryConditions, fmt.Errorf("conditions must be a struct or slice of structs"))
		return nil, nil, fmt.Errorf("conditions must be a struct or slice of structs")
	}

	return conditionFields, conditionValues, nil
}

// getStructFieldsAndValues 获取结构体的字段和对应的值。
// 该函数会根据结构体字段名和对应的值构建查询条件，并检查 db 标签。
//
// 参数:
//   - v reflect.Value: 结构体值。
//
// 返回:
//   - []string: 结构体字段对应的数据库列名列表。
//   - []interface{}: 结构体字段的实际值列表。
//   - error: 如果字段名不符合要求，则返回错误。
func getStructFieldsAndValues(v reflect.Value) ([]string, []interface{}, error) {
	var fields []string
	var values []interface{}

	// 获取结构体类型
	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := t.Field(i)

		// 跳过匿名字段
		if fieldType.Anonymous {
			continue
		}

		// 获取 db 标签
		dbTag := fieldType.Tag.Get("db")
		if dbTag == "" {
			// 如果没有 db 标签，则使用结构体字段名，并将驼峰命名法转换为蛇形命名法
			dbTag = toSnakeCase(fieldType.Name)

			// 检查转换后的蛇形命名是否符合规则
			if !isSnakeCase(dbTag) {
				exception.PrintError(getStructFieldsAndValues, fmt.Errorf("field name '%s' is not in snake_case format", fieldType.Name))
				return nil, nil, fmt.Errorf("field name '%s' is not in snake_case format", fieldType.Name)
			}
		}

		// 如果值为空且字段标签为 "-"，跳过该字段
		if dbTag == "-" {
			continue
		}

		// 构建字段和值
		fields = append(fields, dbTag)
		values = append(values, field.Interface())
	}

	if len(fields) == 0 {
		exception.PrintError(getStructFieldsAndValues, fmt.Errorf("no valid fields found in struct %s", t.Name()))
		return nil, nil, fmt.Errorf("no valid fields found in struct %s", t.Name())
	}

	return fields, values, nil
}

// containsUnsafeSQL 检查 SELECT 查询是否包含潜在的危险 SQL 内容。
// 该函数仅检查 SELECT 查询中可能会被恶意构造的部分，避免执行删除、更新、插入等危险操作。
// 例如，检查是否存在 "DROP TABLE"、"DELETE FROM" 这种潜在的危险操作。
//
// Parameters:
//   - sqlQuery string: SQL 查询语句。
//
// Returns:
//   - bool: 如果 SQL 查询包含潜在的危险操作，返回 true，否则返回 false。
func containsUnsafeSQL(sqlQuery string) bool {
	// 只检查 SELECT 查询中的潜在危险 SQL 关键字
	unsafePatterns := []string{
		"DROP TABLE",    // 删除表
		"DELETE FROM",   // 删除数据
		"UPDATE",        // 更新数据
		"INSERT INTO",   // 插入数据
		"ALTER TABLE",   // 修改表结构
		"TRUNCATE",      // 清空表数据
		"DROP DATABASE", // 删除数据库
		";",             // SQL 语句分隔符，可能导致 SQL 注入链
		"/*",            // SQL 注释
		"*/",            // SQL 注释
		"--",            // SQL 行注释
	}

	// 转换为大写，避免大小写问题
	sqlQueryUpper := strings.ToUpper(sqlQuery)

	// 检查 SQL 查询是否包含任何不安全的模式
	for _, pattern := range unsafePatterns {
		if strings.Contains(sqlQueryUpper, pattern) {
			fmt.Println("SQL 查询包含不安全的内容：", sqlQuery)
			fmt.Println("此行为被containsUnsafeSQL函数过滤，如需关闭，请前往此函数修改")
			return true
		}
	}

	// 如果没有发现任何不安全的模式，返回 false
	return false
}

// getStructFieldsWithNonEmptyValues 获取非空值对应的字段名。
func getStructFieldsWithNonEmptyValues(v reflect.Value) ([]string, error) {
	var fields []string

	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	typeOfStruct := v.Type()
	for i := 0; i < typeOfStruct.NumField(); i++ {
		field := typeOfStruct.Field(i)
		fieldValue := v.Field(i)

		// 跳过匿名字段
		if field.Anonymous {
			continue
		}

		// 获取 db 标签的值
		dbTag := field.Tag.Get("db")
		if dbTag == "" {
			// 如果没有 db 标签，则使用结构体字段名，并将驼峰命名法转换为蛇形命名法
			dbTag = toSnakeCase(field.Name)

			// 检查转换后的蛇形命名是否符合规则
			if !isSnakeCase(dbTag) {
				exception.PrintError(getStructFieldsWithNonEmptyValues, fmt.Errorf("field name '%s' is not in snake_case format", field.Name))
				return nil, fmt.Errorf("field name '%s' is not in snake_case format", field.Name)
			}
		}

		// 跳过被标记为 "-" 的字段
		if dbTag == "-" {
			continue
		}

		// 跳过零值字段
		if isEmptyValue(fieldValue) {
			continue
		}

		fields = append(fields, dbTag)
	}

	if len(fields) == 0 {
		exception.PrintError(getStructFieldsWithNonEmptyValues, fmt.Errorf("no valid fields with non-empty values found in struct %s", typeOfStruct.Name()))
		return nil, fmt.Errorf("no valid fields with non-empty values found in struct %s", typeOfStruct.Name())
	}

	return fields, nil
}

// buildInsertPlaceholdersAndValuesSkippingEmpty 构建 SQL 插入语句的占位符和非空字段值。
func buildInsertPlaceholdersAndValuesSkippingEmpty(v reflect.Value) (string, []interface{}) {
	var placeholders []string
	var values []interface{}

	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	for i := 0; i < v.NumField(); i++ {
		fieldValue := v.Field(i)
		if isEmptyValue(fieldValue) {
			continue
		}
		placeholders = append(placeholders, "?")
		values = append(values, fieldValue.Interface())
	}

	return strings.Join(placeholders, ","), values
}

// isEmptyValue 检查字段值是否为空值。
func isEmptyValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.String:
		return v.Len() == 0
	case reflect.Bool:
		return false // 布尔值永远不会被视为 "空值"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return false // 整型即使为 0 也不是 "空值"
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return false // 无符号整型即使为 0 也不是 "空值"
	case reflect.Float32, reflect.Float64:
		return false // 浮点数即使为 0 也不是 "空值"
	case reflect.Interface, reflect.Ptr, reflect.Slice, reflect.Map:
		return v.IsNil() // 引用类型为 nil 才算空值
	default:
		return false
	}
}

// 获取对应的数据库连接
func getConn(role config.Role, db **sqlx.DB) error {
	switch role {
	case config.RoleAdmin:
		*db = db1
	case config.RolePassenger:
		*db = db2
	case config.RoleDriver:
		*db = db3
	default:
		exception.PrintError(ExecuteSQL, fmt.Errorf("role is a iota enum data structure in identity.go\n and you provide a wrong value, please check it"))
		return fmt.Errorf("role is a iota enum data structure in identity.go\n and you provide a wrong value, please check it")
	}
	return nil
}

// 检查参数
func checkArgs(statement string, args []interface{}) error {
	count := strings.Count(statement, "?")
	if count != len(args) {
		exception.PrintError(checkArgs, fmt.Errorf("参数数量不匹配，SQL 语句中有 %d 个占位符，但传入了 %d 个参数", count, len(args)))
		return fmt.Errorf("参数数量不匹配，SQL 语句中有 %d 个占位符，但传入了 %d 个参数", count, len(args))
	}
	if count == 0 {
		exception.PrintWarning(checkArgs, fmt.Errorf("SQL 语句中没有占位符，请检查："+statement))
	}
	return nil
}
