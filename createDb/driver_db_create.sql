CREATE TABLE car_table (
    car_id VARCHAR(8) NOT NULL,             -- 车牌号，使用 CHAR(8) 类型
    car_stime DATETIME NOT NULL,         -- 开始投入使用时间，使用 DATETIME 类型
    car_isusing TINYINT(1) NOT NULL,     -- 是否停用，使用 TINYINT(1) 类型表示布尔值
    car_isworking TINYINT(1) NOT NULL,   -- 是否正在使用，使用 TINYINT(1) 类型表示布尔值
    route_id TINYINT NOT NULL,           -- 线路编号，使用 TINYINT 类型
		car_passenger INT(3) NOT NULL,			 -- 乘客人数，使用 int（3） 类型
    PRIMARY KEY (car_id)                 -- 设置车牌号为主键
);

CREATE TABLE driver_table (
    driver_id INT(6) NOT NULL,            -- 驾驶员编号，使用 INT 类型
		driver_password VARCHAR(30),						-- 驾驶员登录密码，使用 VARCHAR (30)
    driver_name VARCHAR(15) NOT NULL,     -- 姓名，使用 VARCHAR(15) 类型
    driver_sex TINYINT(1) NOT NULL,       -- 性别，使用 TINYINT(1) 类型表示 0（女）或 1（男）
    driver_tel VARCHAR(11) NOT NULL,      -- 电话号码，使用 VARCHAR(11) 类型
    driver_wages INT(10) NOT NULL,       -- 工资，使用 INT 类型
    driver_isworking TINYINT(1) NOT NULL, -- 表示驾驶员状态，使用 TINYINT(1) 类型表示状态（0:离职，1:工作，2:休息）
    PRIMARY KEY (driver_id)               -- 驾驶员编号作为主键
);

CREATE TABLE work_table (
    work_stime DATETIME NOT NULL,         -- 开始时间，用于记录驾驶员开始工作的时间
    work_etime DATETIME NOT NULL,         -- 结束时间，用于记录驾驶员结束工作的时间
    driver_id INT(6) NOT NULL,   					-- 驾驶员编号，用于区分驾驶员
    route_id INT(3)  NOT NULL,    				-- 线路编号，用于标识车辆投入哪条线路的使用
    car_id VARCHAR(8) NOT NULL,           -- 车牌号，用于区分车辆
    remark TEXT,                          -- 意见反馈，用于存储驾驶员的意见
    record_route TEXT,                    -- 路径记录，用于存储车辆的行驶轨迹（包含时间和GPS信息）
    PRIMARY KEY (work_stime, driver_id),  -- 复合主键：开始时间和驾驶员编号
    FOREIGN KEY (driver_id) REFERENCES driver_table(driver_id), -- 驾驶员编号的外键
    FOREIGN KEY (route_id) REFERENCES route_table(route_id),    -- 线路编号的外键
    FOREIGN KEY (car_id) REFERENCES car_table(car_id)            -- 车牌号的外键
);


CREATE TABLE route_table (
    route_id INT(3)  NOT NULL,         			 				 -- 线路编号，使用 INT 类型
    route_include VARCHAR(30) NOT NULL,    					 -- 包含站点，使用 VARCHAR(30) 类型来存储有序的站点信息（站点编号）
    route_isusing BOOL NOT NULL,               			 -- 是否仍在使用，使用 BOOL 类型（True/False）
    PRIMARY KEY (route_id)                     			 -- 主键：线路编号

);

CREATE TABLE site_table (
    site_id TINYINT NOT NULL,									-- 站点编号，使用 TINYINT 类型
    site_name VARCHAR(30) NOT NULL,						-- 站点名称，使用 VARCHAR(30) 类型
    site_position POINT NOT NULL,             -- 使用 POINT 类型存储经纬度
		site_passenger INT(3) NOT NULL,						-- 站点等待人数，使用 INT(3) 类型
    site_note TEXT,
    PRIMARY KEY (site_id),
    SPATIAL INDEX (site_position)             -- 为 site_position列创建空间索引
);

CREATE TABLE fare_table (
    fare_time DATETIME NOT NULL,               			-- 记录车费时间，使用 DATETIME 类型
    route_id INT   NOT NULL,        						-- 线路编号，使用 INT 类型（1~4）
    car_id VARCHAR(8) NOT NULL,                			-- 车牌号，使用 VARCHAR(8) 类型
    driver_id INT(6) NOT NULL,           						-- 驾驶员编号，使用 INT 类型（6位数字）
    user_id INT(8) NOT NULL,             						-- 用户编号，使用 INT 类型（假设是整数类型）
    PRIMARY KEY (fare_time, user_id),          			-- 主键：时间 + 用户编号
    FOREIGN KEY (route_id) REFERENCES route_table(route_id),  		-- 外键约束：线路编号
    FOREIGN KEY (car_id) REFERENCES car_table(car_id),        		-- 外键约束：车牌号
    FOREIGN KEY (driver_id) REFERENCES driver_table(driver_id)   -- 外键约束：驾驶员编号
    -- FOREIGN KEY (user_id) REFERENCES user_table(user_id)        	-- 外键约束：用户编号（假设有 user_table 表）
);