package db

import (
	"fmt"
	_ "github.com/go-sql-driver/mysql" // 引入MySQL驱动
	"github.com/jmoiron/sqlx"
	_ "log"
	"login/config" // 引入config包
	"login/exception"
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
		dbName = config.AppConfig.DBNames.PassengerDB
		db = &db2
	} else if chooseDB == config.RoleDriver {
		dbName = config.AppConfig.DBNames.DriverDB
		db = &db3
	} else {
		exception.PrintError(InitDB, fmt.Errorf("chooseDB is a iota enum data structure in identity.go\n and you provide a wrong value, please check it"))
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
		exception.PrintError(InitDB, fmt.Errorf("failed to connect to database: %v", err))
		return fmt.Errorf("failed to connect to database: %v", err)
	}

	// 验证数据库连接是否有效
	err = (*db).Ping()
	if err != nil {
		exception.PrintError(InitDB, fmt.Errorf("failed to ping database: %v", err))
		return fmt.Errorf("failed to ping database: %v", err)
	}

	return nil
}

// ExecuteSQL 执行通用的 SQL 语句，支持 SELECT、INSERT、UPDATE、DELETE 等。
// 参数：
//
//	role: 选择数据库
//	sqlStatement: 带占位符的 SQL 语句
//	args: SQL 语句中占位符的对应参数
//
// 返回值：
//
//	result: 执行结果，若是查询，则为 sql.Rows 或 sql.Result
//	err: 错误信息，若发生错误则返回
//
// 示例：插入一条新用户数据
// result, err := ExecuteSQL(role, "INSERT INTO users (name, age) VALUES (?, ?)", "Alice", 30)
func ExecuteSQL(role config.Role, sqlStatement string, args ...interface{}) (interface{}, error) {
	// 0. 获取对应的数据库连接
	var db *sqlx.DB
	err := getConn(role, &db)
	if err != nil {
		exception.PrintError(ExecuteSQL, err)
		return nil, err
	}

	// 若果相应数据库连接为空，退出

	// 0.5. 确保参数匹配
	err = checkArgs(sqlStatement, args)
	if err != nil {
		exception.PrintError(ExecuteSQL, err)
		return nil, err
	}

	// 1. 确保 SQL 语句非空
	if strings.TrimSpace(sqlStatement) == "" {
		exception.PrintError(ExecuteSQL, fmt.Errorf("SQL 语句不能为空"))
		return nil, fmt.Errorf("SQL 语句不能为空")
	}

	// 2. 判断 SQL 类型并执行对应操作
	// 判断 SELECT 操作
	if strings.HasPrefix(strings.ToUpper(sqlStatement), "SELECT") {
		rows, err := db.Query(sqlStatement, args...)
		if err != nil {
			exception.PrintError(ExecuteSQL, err)
			return nil, &DBError{"ExecuteSQL", err, sqlStatement, args}
		}
		return rows, nil
	}

	// 判断 INSERT 操作
	if strings.HasPrefix(strings.ToUpper(sqlStatement), "INSERT") {
		result, err := db.Exec(sqlStatement, args...)
		if err != nil {
			exception.PrintError(ExecuteSQL, err)
			return nil, &DBError{"ExecuteSQL", err, sqlStatement, args}
		}
		// 获取自增长主键值
		lastInsertID, err := result.LastInsertId()
		if err != nil {
			exception.PrintError(ExecuteSQL, err)
			return nil, &DBError{"ExecuteSQL", err, sqlStatement, args}
		}
		return lastInsertID, nil
	}

	// 判断 UPDATE 操作
	if strings.HasPrefix(strings.ToUpper(sqlStatement), "UPDATE") {
		result, err := db.Exec(sqlStatement, args...)
		if err != nil {
			exception.PrintError(ExecuteSQL, err)
			return nil, &DBError{"ExecuteSQL", err, sqlStatement, args}
		}
		return result, nil
	}

	// 判断 DELETE 操作
	if strings.HasPrefix(strings.ToUpper(sqlStatement), "DELETE") {
		result, err := db.Exec(sqlStatement, args...)
		if err != nil {
			exception.PrintError(ExecuteSQL, err)
			return nil, &DBError{"ExecuteSQL", err, sqlStatement, args}
		}
		return result, nil
	}

	// 其它操作
	result, err := db.Exec(sqlStatement, args...)
	exception.PrintWarning(ExecuteSQL, fmt.Errorf("执行了一条不被识别的sql语句："+sqlStatement))
	if err != nil {
		exception.PrintError(ExecuteSQL, err)
		return nil, &DBError{"ExecuteSQL", err, sqlStatement, args}
	}

	return result, err
}

// UnSafeExecuteSQL 执行传入的 SQL 语句，支持常见的操作如 SELECT、INSERT、UPDATE、DELETE，
// 以及较少使用的操作如 CREATE、ALTER、DROP、TRUNCATE 等。
//
// 此函数直接执行传入的 SQL 字符串，不进行 SQL 注入防护，因此应确保 SQL 字符串来源安全，
// 避免恶意输入引发 SQL 注入攻击。
//
// 参数：
//
//	db: 数据库连接对象，使用该连接对象来执行 SQL 语句。
//	sqlStatement: 待执行的 SQL 语句字符串，可以是任何合法的 SQL 语句。
//
// 返回值：
//
//	result: 执行结果。对于 SELECT 查询，返回 *sql.Rows；对于其他操作，返回 sql.Result。
//	err: 执行过程中可能产生的错误。如果 SQL 语句为空或执行失败，返回相应的错误信息。
//
// 示例 1：创建一个新表
// sqlCreate := `
//
//	CREATE TABLE IF NOT EXISTS products (
//	    id INT PRIMARY KEY AUTO_INCREMENT,
//	    name VARCHAR(100),
//	    price DECIMAL(10, 2)
//	)
//
// `
// result, err := UnSafeExecuteSQL(role, sqlCreate)
//
//	if err != nil {
//	    fmt.Println("创建表失败:", err)
//	} else {
//
//	    fmt.Println("表创建成功:", result)
//	}
//
// 示例 2：向表中添加一个新列
// sqlAlter := "ALTER TABLE products ADD COLUMN stock INT DEFAULT 0"
// result, err = UnSafeExecuteSQL(role, sqlAlter)
//
//	if err != nil {
//	    fmt.Println("修改表结构失败:", err)
//	} else {
//
//	    fmt.Println("表结构修改成功:", result)
//	}
//
// 示例 3：删除表
// sqlDrop := "DROP TABLE IF EXISTS products"
// result, err = UnSafeExecuteSQL(role, sqlDrop)
//
//	if err != nil {
//	    fmt.Println("删除表失败:", err)
//	} else {
//
//	    fmt.Println("表删除成功:", result)
//	}
//
// 示例 4：清空表数据
// sqlTruncate := "TRUNCATE TABLE products"
// result, err = UnSafeExecuteSQL(role, sqlTruncate)
//
//	if err != nil {
//	    fmt.Println("清空表数据失败:", err)
//	} else {
//
//	    fmt.Println("表数据已清空:", result)
//	}
func UnSafeExecuteSQL(role config.Role, sqlStatement string) (interface{}, error) {
	var db *sqlx.DB
	err := getConn(role, &db)
	if err != nil {
		exception.PrintError(UnSafeExecuteSQL, err)
		return nil, err
	}

	// 确保 SQL 语句不为空
	if len(sqlStatement) == 0 {
		return nil, fmt.Errorf("SQL 语句不能为空")
	}

	// 如果是查询（SELECT）语句
	if len(sqlStatement) > 6 && sqlStatement[:6] == "SELECT" {
		rows, err := db.Query(sqlStatement)
		if err != nil {
			return nil, fmt.Errorf("查询执行失败: %v", err)
		}
		return rows, nil
	}

	// 其他操作，如 INSERT、UPDATE、DELETE、CREATE、ALTER 等
	result, err := db.Exec(sqlStatement)
	if err != nil {
		return nil, fmt.Errorf("执行失败: %v", err)
	}
	return result, nil
}

// Insert 通用插入函数，支持单条记录和批量插入。
// 根据传入的结构体或结构体切片，生成相应的 SQL 插入语句，插入数据到指定的数据库表。
// 使用警告：所有records目前必须是**相同**的结构，而非不同的结构，如果需要不同的结构插入，请分次插入！！！***
//
// Parameters:
//   - role 需要修改数据库的对应role
//   - tableName string: 目标数据库表名。
//   - records interface{}: 传入的记录，支持结构体或者结构体切片。
//
// Returns:
//   - int64: 插入的最后一条记录的 ID（单条插入时）
//   - error: 执行插入操作时可能出现的错误。
//
// example:
//   - Insert(config.RoleAdmin, "users", []User{user1, user2}) user1、2可以是{ID: 1, Name: "Alice", Age: 30}这样的
//   - Insert(config.RoleAdmin, "users", user1)
func Insert(role config.Role, tableName string, records interface{}) (int64, error) {
	rv := reflect.ValueOf(records)
	// 获取对应db连接
	var db *sqlx.DB
	err := getConn(role, &db)
	if err != nil {
		exception.PrintError(Insert, err)
		return 0, err
	}

	// 判断传入的 records 是单条数据还是切片（批量插入）
	var insertQuery string
	var values []interface{}

	switch rv.Kind() {
	case reflect.Slice:
		// 批量插入
		if rv.Len() == 0 {
			exception.PrintError(Insert, fmt.Errorf("no records to insert"))
			return 0, fmt.Errorf("no records to insert")
		}
		// ****假设所有记录有相同的字段，我们取第一条记录的字段名
		//recordType := rv.Index(0).Type()
		fields, err := getStructFieldsWithNonEmptyValues(rv.Index(0))
		if err != nil {
			exception.PrintError(Insert, fmt.Errorf("error getting struct fields: %v，请确保您的字段是蛇形", err))
			return 0, fmt.Errorf("error getting struct fields: %v，请确保您的字段是蛇形", err)
		}
		insertQuery = fmt.Sprintf("INSERT INTO %s (%s) VALUES ", tableName, strings.Join(fields, ", "))

		// 批量插入的值
		for i := 0; i < rv.Len(); i++ {
			record := rv.Index(i)
			valuePlaceholders, recordValues := buildInsertPlaceholdersAndValuesSkippingEmpty(record)
			insertQuery += fmt.Sprintf("(%s),", valuePlaceholders)
			values = append(values, recordValues...)
		}
		// 移除最后一个多余的逗号
		insertQuery = insertQuery[:len(insertQuery)-1]

	case reflect.Struct:
		// 单条插入
		fields, err := getStructFieldsWithNonEmptyValues(rv)
		if err != nil {
			exception.PrintError(Insert, fmt.Errorf("error getting struct fields: %v，请确保您的字段是蛇形", err))
			return 0, fmt.Errorf("error getting struct fields: %v，请确保您的字段是蛇形", err)
		}
		insertQuery = fmt.Sprintf("INSERT INTO %s (%s) VALUES ", tableName, strings.Join(fields, ", "))
		valuePlaceholders, recordValues := buildInsertPlaceholdersAndValuesSkippingEmpty(rv)
		insertQuery += fmt.Sprintf("(%s)", valuePlaceholders)
		values = append(values, recordValues...)

	default:
		exception.PrintError(Insert, fmt.Errorf("records must be a struct or slice of structs"))
		return 0, fmt.Errorf("records must be a struct or slice of structs")
	}

	// 执行插入操作
	result, err := db.Exec(insertQuery, values...)
	if err != nil {
		exception.PrintError(Insert, fmt.Errorf("error in Insert: executing insert: %v", err))
		return 0, fmt.Errorf("error in Insert: executing insert: %v", err)
	}

	// 获取插入数据的 ID（仅限单条插入）
	lastInsertID, _ := result.LastInsertId()
	return lastInsertID, nil
}

// SelectEasy 构造 SQL 查询语句并执行查询。
// 支持动态查询条件、排序、分页、字段筛选、分组和聚合等功能。
// Parameters:
//   - role config.Role: 需要使用的数据库角色。
//   - tableName string: 要查询的表名。
//   - dest interface{}: 用户传入的结构体类型的数组指针。
//   - if_select_all_columns bool: 是否查询所有字段，若启用此字段，下个字段不用填。
//   - columns []string: 要查询的字段，若前者为true，此字段被忽略。
//   - conditionFields []string: 查询条件字段，形如 ["age > ?", "name = ?"]。
//   - params []interface{}: 查询条件的参数。
//   - orderBy string: 排序字段及方向，如 "age DESC"。
//   - limit int: 查询的最大记录数。
//   - offset int: 偏移量，用于分页。
//   - groupBy string: 分组字段。
//   - having string: 分组后的条件。
//
// Returns:
//   - error: 执行查询时可能出现的错误。
//
// 查询结果被放回到dest数组里
func SelectEasy(
	role config.Role,
	tableName string,
	dest interface{}, // 用户传入的结构体类型的指针
	if_select_all_columns bool,
	columns []string,
	conditionFields []string,
	params []interface{},
	orderBy string,
	limit int,
	offset int,
	groupBy string,
	having string,
) error {
	// 获取对应的数据库连接
	var db *sqlx.DB
	err := getConn(role, &db)
	if err != nil {
		exception.PrintError(SelectEasy, err)
		return err
	}

	// 确保传入的 dest 是结构体的指针
	destValue := reflect.ValueOf(dest)
	if destValue.Kind() != reflect.Ptr || destValue.Elem().Kind() != reflect.Slice {
		exception.PrintError(SelectEasy, fmt.Errorf("dest must be a pointer to a slice"))
		return fmt.Errorf("dest must be a pointer to a slice")
	}

	// 构建查询的 SQL
	if len(columns) == 0 {
		columns = []string{"*"} // 默认查询所有字段
	}

	// 构建 SELECT 语句
	query := ""
	if !if_select_all_columns {
		query = fmt.Sprintf("SELECT %s FROM %s", strings.Join(columns, ", "), tableName)
	} else {
		query = fmt.Sprintf("SELECT * FROM %s", tableName)
	}

	// 构建 WHERE 子句
	if len(conditionFields) > 0 {
		query += fmt.Sprintf(" WHERE %s", strings.Join(conditionFields, " AND "))
	}

	// 构建 GROUP BY 子句
	if groupBy != "" {
		query += fmt.Sprintf(" GROUP BY %s", groupBy)
	}

	// 构建 HAVING 子句
	if having != "" {
		query += fmt.Sprintf(" HAVING %s", having)
	}

	// 构建 ORDER BY 子句
	if orderBy != "" {
		query += fmt.Sprintf(" ORDER BY %s", orderBy)
	}

	// 构建 LIMIT 和 OFFSET 子句
	if limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", limit)
	}
	if offset > 0 {
		query += fmt.Sprintf(" OFFSET %d", offset)
	}

	// 执行查询
	rows, err := db.Queryx(query, params...)
	if err != nil {
		exception.PrintError(SelectEasy, err)
		return fmt.Errorf("query error: %v", err)
	}
	defer rows.Close()

	// 确保 dest 是一个指向切片的指针
	vDest := reflect.ValueOf(dest)
	if vDest.Kind() != reflect.Ptr || vDest.Elem().Kind() != reflect.Slice {
		exception.PrintError(SelectEasy, fmt.Errorf("dest must be a pointer to a slice"))
		return fmt.Errorf("dest must be a pointer to a slice")
	}

	// 获取目标切片的元素类型
	sliceType := vDest.Elem().Type().Elem()

	// 为每一行查询结果动态构造目标切片的元素
	for rows.Next() {
		// 创建切片元素的实例
		elementPtr := reflect.New(sliceType).Interface()

		// 将查询结果扫描到切片元素中
		if err := rows.StructScan(elementPtr); err != nil {
			exception.PrintError(SelectEasy, err)
			return fmt.Errorf("error scanning row: %v", err)
		}

		// 将扫描结果追加到目标切片中
		vDest.Elem().Set(reflect.Append(vDest.Elem(), reflect.ValueOf(elementPtr).Elem()))
	}

	return nil
}

func Select(role config.Role, sqlQuery string, params []interface{}, dest interface{}) error {
	// 获取对应的数据库连接
	var db *sqlx.DB
	err := getConn(role, &db)
	if err != nil {
		exception.PrintError(Select, err)
		return err
	}

	// 检查 SQL 语句中是否包含不安全的部分
	if containsUnsafeSQL(sqlQuery) {
		exception.PrintError(Select, fmt.Errorf("SQL query contains potentially unsafe content: %s", sqlQuery))
		return fmt.Errorf("SQL query contains potentially unsafe content: %s", sqlQuery)
	}

	// 如果 params 中有切片，处理 IN 语法
	for _, param := range params {
		if reflect.TypeOf(param).Kind() == reflect.Slice {
			// 获取切片的长度
			sliceLength := reflect.ValueOf(param).Len()
			// 动态生成占位符
			placeholders := make([]string, sliceLength)
			for i := 0; i < sliceLength; i++ {
				placeholders[i] = "?"
			}
			// 替换原 SQL 中的占位符
			sqlQuery = strings.Replace(sqlQuery, "?", strings.Join(placeholders, ","), 1)
		}
	}

	// 执行 SQL 查询
	rows, err := db.Queryx(sqlQuery, params...)
	if err != nil {
		exception.PrintError(Select, err)
		return fmt.Errorf("error executing SQL query: %v", err)
	}
	defer rows.Close()

	// 确保传入的目标结构体是指针类型
	if reflect.TypeOf(dest).Kind() != reflect.Ptr {
		exception.PrintError(Select, fmt.Errorf("destination parameter must be a pointer"))
		return fmt.Errorf("destination parameter must be a pointer")
	}

	// 获取切片的类型（即目标结构体的类型）
	sliceType := reflect.TypeOf(dest).Elem()

	// 确保目标类型是一个切片类型
	if sliceType.Kind() != reflect.Slice {
		exception.PrintError(Select, fmt.Errorf("destination parameter must be a pointer to a slice"))
		return fmt.Errorf("destination parameter must be a pointer to a slice")
	}

	// 动态获取切片元素的类型（即单个结构体类型）
	elemType := sliceType.Elem()

	// 使用结构体数组指针接收结果
	for rows.Next() {
		// 创建一个新的结构体实例
		elem := reflect.New(elemType).Interface()

		// 使用 rows.StructScan 扫描每一行数据到结构体
		if err := rows.StructScan(elem); err != nil {
			exception.PrintError(Select, err)
			return fmt.Errorf("error scanning row: %v", err)
		}

		// 将扫描到的结构体添加到切片中
		reflect.ValueOf(dest).Elem().Set(reflect.Append(reflect.ValueOf(dest).Elem(), reflect.ValueOf(elem).Elem()))
	}

	if err := rows.Err(); err != nil {
		exception.PrintError(Select, err)
		return fmt.Errorf("error iterating rows: %v", err)
	}

	return nil
}

// ConstructInsertSQL 构造插入语句
// 警告：不负责插入value，而是构造占位符
func ConstructInsertSQL(tableName string, dataNames []string) string {
	// 占位符
	placeHolders := make([]string, len(dataNames))
	for i := 0; i < len(dataNames); i++ {
		placeHolders[i] = "?"
	}
	stringWithPlaceHolders := fmt.Sprintf(" %s ", strings.Join(placeHolders, ", "))

	sqlStatement := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", tableName, strings.Join(dataNames, ", "),
		stringWithPlaceHolders)

	return sqlStatement
}
