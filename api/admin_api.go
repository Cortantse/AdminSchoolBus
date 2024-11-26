package api

import (
	"database/sql"
	"encoding/json"
	"login/config"
	"login/db"
	"net/http"
)

// @Summary 发送dashboard需要的信息
// @Description 接收前端fetch请求，返回dashboard需要信息
// @Tags users
// @Accept json
// @Produce json
// @Param user body User true "User data"
// @Success 201 {object} User
// @Failure 400 {object} ErrorResponse
// @Router /users [post]
func GiveDashBoardInfo(w http.ResponseWriter, r *http.Request) {

	type DashBoardStatus struct {
		TotalUsers   int `json:"total_users"`
		TotalDrivers int `json:"total_drivers"`
		TotalAdmins  int `json:"total_admins"`
	}
	var totalUsers int
	var totalDrivers int
	var totalAdmins int
	// 获取当前所有数字
	resultUser, _ := db.ExecuteSQL(config.RoleAdmin, "SELECT COUNT(*) FROM userspass WHERE user_type = ?", 1)
	resultDrivers, _ := db.ExecuteSQL(config.RoleAdmin, "SELECT COUNT(*) FROM userspass WHERE user_type = ?", 2)
	resultAdmins, _ := db.ExecuteSQL(config.RoleAdmin, "SELECT COUNT(*) FROM userspass WHERE user_type = ?", 3)
	// 断言类型
	result1, _ := resultUser.(*sql.Rows)
	result2, _ := resultDrivers.(*sql.Rows)
	result3, _ := resultAdmins.(*sql.Rows)

	if result1.Next() && result2.Next() && result3.Next() {
		_ = result1.Scan(&totalUsers)
		_ = result2.Scan(&totalDrivers)
		_ = result3.Scan(&totalAdmins)
	}

	data := DashBoardStatus{
		TotalUsers:   totalUsers,
		TotalDrivers: totalDrivers,
		TotalAdmins:  totalAdmins,
	}

	// 设置响应头
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// 将数据转化为JSON格式并写入响应体
	json.NewEncoder(w).Encode(data)
}

// AnswerHeartBeat 接收心跳检测请求
func AnswerHeartBeat(w http.ResponseWriter, r *http.Request) {
	// 正常就回复200即可
	w.WriteHeader(http.StatusOK)
}
