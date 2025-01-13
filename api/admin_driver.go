package api

import (
	"database/sql"
	"encoding/json"
	"login/config"
	"login/db"
	"login/exception"
	"net/http"
	"strconv"
	"strings"
)

// @Summary 获得表格数据
// @Description 根据具体内容获得数据
// @Tags admins
// @Accept json
// @Produce json
// @Param user body User true "User data"
// @Success 201 {object} User
// @Failure 400 {object} ErrorResponse
// @Router /unknownnow [post]
func GetDriversTableData(w http.ResponseWriter, r *http.Request) {
	// 定义用户结构
	type Driver struct {
		DriverID        int    `json:"driver_id"`
		DriverName      string `json:"driver_name"`
		DriverSex       int    `json:"driver_sex"`
		DriverTel       string `json:"driver_tel"`
		DriverWages     int    `json:"driver_wages"`
		DriverIsworking int    `json:"driver_isworking"`
	}

	type DriverResponse struct {
		Data  []Driver `json:"data"`
		Total int      `json:"total"`
		Page  int      `json:"page"`
		Size  int      `json:"size"`
	}

	// 获取查询参数
	keyword := r.URL.Query().Get("keyword")
	workingStatus := r.URL.Query().Get("driver_isworking")
	pageStr := r.URL.Query().Get("page")
	sizeStr := r.URL.Query().Get("size")

	var drivers []Driver

	config.AllowWarning = false

	// 查询数据库获取所有司机数据
	sqlS := "SELECT driver_id, driver_name, driver_sex, driver_tel, driver_wages, driver_isworking FROM driver_table"
	result, err := db.ExecuteSQL(config.RoleDriver, sqlS)
	if err != nil {
		exception.PrintError(GetDriversTableData, err)
		return
	}
	row := result.(*sql.Rows)
	for row.Next() {
		var driver Driver
		if err := row.Scan(&driver.DriverID, &driver.DriverName, &driver.DriverSex, &driver.DriverTel, &driver.DriverWages, &driver.DriverIsworking); err != nil {
			exception.PrintError(GetDriversTableData, err)
			return
		}
		drivers = append(drivers, driver)
	}

	config.AllowWarning = true

	// 默认分页参数
	page, _ := strconv.Atoi(pageStr)
	if page < 1 {
		page = 1
	}
	size, _ := strconv.Atoi(sizeStr)
	if size < 1 {
		size = 10
	}

	// 过滤用户数据
	var filteredDrivers []Driver
	for _, driver := range drivers {
		// 过滤关键字 (搜索 name 或 tel)
		if keyword != "" && !strings.Contains(driver.DriverName, keyword) && !strings.Contains(driver.DriverTel, keyword) {
			continue
		}
		// 过滤是否在职状态 (driver_isworking)
		if workingStatus != "" {
			statusFilter, err := strconv.Atoi(workingStatus)
			if err == nil && driver.DriverIsworking != statusFilter {
				continue
			}
		}
		filteredDrivers = append(filteredDrivers, driver)
	}

	// 分页处理
	start := (page - 1) * size
	end := start + size
	if start > len(filteredDrivers) {
		start = len(filteredDrivers)
	}
	if end > len(filteredDrivers) {
		end = len(filteredDrivers)
	}
	paginatedDrivers := filteredDrivers[start:end]

	// 构造返回数据
	response := DriverResponse{
		Data:  paginatedDrivers,
		Total: len(filteredDrivers),
		Page:  page,
		Size:  size,
	}

	// 设置响应头并返回 JSON 数据
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		exception.PrintError(GetDriversTableData, err)
	}
}

// @Summary 获得表格数据
// @Description 根据具体内容获得数据
// @Tags admins
// @Accept json
// @Produce json
// @Param car body Car true "Car data"
// @Success 201 {object} Car
// @Failure 400 {object} ErrorResponse
// @Router /car_table [post]
func GetCarsTableData(w http.ResponseWriter, r *http.Request) {
	config.AllowWarning = false
	// 定义车辆结构
	type Car struct {
		CarID        string `json:"car_id"`
		CarStime     string `json:"car_stime"`
		CarIsUsing   int    `json:"car_isusing"`
		CarIsWorking int    `json:"car_isworking"`
		RouteID      int    `json:"route_id"`
		CarPassenger int    `json:"car_passenger"`
	}

	type CarResponse struct {
		Data  []Car `json:"data"`
		Total int   `json:"total"`
		Page  int   `json:"page"`
		Size  int   `json:"size"`
	}

	// 获取查询参数
	keyword := r.URL.Query().Get("keyword")
	isUsing := r.URL.Query().Get("car_isusing")
	isWorking := r.URL.Query().Get("car_isworking")
	pageStr := r.URL.Query().Get("page")
	sizeStr := r.URL.Query().Get("size")

	var cars []Car

	// 查询数据库获取所有车辆数据
	sqlS := "SELECT car_id, car_stime, car_isusing, car_isworking, route_id, car_passenger FROM car_table"
	result, err := db.ExecuteSQL(config.RoleDriver, sqlS)
	if err != nil {
		exception.PrintError(GetCarsTableData, err)
		return
	}
	row := result.(*sql.Rows)
	for row.Next() {
		var car Car
		if err := row.Scan(&car.CarID, &car.CarStime, &car.CarIsUsing, &car.CarIsWorking, &car.RouteID, &car.CarPassenger); err != nil {
			exception.PrintError(GetCarsTableData, err)
			return
		}
		cars = append(cars, car)
	}

	defer row.Close()

	config.AllowWarning = true

	// 默认分页参数
	page, _ := strconv.Atoi(pageStr)
	if page < 1 {
		page = 1
	}
	size, _ := strconv.Atoi(sizeStr)
	if size < 1 {
		size = 10
	}

	// 过滤车辆数据
	var filteredCars []Car
	for _, car := range cars {
		// 过滤关键字 (搜索 car_id 或 route_id)
		if keyword != "" && !strings.Contains(car.CarID, keyword) && strconv.Itoa(car.RouteID) != keyword {
			continue
		}
		// 过滤是否停用 (car_isusing)
		if isUsing != "" {
			usingFilter, err := strconv.Atoi(isUsing)
			if err == nil && car.CarIsUsing != usingFilter {
				continue
			}
		}
		// 过滤是否正在使用 (car_isworking)
		if isWorking != "" {
			workingFilter, err := strconv.Atoi(isWorking)
			if err == nil && car.CarIsWorking != workingFilter {
				continue
			}
		}
		filteredCars = append(filteredCars, car)
	}

	// 分页处理
	start := (page - 1) * size
	end := start + size
	if start > len(filteredCars) {
		start = len(filteredCars)
	}
	if end > len(filteredCars) {
		end = len(filteredCars)
	}
	paginatedCars := filteredCars[start:end]

	// 构造返回数据
	response := CarResponse{
		Data:  paginatedCars,
		Total: len(filteredCars),
		Page:  page,
		Size:  size,
	}

	// 设置响应头并返回 JSON 数据
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		exception.PrintError(GetCarsTableData, err)
	}
}

// @Summary 获取 work_table 表格数据
// @Description 根据查询参数获取 work_table 的数据
// @Tags admins
// @Accept  json
// @Produce  json
// @Param keyword  query string false "搜索关键字（可搜索 driver_id, car_id, route_id 等）"
// @Param page     query int    false "当前页码，默认 1"
// @Param size     query int    false "每页条数，默认 10"
// @Success 200 {object} WorkResponse
// @Failure 400 {object} ErrorResponse
// @Router /work_table [get]
func GetWorkTableData(w http.ResponseWriter, r *http.Request) {
	config.AllowWarning = false

	// 定义 work_table 的结构体（根据实际字段类型决定是否使用 string/int/time.Time 等）
	type Work struct {
		WorkStime   string `json:"work_stime"`   // 开始时间
		WorkEtime   string `json:"work_etime"`   // 结束时间
		DriverID    int    `json:"driver_id"`    // 驾驶员编号
		RouteID     int    `json:"route_id"`     // 线路编号
		CarID       string `json:"car_id"`       // 车牌号
		Remark      string `json:"remark"`       // 意见反馈
		RecordRoute string `json:"record_route"` // 路径记录
	}

	// 定义返回给前端的结构
	type WorkResponse struct {
		Data  []Work `json:"data"`  // 当前页的记录
		Total int    `json:"total"` // 过滤后的总记录数
		Page  int    `json:"page"`  // 当前页码
		Size  int    `json:"size"`  // 每页条数
	}

	// 获取查询参数
	keyword := r.URL.Query().Get("keyword")
	pageStr := r.URL.Query().Get("page")
	sizeStr := r.URL.Query().Get("size")

	// 默认分页参数
	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}
	size, err := strconv.Atoi(sizeStr)
	if err != nil || size < 1 {
		size = 10
	}

	// 查询数据库获取所有数据（可根据实际情况拼装更合理的 SQL 语句）
	sqlS := `SELECT work_stime, work_etime, driver_id, route_id, car_id, remark, record_route FROM work_table`
	result, err := db.ExecuteSQL(config.RoleDriver, sqlS)
	if err != nil {
		exception.PrintError(GetWorkTableData, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer func() {
		if rows, ok := result.(*sql.Rows); ok {
			_ = rows.Close()
		}
	}()

	rows, _ := result.(*sql.Rows)

	// 将查询结果保存到切片
	var works []Work
	for rows.Next() {
		var wdata Work
		if err := rows.Scan(
			&wdata.WorkStime,
			&wdata.WorkEtime,
			&wdata.DriverID,
			&wdata.RouteID,
			&wdata.CarID,
			&wdata.Remark,
			&wdata.RecordRoute,
		); err != nil {
			exception.PrintError(GetWorkTableData, err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		works = append(works, wdata)
	}

	config.AllowWarning = true

	// 如果需要对查询结果进行“关键词”过滤，可以在这里实现
	// 例如 keyword 同时匹配 driver_id、route_id、car_id 等
	var filteredWorks []Work
	for _, item := range works {
		// 假设关键词可以匹配 driver_id、route_id 或 car_id（字符串包含或相等）
		// 可根据需要灵活调整
		if keyword != "" {
			// 这里示例：只要符合其中一个即认为匹配
			driverIDStr := strconv.Itoa(item.DriverID)
			routeIDStr := strconv.Itoa(item.RouteID)
			if !strings.Contains(item.CarID, keyword) &&
				!strings.Contains(driverIDStr, keyword) &&
				!strings.Contains(routeIDStr, keyword) {
				continue
			}
		}
		filteredWorks = append(filteredWorks, item)
	}

	// 分页
	total := len(filteredWorks)
	start := (page - 1) * size
	end := start + size
	if start > total {
		start = total
	}
	if end > total {
		end = total
	}

	paginatedWorks := filteredWorks[start:end]

	// 构造返回数据
	response := WorkResponse{
		Data:  paginatedWorks,
		Total: total,
		Page:  page,
		Size:  size,
	}

	// 设置响应头并返回 JSON 数据
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		exception.PrintError(GetWorkTableData, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func GiveDriverInfo(w http.ResponseWriter, r *http.Request) {
	// 提供html
	// 关于司机的名字，性别，电话，评星，注意动态生成评星
	// 获取driverID

	var htmls []string

	var driverID string
	err := json.NewDecoder(r.Body).Decode(&driverID)
	if err != nil {
		exception.PrintError(ReceiveAIRequest, err)
		return
	}

	// 查询数据库获取所有司机数据
	sqlS := "SELECT driver_name, driver_sex, driver_tel FROM driver_table WHERE driver_id = ?"
	result, err := db.ExecuteSQL(config.RoleDriver, sqlS, driverID)
	if err != nil {
		exception.PrintError(GiveDriverInfo, err)
		return
	}
	row := result.(*sql.Rows)

	var driverName string
	var driverSex int
	var driverTel string
	for row.Next() {
		if err := row.Scan(&driverName, &driverSex, &driverTel); err != nil {
			exception.PrintError(GiveDriverInfo, err)
			return
		}
	}

	driverRealSex := "男"
	if driverSex == 0 {
		driverRealSex = "女"
	}

	htmls = append(htmls, "<p>司机姓名："+driverName+"</p>")
	htmls = append(htmls, "<p>司机性别："+driverRealSex+"</p>")
	htmls = append(htmls, "<p>司机电话："+driverTel+"</p>")

	// 获取评星并动态生成
	sqlS = "SELECT AVG(f.rating) FROM passenger_db.feedback f, passenger_db.order_information o WHERE o.driver_id = ? AND f.order_id = o.order_id;"
	result, err = db.ExecuteSQL(config.RoleDriver, sqlS, driverID)
	if err != nil {
		exception.PrintError(GiveDriverInfo, err)
		return
	}
	row = result.(*sql.Rows)
	var rating float64
	for row.Next() {
		if err := row.Scan(&rating); err != nil {
			exception.PrintError(GiveDriverInfo, err)
			return
		}
	}

	defer row.Close()

	// 添加评星
	tem := "<p>司机评星："
	for i := 0; i < 5; i++ {
		if rating >= float64(i+1) {
			tem += "<img src=\"@/assets/star.jpg\" width=\"20\" height=\"20\">"
		} else {
			tem += "<img src=\"@/assets/star_gray.jpg\" width=\"20\" height=\"20\">"
		}
	}
	tem += "</p>"

	htmls = append(htmls, tem)

	type response struct {
		Htmls []string `json:"htmls"`
	}

	// 返回html
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response{Htmls: htmls}); err != nil {
		exception.PrintError(GiveDriverInfo, err)
		return
	}
}
