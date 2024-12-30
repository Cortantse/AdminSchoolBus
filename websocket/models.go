package websocket

// WebSocket 消息的通用结构
type WebSocketMessage struct {
	Type           string   `json:"type"`      // 消息类型
	DriverID       string   `json:"driver_id"` // 驾驶员ID（例如GPS定位的消息）
	Location       Location `json:"location"`  // 地理位置（例如GPS定位）
	Count          int      `json:"count"`     // 付款人数（例如付款消息）
	From           Location `json:"from"`      // 起点位置（例如车辆呼叫）
	To             Location `json:"to"`        // 终点位置（例如车辆呼叫）
	From_Str       string   `json:"from_str"`  // 起点位置(名称)
	To_Str         string   `json:"to_str"`    // 终点位置(名称)
	Status         string   `json:"status"`
	Time           string   `json:"time"`
	PassengerID    string   `json:"passenger_id"`
	CarID          string   `json:"car_id"`
	BoardingCount  int      `json:"boardingCount"`
	AlightingCount int      `json:"alightingCount"`
	Sites          []Site   `json:"sites"`
	Routes         []Route  `json:"routes"`
}

// 地理位置结构体
type Location struct {
	Latitude  float64 `json:"latitude"`  // 纬度
	Longitude float64 `json:"longitude"` // 经度
}

type Site struct {
	ID            int      `json:"id"`
	Name          string   `json:"name"`
	Location      Location `json:"location"` // 地理位置（例如GPS定位）
	SitePassenger int      `json:"site_passenger"`
	IsUsed        int      `json:"is_used"`
	Note          string   `json:"site_note"`
}

type Route struct {
	ID   int         `json:"id"`
	Path [][]float64 `json:"path"` // 假设路径是一个二维数组，例如 [[lng, lat], [lng, lat]]
}
