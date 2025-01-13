package api

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"gopkg.in/yaml.v3"
	"login/auth"
	"login/config"
	"login/db"
	"login/exception"
	"login/utils"
	"net/http"
	"os"
	"strconv"
	"time"
)

func UpdateVariable(w http.ResponseWriter, r *http.Request) {
	// 获取用户请求数据
	type updateVariableRequest struct {
		Variable_name  string `json:"variable_name"`
		Variable_value string `json:"variable_value"`
		Token          string `json:"token"`
	}

	var request updateVariableRequest

	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		// 无法解析
		w.WriteHeader(http.StatusBadRequest)
		exception.PrintError(ChangeDataRequest, err)
		return
	}

	// 验证权限
	_, role, err := auth.VerifyAToken(request.Token)
	if err != nil {
		exception.PrintError(ChangeDataRequest, err)
		// 无权限
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// 判断权限
	if role != config.RoleAdmin {
		exception.PrintError(ChangeDataRequest, err)
		// 无权限
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// 从token获取user_id
	user_id, err := auth.ReturnUserIDFromToken(request.Token)
	if err != nil {
		exception.PrintError(ChangeDataRequest, err)
		return
	}
	//updates := map[string]int{
	//	"expiration_hours_passenger": 2,
	//	"expiration_hours_admin":     8888,
	//}

	update := map[string]string{
		request.Variable_name: request.Variable_value,
	}
	// 写入配置文件
	err = UpdateConfig(user_id, update)
	if err != nil {
		return
	}
}

// UpdateConfig 更新配置文件中的参数.
func UpdateConfig(userId string, updates map[string]string) error {
	filename := "config.yaml"
	// 更新字段
	for key, value := range updates {
		var oldValue string
		newValue := fmt.Sprint(value)

		switch key {
		case "expiration_hours_passenger":
			valueInt, err := tryConvertToInt(value)
			if err != nil {
				exception.PrintError(UpdateConfig, err)
				return err
			}
			oldValue = fmt.Sprint(config.AppConfig.Jwt.ExpirationHoursPass)
			config.AppConfig.Jwt.ExpirationHoursPass = valueInt
			break
		case "expiration_hours_admin":
			valueInt, err := tryConvertToInt(value)
			if err != nil {
				exception.PrintError(UpdateConfig, err)
				return err
			}
			oldValue = fmt.Sprint(config.AppConfig.Jwt.ExpirationHoursAdmin)
			config.AppConfig.Jwt.ExpirationHoursAdmin = valueInt
			break
		case "expiration_hours_driver":
			valueInt, err := tryConvertToInt(value)
			if err != nil {
				exception.PrintError(UpdateConfig, err)
				return err
			}
			oldValue = fmt.Sprint(config.AppConfig.Jwt.ExpirationHoursDriver)
			config.AppConfig.Jwt.ExpirationHoursDriver = valueInt
			break
		// **继续添加
		case "expiration_ride_coupon":
			valueInt, err := tryConvertToInt(value)
			if err != nil {
				exception.PrintError(UpdateConfig, err)
				return err
			}
			oldValue = fmt.Sprint(config.AppConfig.Other.ExpirationRideCoupon)
			config.AppConfig.Other.ExpirationRideCoupon = valueInt
			break
		default:
			exception.PrintError(UpdateConfig, fmt.Errorf("unknown config key: %s", key))
			return fmt.Errorf("unknown config key: %s", key)
		}

		// 保存修改记录

		if err := saveUpdateLog(userId, key, oldValue, newValue); err != nil {
			exception.PrintError(UpdateConfig, err)
			return fmt.Errorf("error saving update log: %v", err)
		}
	}

	// 保存更新后的配置到文件
	file, err := os.Create(filename)
	if err != nil {
		exception.PrintError(UpdateConfig, err)
		return fmt.Errorf("error creating config file: %v", err)
	}
	defer file.Close()

	encoder := yaml.NewEncoder(file)
	defer encoder.Close()

	if err := encoder.Encode(config.AppConfig); err != nil {
		exception.PrintError(UpdateConfig, err)
		return fmt.Errorf("error encoding config file: %v", err)
	}

	return nil
}

func tryConvertToInt(value string) (int, error) {
	value_int, err := strconv.Atoi(value)
	if err != nil {
		exception.PrintError(UpdateConfig, err)
		return 0, fmt.Errorf("err converting value to int: %v", err)
	}
	return value_int, nil
}

// 保存记录
func saveUpdateLog(id string, key string, old_value string, new_value string) error {
	sql_statement := db.ConstructInsertSQL("operations", []string{"user_id", "operation_type", "operation_time"})

	// 时间
	timeForMySQL, _ := utils.RegularizeTimeForMySQL(time.Now().String())

	// 先插入记录，不管操作内容
	result, err := db.ExecuteSQL(config.RoleAdmin, sql_statement, id, "update", timeForMySQL)
	if err != nil {
		exception.PrintError(saveUpdateLog, err)
		return fmt.Errorf("error executing sql: %v", err)
	}
	// 获取自增主键，注意类型转换
	operation_id := int(result.(int64))

	// 插入content，使用json格式
	type UpdateLog struct {
		VariableName string `json:"variable_name"`
		OldValue     string `json:"old_value"`
		NewValue     string `json:"new_value"`
	}

	update_log := UpdateLog{
		VariableName: key,
		OldValue:     old_value,
		NewValue:     new_value,
	}

	// 将json格式转换为字符串
	updateLogString, err := json.Marshal(update_log)

	// 插入content
	sql_statement = db.ConstructInsertSQL("operationscontent", []string{"operation_id", "operation_content"})
	_, err = db.ExecuteSQL(config.RoleAdmin, sql_statement, operation_id, updateLogString)
	if err != nil {
		exception.PrintError(saveUpdateLog, err)
		return fmt.Errorf("error executing sql: %v", err)
	}

	// 由于是环境变量修改，打印控制台警告修改成功，使用绿色字体，并保证恢复之后的字体
	exception.JustPrint(fmt.Sprintf("Successfully updated config variable: %s = %s", key, new_value))

	return nil
}

func GetVariable(w http.ResponseWriter, r *http.Request) {
	// 后端需要传来的数组
	// {
	//   "key": "expiration_hours_passenger",
	//   "value": "114514",
	//   "lastModified": "2024-12-20 10:00:00",
	//   "lastModifiedBy": "7"
	// }
	type Variable struct {
		Key            string `json:"key"`
		Value          string `json:"value"`
		LastModified   string `json:"lastModified"`
		LastModifiedBy string `json:"lastModifiedBy"`
	}

	allVariables := []Variable{}
	// 需要获取的所有字段，使用map ***
	variablesNames := map[string]bool{"expiration_hours_passenger": false,
		"expiration_hours_admin": false, "expiration_hours_driver": false,
		"expiration_ride_coupon": false}
	Values := map[string]string{
		"expiration_hours_passenger": fmt.Sprint(config.AppConfig.Jwt.ExpirationHoursPass),
		"expiration_hours_admin":     fmt.Sprint(config.AppConfig.Jwt.ExpirationHoursAdmin),
		"expiration_hours_driver":    fmt.Sprint(config.AppConfig.Jwt.ExpirationHoursDriver),
		"expiration_ride_coupon":     fmt.Sprint(config.AppConfig.Other.ExpirationRideCoupon),
	}
	LasModifieds := map[string]string{
		"expiration_hours_passenger": "null",
		"expiration_hours_admin":     "null",
		"expiration_hours_driver":    "null",
		"expiration_ride_coupon":     "null",
	}
	LasModifiedBys := map[string]string{
		"expiration_hours_passenger": "null",
		"expiration_hours_admin":     "null",
		"expiration_hours_driver":    "null",
		"expiration_ride_coupon":     "null",
	}

	// 从数据库中找到最新的，从而获取其他信息
	type UpdateLog struct {
		VariableName string `json:"variable_name"`
		OldValue     string `json:"old_value"`
		NewValue     string `json:"new_value"`
	}

	// 首先找到，然后获取id，然后获取lastModified，lastModifiedBy 对应 operation_time, user_id

	// 暂时关闭warning
	config.AllowWarning = false

	sqlStatement := "SELECT o.operation_id, o.operation_content, os.operation_time, os.user_id FROM operationscontent o, operations os WHERE o.operation_id = os.operation_id AND operation_content LIKE '%variable_name%' ORDER BY o.operation_id DESC"
	result, err := db.ExecuteSQL(config.RoleAdmin, sqlStatement)
	if err != nil {
		exception.PrintError(GetVariable, err)
		return
	}
	config.AllowWarning = true

	row, _ := result.(*sql.Rows)
	for row.Next() {
		var operation_id int
		var operation_content string
		var operation_time string
		var user_id string
		err = row.Scan(&operation_id, &operation_content, &operation_time, &user_id)
		if err != nil {
			exception.PrintError(GetVariable, err)
			return
		}

		// 解析为json
		var updateLog UpdateLog
		err = json.Unmarshal([]byte(operation_content), &updateLog)
		if err != nil {
			exception.PrintError(GetVariable, err)
			return
		}

		// 如果已经找到，则跳过
		if variablesNames[updateLog.VariableName] {
			continue
		} else {
			// 未找到，写入
			variablesNames[updateLog.VariableName] = true
			LasModifieds[updateLog.VariableName] = operation_time
			LasModifiedBys[updateLog.VariableName] = user_id

		}
	}

	// 已经获得了
	for key, _ := range variablesNames {
		allVariables = append(allVariables, Variable{
			Key:            key,
			Value:          Values[key],
			LastModified:   LasModifieds[key],
			LastModifiedBy: LasModifiedBys[key],
		})
	}

	// 返回
	json.NewEncoder(w).Encode(allVariables)

}
