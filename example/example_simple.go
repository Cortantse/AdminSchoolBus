package example

import (
	"fmt"
	"login/config"
	"login/db"
	"login/exception"
)

// example_complicated.go 中提供了两个无需写sql的函数方法，实现了Insert和Select

// 此文件主要提供例子如何使用 **手动写sql的** 函数调用

// 无论是写sql/还是不写sql 的函数，我都会维护，请您根据需求选择，有bug随时滴滴我，这几个模块都是赶工出来的，难免有bug，**望谅解**

// 目前db_api只支持Insert, Update, Delete, Select 常用的sql函数，如果您有**大量**的其它需求，
// 如View视图，事务管理（这个我后面会加上），请联系我

// 对于您可能是用的这四个函数以外的函数，如果您使用的频率很低，我也提供了 SQLUnsafe 函数，不做任何安全处理，请您** 谨慎 ** 调用
// 其使用很简单，直接传入sql语句，甚至不防止sql注入

func InsertData() {
	// 1、写带占位符的sql语句，防止sql注入
	sql := "INSERT INTO userspass (user_id, user_password_hash, user_type, user_status) VALUES (?, ?, ?, ?)"

	// 2、第一个返回值获得result，这里因为是插入，没有result，以及err
	// 调用db.ExecuteSQL来运行sql语句，第一个填入了Admin这个身份，代表访问admin的数据库连接
	// 如果是driver的数据库，就换一个role就行
	// 最后直接传入参数就行，有多少个传入多少个
	_, err := db.ExecuteSQL(config.RoleAdmin, sql, 3, "123456", 1, "active")
	if err != nil {
		// 这里使用了自定义错误处理函数PrintError，方便进行打印错误，**也建议你在接收到任何error的时候使用这个函
		// 传入的的第一个函数是当前出错函数的名字，到时候出错会打印哪里出错了，然后传入error错误信息
		exception.PrintError(InsertData, err)
		return
	}
}

// 其它也类似，这个函数通用Insert, Delete, Update, 以及Select等

// 那如果我们想进行其它sql操作，比如View视图，事务管理，删除表，锁表（多线程环境），那该怎么办
// 我们因此提供通用sql处理函数，但是其不适用占位符，不预防sql注入，请您自行解决可能的安全问题

// 具体例子可以去看db_api.go

func SQLUnsafe() {
	sqlTruncate := "TRUNCATE TABLE userspass"
	result, err := db.UnSafeExecuteSQL(config.RoleAdmin, sqlTruncate)
	if err != nil {
		fmt.Println("清空表数据失败:", err)
	} else {
		fmt.Println("表数据已清空:", result)
	}
}
