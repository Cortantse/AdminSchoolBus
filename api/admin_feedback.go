package api

import (
	"database/sql"
	"encoding/json"
	"login/config"
	"login/db"
	"login/exception"
	"login/utils"
	"net/http"
	"strings"
	"time"
)

type Feedback struct {
	FeedbackId      int     `json:"feedback_id"`      //
	UserID          int     `json:"user_id"`          //
	StudentNumber   int     `json:"student_number"`   //
	Contact         string  `json:"contact"`          //
	TotalSpending   float64 `json:"total_spending"`   //
	DriverId        string  `json:"driver_id"`        //
	VehicleNumber   string  `json:"vehicle_number"`   //
	OrderTime       string  `json:"order_time"`       //
	FeedbackContent string  `json:"feedback_content"` //
	Rating          int     `json:"rating"`           //
	Priority        bool    `json:"priority"`         //
}

// 对司机的投诉信息，是一个string(司机id) -》 []string的映射
var ComplaintToDriver map[string][]string

// GetComplaintCorrespondingToDriver 用来获取司机的投诉信息
// 接收前端发送的driver_id
// 返回一个response结构体
func GetComplaintCorrespondingToDriver(w http.ResponseWriter, r *http.Request) {
	// 需要司机端发送driver_id
	var driverID string
	err := json.NewDecoder(r.Body).Decode(&driverID)
	if err != nil {
		exception.PrintError(GetComplaintCorrespondingToDriver, err)
		return
	}

	// 获取投诉信息
	var complaintArray []string
	complaintArray = ComplaintToDriver[driverID]

	type Response struct {
		ComplaintArray []string `json:"complaint_array"`
		IsEmpty        bool     `json:"is_empty"`
	}

	var response Response
	response.ComplaintArray = complaintArray
	if len(complaintArray) == 0 {
		response.IsEmpty = true
	} else {
		response.IsEmpty = false
	}

	// 返回
	json.NewEncoder(w).Encode(response)

}

// 将给司机的投诉内容存在一个消息队列中
func temporarySaveMessagesToDrivers(driverID string, complaintContent string) {
	// 运行一个消息队列，实际是map

	if ComplaintToDriver[driverID] == nil {
		ComplaintToDriver[driverID] = []string{complaintContent}
	} else {
		ComplaintToDriver[driverID] = append(ComplaintToDriver[driverID], complaintContent)
	}
}

// GetFeedBack 用来返回所有的反馈信息
// w http.ResponseWriter, r *http.Request
func GetFeedBack(w http.ResponseWriter, r *http.Request) {
	// 暂时关闭占位符警告
	config.AllowWarning = false

	// 获取userID
	sqlStatement := `SELECT sc.user_id, st.student_account FROM
             (
                 SELECT * FROM schoolbus.usersaliases
             ) AS sc,
             (
                 SELECT * FROM passenger_db.student_information
             ) AS st
            WHERE sc.user_name = st.student_account`

	result, err := db.ExecuteSQL(config.RolePassenger, sqlStatement)
	if err != nil {
		exception.PrintError(GetFeedBack, err)
		return
	}
	rows := result.(*sql.Rows)
	// 创建快速映射
	var accountToID map[string]int
	accountToID = make(map[string]int)
	for rows.Next() {
		var userID int
		var studentAccount string
		err = rows.Scan(&userID, &studentAccount)
		if err != nil {
			exception.PrintError(GetFeedBack, err)
			return
		}
		accountToID[studentAccount] = userID
	}
	// 关闭资源
	defer rows.Close()

	// 获取所有的反馈信息

	sqlStatement = `SELECT fe.feedback_id, stu.student_number, stu.phone, ord.driver_id, pa.vehicle_id,
			  			    pa.payment_time, fe.feedback_content, fe.rating, stu.student_account, SUM(pa.payment_amount)
					 FROM feedback fe
					 JOIN student_information stu ON fe.student_number = stu.student_number
					 JOIN order_information ord ON stu.student_account = ord.student_account
					 JOIN payment_record pa ON ord.order_id = pa.order_id
					 WHERE pa.payment_status = '1'
					 GROUP BY fe.feedback_id, stu.student_number, stu.phone, ord.driver_id, pa.vehicle_id,
							 pa.payment_time, fe.feedback_content, fe.rating, stu.student_account`
	result, err = db.ExecuteSQL(config.RolePassenger, sqlStatement)

	if err != nil {
		exception.PrintError(GetFeedBack, err)
		return
	}
	// feedbacks as map
	var feedbacks map[int]*Feedback
	feedbacks = make(map[int]*Feedback)

	rows = result.(*sql.Rows)
	for rows.Next() {
		var feedback Feedback
		var stuAccount string
		err = rows.Scan(&feedback.FeedbackId, &feedback.StudentNumber, &feedback.Contact, &feedback.DriverId, &feedback.VehicleNumber, &feedback.OrderTime, &feedback.FeedbackContent, &feedback.Rating, &stuAccount, &feedback.TotalSpending)
		if err != nil {
			exception.PrintError(GetFeedBack, err)
			return
		}
		// 获取userID
		feedback.UserID = accountToID[stuAccount]
		// 方便后续插入
		feedbacks[feedback.FeedbackId] = &feedback
	}

	defer rows.Close()

	// 恢复警告
	config.AllowWarning = true

	// 计算优先级
	for _, feedback := range feedbacks {
		feedback.Priority = calPriority(feedback)
	}
	// drop data if contain both two words
	filterWords := []string{"<complaintHandled>", "<couponIssued>"}

	// 返回一个数组
	var feedbackArray []Feedback
	for _, feedback := range feedbacks {
		if strings.Contains(feedback.FeedbackContent, filterWords[0]) && strings.Contains(feedback.FeedbackContent, filterWords[1]) {
			continue
		}
		feedbackArray = append(feedbackArray, *feedback)
	}
	json.NewEncoder(w).Encode(feedbackArray)
}

// calPriority 用来计算优先级，后续考虑增加更加复杂的机制
func calPriority(feedbackPointer *Feedback) bool {
	feedbackCopy := *feedbackPointer
	// 执行运算逻辑
	if feedbackCopy.Rating < 2 {
		return true
	}
	return false
}

type DealWithFeedbackRequest struct {
	Type       string      `json:"type"`
	FeedbackId interface{} `json:"feedback_id"`
	Complaint  string      `json:"complaint"`
	DriverID   string      `json:"driver_id"`
}

// DealWithFeedback 用来处理反馈信息
func DealWithFeedback(w http.ResponseWriter, r *http.Request) {
	var request DealWithFeedbackRequest
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		exception.PrintError(DealWithFeedback, err)
		return
	}

	if request.Type == "coupon" {
		// TO DO 发放优惠券

		// 发放完毕
		sqlStatement := `UPDATE feedback SET feedback_content = CONCAT('<couponIssued>', feedback_content) WHERE feedback_id = ?`
		_, err = db.ExecuteSQL(config.RolePassenger, sqlStatement, request.FeedbackId)
		if err != nil {
			exception.PrintError(DealWithFeedback, err)
			return
		}
		// 发放，获取studentID
		sqlStatement = `SELECT feedback.student_number from feedback where feedback_id = ?`
		result, err := db.ExecuteSQL(config.RolePassenger, sqlStatement, request.FeedbackId)
		if err != nil {
			exception.PrintError(DealWithFeedback, err)
			return
		}
		rows := result.(*sql.Rows)
		var studentNumber int
		if rows.Next() {
			err = rows.Scan(&studentNumber)
			if err != nil {
				exception.PrintError(DealWithFeedback, err)
				return
			}
		}

		// 根据天发放coupon
		sqlStatement = db.ConstructInsertSQL("ride_coupon", []string{"student_number", "expiry_date", "use_status"})

		// 计算天数
		expirDate := utils.AddTime(0, 0, 0, config.AppConfig.Other.ExpirationRideCoupon)
		// 只需要到天
		expirDateStr := expirDate.Format("2006-01-02")

		_, err = db.ExecuteSQL(config.RolePassenger, sqlStatement, studentNumber, expirDateStr, "0")
		if err != nil {
			exception.PrintError(DealWithFeedback, err)
			return
		}
	} else if request.Type == "complaint" {
		// 如果是投诉，则把投诉内容存在消息队列中
		temporarySaveMessagesToDrivers(request.DriverID, request.Complaint)

		// 处理完毕
		sqlStatement := `UPDATE feedback SET feedback_content = CONCAT('<complaintHandled>', feedback_content) WHERE feedback_id = ?`
		_, err = db.ExecuteSQL(config.RolePassenger, sqlStatement, request.FeedbackId)
		if err != nil {
			exception.PrintError(DealWithFeedback, err)
			return
		}

		nowTime, _ := utils.RegularizeTimeForMySQL(time.Now().String())

		_, err := db.ExecuteSQL(config.RolePassenger, "INSERT INTO passenger_comment (student_name, comment_content, comment_time, avatar) VALUES (?,?,?,?)",
			"admin", request.Complaint, nowTime, "/uploads/avatars/avatar_9_1736861444.gif")
		if err != nil {
			exception.PrintError(DealWithFeedback, err)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
}
