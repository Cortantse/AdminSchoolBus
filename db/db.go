package db

import (
	"fmt"
	_ "github.com/go-sql-driver/mysql" // 引入MySQL驱动
	"github.com/jmoiron/sqlx"
	_ "log"
	"login/config" // 引入config包
	"reflect"
	"strings"
)

// 数据库实例       高度相关***注意1是admin的db连接，2是user的，3是driver
var db1 *sqlx.DB
var db2 *sqlx.DB
var db3 *sqlx.DB

// InitDB 连接指定的数据库
//
// Parameters:
//   - *sqlx.DB: 数据库连接实例
//   - chooseDB 枚举iota类型，在identity.go中
//
// Returns:
//   - error: 错误信息
func InitDB(chooseDB config.Role) error {
	// 从 config 包中获取数据库配置信息
	dbConfig := config.AppConfig.Database

	// 选择正确的
	var dbName string
	var db **sqlx.DB

	if chooseDB == config.RoleAdmin {
		dbName = config.AppConfig.DBNames.AdminDB
		db = &db1
	} else if chooseDB == config.RolePassenger {
		dbName = config.AppConfig.DBNames.AdminDB
		db = &db2
	} else if chooseDB == config.RoleDriver {
		dbName = config.AppConfig.DBNames.AdminDB
		db = &db3
	} else {
		return fmt.Errorf("chooseDB is a iota enum data structure in identity.go\n and you provide a wrong value, please check it")
	}

	// 构造数据库连接字符串（DSN）
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s",
		dbConfig.User,
		dbConfig.Password,
		dbConfig.Host,
		dbConfig.Port,
		dbName,
	)

	// 连接数据库
	var err error
	*db, err = sqlx.Connect("mysql", dsn)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %v", err)
	}

	return nil
}

// Insert 通用插入函数，支持单条记录和批量插入。
// 根据传入的结构体或结构体切片，生成相应的 SQL 插入语句，插入数据到指定的数据库表。
//
// 参数:
//   - db *sqlx.DB: 数据库连接对象。
//   - tableName string: 目标数据库表名。
//   - records interface{}: 传入的记录，支持结构体或者结构体切片。
//
// 返回:
//   - int64: 插入的最后一条记录的 ID（单条插入时）
//   - error: 执行插入操作时可能出现的错误。
func Insert(db *sqlx.DB, tableName string, records interface{}) (int64, error) {
	rv := reflect.ValueOf(records)

	// 判断传入的 records 是单条数据还是切片（批量插入）
	var insertQuery string
	var values []interface{}

	switch rv.Kind() {
	case reflect.Slice:
		// 批量插入
		if rv.Len() == 0 {
			return 0, fmt.Errorf("no records to insert")
		}
		// 假设所有记录有相同的字段，我们取第一条记录的字段名
		recordType := rv.Index(0).Type()
		fields := getStructFields(recordType)
		insertQuery = fmt.Sprintf("INSERT INTO %s (%s) VALUES ", tableName, strings.Join(fields, ", "))

		// 批量插入的值
		for i := 0; i < rv.Len(); i++ {
			record := rv.Index(i)
			valuePlaceholders, recordValues := buildInsertPlaceholdersAndValues(record)
			insertQuery += fmt.Sprintf("(%s),", valuePlaceholders)
			values = append(values, recordValues...)
		}
		// 移除最后一个多余的逗号
		insertQuery = insertQuery[:len(insertQuery)-1]

	case reflect.Struct:
		// 单条插入
		fields := getStructFields(rv.Type())
		insertQuery = fmt.Sprintf("INSERT INTO %s (%s) VALUES ", tableName, strings.Join(fields, ", "))
		valuePlaceholders, recordValues := buildInsertPlaceholdersAndValues(rv)
		insertQuery += fmt.Sprintf("(%s)", valuePlaceholders)
		values = append(values, recordValues...)

	default:
		return 0, fmt.Errorf("records must be a struct or slice of structs")
	}

	// 执行插入操作
	result, err := db.Exec(insertQuery, values...)
	if err != nil {
		return 0, fmt.Errorf("error executing insert: %v", err)
	}

	// 获取插入数据的 ID（仅限单条插入）
	lastInsertID, _ := result.LastInsertId()
	return lastInsertID, nil
}

// getStructFields 获取结构体的字段名。
// 该函数通过反射获取结构体类型的所有字段名称，并返回字段名称的字符串切片。
//
// 参数:
//   - t reflect.Type: 要获取字段名的结构体类型。
//
// 返回:
//   - []string: 结构体字段的名称列表。
func getStructFields(t reflect.Type) []string {
	var fields []string
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		fields = append(fields, field.Name)
	}
	return fields
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
