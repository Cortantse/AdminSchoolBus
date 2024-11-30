/*
 Navicat Premium Data Transfer

 Source Server         : defaultConnection
 Source Server Type    : MySQL
 Source Server Version : 90001
 Source Host           : localhost:3306
 Source Schema         : schoolbus

 Target Server Type    : MySQL
 Target Server Version : 90001
 File Encoding         : 65001

 Date: 30/11/2024 16:24:33
*/

SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

-- ----------------------------
-- Table structure for loginsessions
-- ----------------------------
DROP TABLE IF EXISTS `loginsessions`;
CREATE TABLE `loginsessions`  (
  `login_session_id` int NOT NULL AUTO_INCREMENT,
  `login_status` binary(1) NOT NULL,
  `login_time` timestamp NOT NULL,
  `login_ip_address` varchar(46) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL,
  `user_id` int NOT NULL,
  `token_id` int NOT NULL,
  PRIMARY KEY (`login_session_id`) USING BTREE,
  INDEX `main7`(`user_id` ASC) USING BTREE,
  INDEX `tk`(`token_id` ASC) USING BTREE,
  CONSTRAINT `main7` FOREIGN KEY (`user_id`) REFERENCES `userspass` (`user_id`) ON DELETE CASCADE ON UPDATE CASCADE,
  CONSTRAINT `tk` FOREIGN KEY (`token_id`) REFERENCES `tokens` (`token_id`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE = InnoDB AUTO_INCREMENT = 1 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of loginsessions
-- ----------------------------

-- ----------------------------
-- Table structure for operations
-- ----------------------------
DROP TABLE IF EXISTS `operations`;
CREATE TABLE `operations`  (
  `operation_id` int NOT NULL AUTO_INCREMENT,
  `operation_type` enum('create','update','delete','query') CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL,
  `operation_time` timestamp NOT NULL,
  `user_id` int NOT NULL,
  PRIMARY KEY (`operation_id`) USING BTREE,
  INDEX `main6`(`user_id` ASC) USING BTREE,
  CONSTRAINT `main6` FOREIGN KEY (`user_id`) REFERENCES `userspass` (`user_id`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE = InnoDB AUTO_INCREMENT = 1 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of operations
-- ----------------------------

-- ----------------------------
-- Table structure for operationscontent
-- ----------------------------
DROP TABLE IF EXISTS `operationscontent`;
CREATE TABLE `operationscontent`  (
  `operation_id` int NOT NULL,
  `operation_content` varchar(512) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL,
  PRIMARY KEY (`operation_id`) USING BTREE,
  CONSTRAINT `op` FOREIGN KEY (`operation_id`) REFERENCES `operations` (`operation_id`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of operationscontent
-- ----------------------------

-- ----------------------------
-- Table structure for tokens
-- ----------------------------
DROP TABLE IF EXISTS `tokens`;
CREATE TABLE `tokens`  (
  `token_id` int NOT NULL AUTO_INCREMENT,
  `token_hash` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL,
  `token_revoked` binary(1) NOT NULL,
  `token_expiry` datetime NOT NULL,
  `user_id` int NOT NULL,
  PRIMARY KEY (`token_id`) USING BTREE,
  INDEX `main88`(`user_id` ASC) USING BTREE,
  CONSTRAINT `main88` FOREIGN KEY (`user_id`) REFERENCES `userspass` (`user_id`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE = InnoDB AUTO_INCREMENT = 47 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of tokens
-- ----------------------------

-- ----------------------------
-- Table structure for tokensdetails
-- ----------------------------
DROP TABLE IF EXISTS `tokensdetails`;
CREATE TABLE `tokensdetails`  (
  `token_id` int NOT NULL,
  `token_created_at` timestamp NOT NULL,
  `token_client` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL,
  PRIMARY KEY (`token_id`) USING BTREE,
  CONSTRAINT `tk2` FOREIGN KEY (`token_id`) REFERENCES `tokens` (`token_id`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of tokensdetails
-- ----------------------------

-- ----------------------------
-- Table structure for usersaliases
-- ----------------------------
DROP TABLE IF EXISTS `usersaliases`;
CREATE TABLE `usersaliases`  (
  `user_name` varchar(30) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL,
  `user_id` int NULL DEFAULT NULL,
  PRIMARY KEY (`user_name`) USING BTREE,
  INDEX `main3`(`user_id` ASC) USING BTREE,
  CONSTRAINT `main3` FOREIGN KEY (`user_id`) REFERENCES `userspass` (`user_id`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of usersaliases
-- ----------------------------
INSERT INTO `usersaliases` VALUES ('xuzhj33@mail2.sysu.edu.cn', 7);
INSERT INTO `usersaliases` VALUES ('xzj', 7);
INSERT INTO `usersaliases` VALUES ('425094746@qq.com', 8);
INSERT INTO `usersaliases` VALUES ('sgw', 8);
INSERT INTO `usersaliases` VALUES ('2634315536@qq.com', 9);
INSERT INTO `usersaliases` VALUES ('ws', 9);

-- ----------------------------
-- Table structure for usersinfo
-- ----------------------------
DROP TABLE IF EXISTS `usersinfo`;
CREATE TABLE `usersinfo`  (
  `user_id` int NOT NULL,
  `user_registry_date` timestamp NOT NULL,
  `user_profile` json NULL,
  `user_avater_path` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL,
  PRIMARY KEY (`user_id`) USING BTREE,
  CONSTRAINT `main1` FOREIGN KEY (`user_id`) REFERENCES `userspass` (`user_id`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of usersinfo
-- ----------------------------
INSERT INTO `usersinfo` VALUES (7, '2024-11-30 15:43:04', NULL, NULL);
INSERT INTO `usersinfo` VALUES (8, '2024-11-30 15:44:17', NULL, NULL);
INSERT INTO `usersinfo` VALUES (9, '2024-11-30 15:45:33', NULL, NULL);

-- ----------------------------
-- Table structure for userslocked
-- ----------------------------
DROP TABLE IF EXISTS `userslocked`;
CREATE TABLE `userslocked`  (
  `user_id` int NOT NULL,
  `user_locked_time` timestamp NOT NULL,
  PRIMARY KEY (`user_id`) USING BTREE,
  CONSTRAINT `main` FOREIGN KEY (`user_id`) REFERENCES `userspass` (`user_id`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of userslocked
-- ----------------------------

-- ----------------------------
-- Table structure for userspass
-- ----------------------------
DROP TABLE IF EXISTS `userspass`;
CREATE TABLE `userspass`  (
  `user_id` int NOT NULL AUTO_INCREMENT COMMENT '从1xxxx开始，还未实现',
  `user_password_hash` varchar(60) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL,
  `user_type` tinyint(1) NOT NULL COMMENT '0（admin），1（passenger），2（driver） \n',
  `user_status` enum('active','locked','suspended','deleted') CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '\'active\'（正常）、\'locked\'（登录异常）、\'suspended\'（注销等待期）、\'deleted\'（永久不可使用） \n',
  PRIMARY KEY (`user_id`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 4 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci MAX_ROWS = 99999999 MIN_ROWS = 10000000 ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of userspass
-- ----------------------------
INSERT INTO `userspass` VALUES (7, '123456', 0, 'active');
INSERT INTO `userspass` VALUES (8, '123456', 2, 'active');
INSERT INTO `userspass` VALUES (9, '123456', 1, 'active');

-- ----------------------------
-- Table structure for userspermissions
-- ----------------------------
DROP TABLE IF EXISTS `userspermissions`;
CREATE TABLE `userspermissions`  (
  `user_id` int NOT NULL,
  `user_permission` json NULL,
  PRIMARY KEY (`user_id`) USING BTREE,
  CONSTRAINT `main4` FOREIGN KEY (`user_id`) REFERENCES `userspass` (`user_id`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of userspermissions
-- ----------------------------

-- ----------------------------
-- Table structure for verificationcodes
-- ----------------------------
DROP TABLE IF EXISTS `verificationcodes`;
CREATE TABLE `verificationcodes`  (
  `verification_id` int NOT NULL,
  `verification_code_hash` varchar(60) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL,
  `verification_expiry` datetime NOT NULL,
  `user_id` int NOT NULL,
  PRIMARY KEY (`verification_id`) USING BTREE,
  INDEX `main5`(`user_id` ASC) USING BTREE,
  CONSTRAINT `main5` FOREIGN KEY (`user_id`) REFERENCES `userspass` (`user_id`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of verificationcodes
-- ----------------------------

SET FOREIGN_KEY_CHECKS = 1;
