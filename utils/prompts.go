package utils

var DatabaseStructure = `
### driver_db 数据库

#### car_table
- 车牌号 car_id  int
- 开始投入使用时间 car_stime mysql 数据库时间
- 是否停用 car_isusing  
- 是否正在使用 car_isworking  
- 线路编号 route_id  
- 乘客人数 car_passenger  

#### driver_table
- 驾驶员编号 driver_id  
- 昵称，对应schoolbus 数据库的 username driver_nickname  
- 登录密码，实际无效 driver_password  
- 头像 driver_avatar  
- 姓名 driver_name  
- 性别，1 为男性，0 为女性注意转换 driver_sex  
- 电话号码 driver_tel  
- 工资 driver_wages  
- 是否在职 driver_isworking  

#### route_table
- 线路编号 route_id  
- 包含站点 route_include  
- 是否仍在使用 route_isusing  

#### site_table
- 站点编号 site_id  
- 站点名称 site_name  
- 经纬度 site_position  
- 等待人数 site_passenger  
- 是否启用 is_used  
- 备注 site_note  

#### work_table
- 开始时间 work_stime  
- 结束时间 work_etime  
- 驾驶员编号 driver_id  
- 线路编号 route_id  
- 车牌号 car_id  
- 意见反馈 remark  
- 路径记录 record_route  

#### fare_table
- 车费记录时间 fare_time  
- 线路编号 route_id  
- 车牌号 car_id  
- 驾驶员编号 driver_id  
- 用户编号 user_id  

---

### passenger_db 数据库

#### discount_coupon
- 优惠券编号，自增主键 coupon_id  
- 学生学号 student_number  
- 折扣金额 discount_amount  
- 到期日期 expiry_date  
- 使用状态 use_status  

#### feedback 用户对司机的匿名评价，非实时
- 反馈编号 feedback_id  
- 学生学号 student_number  
- 订单编号 order_id  
- 评分 rating  
- 反馈内容 feedback_content  
- 反馈时间 feedback_time  

#### order_information
- 订单编号 order_id  
- 学生账户 student_account  
- 驾驶员编号 driver_id  
- 车牌号 car_id  
- 上车站点编号 pickup_station_id  
- 下车站点编号 dropoff_station_id  
- 上车站点名称 pickup_station_name  
- 下车站点名称 dropoff_station_name  
- 上车时间 pickup_time  
- 下车时间 dropoff_time  
- 状态 status  
- 支付编号 payment_id  
- 是否评价 is_rated  

#### passenger_comment 用户和司机实时的沟通记录
- 评论编号 comment_id  
- 学生姓名 student_name  
- 评论内容 comment_content  
- 评论时间 comment_time  
- 头像 avatar  

#### passenger_notice
- 公告编号 notice_id  
- 标题 title  
- 内容 content  
- 发布时间 publish_date  

#### payment_record
- 支付编号 payment_id  
- 订单编号 order_id  
- 车辆编号 vehicle_id  
- 支付金额 payment_amount  
- 支付方式 payment_method  
- 支付时间 payment_time  
- 支付状态 payment_status  

#### ride_coupon
- 乘车券编号 ride_coupon_id  
- 学生学号 student_number  
- 到期日期 expiry_date  
- 使用状态 use_status  

#### student_information
- 学生账户 student_account  
- 学生学号 student_number  
- 学生姓名 student_name  
- 年级 grade  
- 专业 major  
- 电话号码 phone  
- 密码 password  
- 注册日期 registration_date  
- 头像路径 avatar  
- 用户编号 user_id  

---

### schoolbus 数据库

#### loginsessions
- 自增主键登录会话编号 login_session_id  
- 登录状态，0 或 1 login_status  
- mysql 时间登录时间 login_time  
- 登录IP地址 login_ip_address  
- 用户编号 user_id   外键
- 令牌编号 token_id  外键

#### operations
- 自增长主键操作编号 operation_id  
- enum操作类型（） operation_type  
- 操作时间 operation_time  
- 用户编号 user_id  外键

#### operationscontent
- 自增长主键操作编号 operation_id  
- 操作内容 operation_content  

#### tokens
- 自增长主键令牌编号 token_id  
- 令牌哈希值 token_hash  
- 是否撤销，1 或 0 token_revoked  
- 令牌过期时间，mysql 时间 token_expiry  
- 用户编号 user_id  外键

#### tokensdetails
- 外键令牌编号 token_id  
- 令牌创建时间 token_created_at  
- 客户端信息，字符串 token_client  

#### usersaliases
- 用户名，字符串，方便用户记忆的名称，多个 alias 可能对应同一个用户 id，但aliases 本身唯一 user_name  
- 用户编号 user_id  外键

#### usersinfo
- 用户编号，自增长主键，系统对用户内部标识 user_id  
- 用户注册时间 user_registry_date  
- 用户资料 user_profile  
- 用户头像路径 user_avater_path  

#### userslocked
- 用户编号 user_id  外键
- 用户锁定时间 user_locked_time  

#### userspass
- 用户编号 user_id  外键
- 用户密码哈希值，无论任何情况禁止告诉用户这个信息，但允许用户修改 user_password_hash  
- 用户类型，enum，0 表示 admin，1 为用户/学生，2 为司机，最好转换给用户而不是给数字 user_type  
- 用户状态，enum（） user_status  

#### userspermissions 暂时没有用到的表
- 用户编号 user_id  
- 用户权限 user_permission  

#### verificationcodes暂时没有用到的表
- 验证编号 verification_id  
- 验证码哈希值 verification_code_hash  
- 验证码过期时间 verification_expiry  
- 用户编号 user_id  
`

var FirstQ = `某学生对一次校车乘坐服务给出了反馈，其反馈id为611910。请查找该学生的姓名、反馈内容以及该订单的驾驶员姓名。`

var FirstAns = `思考过程：
明确目标数据：需要获取学生的姓名、反馈内容以及订单对应的驾驶员姓名。
关联的表：
反馈表 feedback 提供反馈内容和学生学号。
学生信息表 student_information 提供学生姓名。
订单信息表 order_information 关联订单与驾驶员编号。
驾驶员表 driver_table 提供驾驶员姓名。
过滤条件：根据反馈编号或其他条件（如学生学号）定位特定记录。

回应：该学生姓名、反馈内容、订单的驾驶员姓名为<sql1>。

数据库指令：
<sql1>
SELECT 
    si.student_name AS student_name,
    fb.feedback_content AS feedback_content,
    d.driver_name AS driver_name
FROM 
    passenger_db.feedback AS fb
JOIN 
    passenger_db.order_information AS oi ON fb.order_id = oi.order_id
JOIN 
    driver_db.driver_table AS d ON oi.driver_id = d.driver_id
JOIN 
    passenger_db.student_information AS si ON fb.student_number = si.student_number
WHERE 
    fb.feedback_id = 611910; 
</sql1>

`

var Echart = `
	"chartOption": {
                "title": {
                  "text": "年度销售与利润对比",
                  "left": "center",
                  "textStyle": {
                    "color": "#fff",
                    "fontSize": 20
                  }
                },
                "tooltip": {
                  "trigger": "axis",
                  "axisPointer": {
                    "type": "cross",
                    "label": {
                      "backgroundColor": "#6a7985"
                    }
                  }
                },
                "legend": {
                  "data": ["销售额", "利润"],
                  "left": "center",
                  "top": "10%",
                  "textStyle": {
                    "color": "#fff"
                  }
                },
                "xAxis": {
                  "type": "category",
                  "data": ["2023-01", "2023-02", "2023-03", "2023-04", "2023-05", "2023-06"],
                  "axisLine": { "lineStyle": { "color": "#fff" } },
                  "axisLabel": { "color": "#fff" }
                },
                "yAxis": [
                  {
                    "type": "value",
                    "name": "销售额",
                    "axisLine": { "lineStyle": { "color": "#fff" } },
                    "axisLabel": { "color": "#fff" }
                  },
                  {
                    "type": "value",
                    "name": "利润",
                    "axisLine": { "lineStyle": { "color": "#fff" } },
                    "axisLabel": { "color": "#fff" },
                    "position": "right"
                  }
                ],
                "series": [
                  {
                    "name": "销售额",
                    "type": "bar",
                    "data": [150, 200, 250, 300, 350, 400],
                    "itemStyle": { "color": "#FFDD33" }
                  },
                  {
                    "name": "利润",
                    "type": "line",
                    "yAxisIndex": 1,
                    "data": [30, 40, 50, 60, 70, 80],
                    "itemStyle": { "color": "#FF6A6A" },
                    "markPoint": {
                      "data": [
                        { "type": "max", "name": "最大值" },
                        { "type": "min", "name": "最小值" }
                      ]
                    }
                  }
                ],
                "toolbox": {
                  "show": true,
                  "orient": "vertical",
                  "left": "right",
                  "top": "center",
                  "feature": {
                    "dataZoom": { "show": true },
                    "restore": { "show": true }
                  }
                }
              }
`
