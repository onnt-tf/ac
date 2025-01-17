/*
 Navicat Premium Dump SQL

 Source Server         : localhost
 Source Server Type    : MySQL
 Source Server Version : 80403 (8.4.3)
 Source Host           : localhost:3306
 Source Schema         : ac

 Target Server Type    : MySQL
 Target Server Version : 80403 (8.4.3)
 File Encoding         : 65001

 Date: 16/01/2025 14:26:01
*/

SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

-- ----------------------------
-- Table structure for casbin_rule
-- ----------------------------
DROP TABLE IF EXISTS `casbin_rule`;
CREATE TABLE `casbin_rule` (
  `id` int NOT NULL AUTO_INCREMENT COMMENT 'id',
  `ptype` varchar(100) DEFAULT NULL,
  `v0` varchar(100) DEFAULT NULL,
  `v1` varchar(100) DEFAULT NULL,
  `v2` varchar(100) DEFAULT NULL,
  `v3` varchar(100) DEFAULT NULL,
  `v4` varchar(100) DEFAULT NULL,
  `v5` varchar(100) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_ptype_v0_v1` (`ptype`,`v0`,`v1`),
  KEY `idx_ptype` (`ptype`),
  KEY `idx_v0` (`v0`),
  KEY `idx_v1` (`v1`),
  KEY `idx_casbin_rule` (`ptype`,`v0`,`v1`,`v2`,`v3`,`v4`,`v5`) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- ----------------------------
-- Table structure for casbin_rule_deleted
-- ----------------------------
DROP TABLE IF EXISTS `casbin_rule_deleted`;
CREATE TABLE `casbin_rule_deleted` (
  `id` int NOT NULL AUTO_INCREMENT COMMENT 'id',
  `log_id` int NOT NULL DEFAULT '0' COMMENT 'casbin_rule_log ID',
  `ptype` varchar(255) NOT NULL DEFAULT '' COMMENT 'ptype',
  `v0` varchar(255) NOT NULL DEFAULT '' COMMENT 'v0',
  `v1` varchar(255) NOT NULL DEFAULT '' COMMENT 'v1',
  `v2` varchar(255) NOT NULL DEFAULT '' COMMENT 'v2',
  `v3` varchar(255) NOT NULL DEFAULT '' COMMENT 'v3',
  `v4` varchar(255) NOT NULL DEFAULT '' COMMENT 'v4',
  `v5` varchar(255) NOT NULL DEFAULT '' COMMENT 'v5',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT 'created_at',
  PRIMARY KEY (`id`),
  KEY `idx_ptype` (`ptype`),
  KEY `idx_v0` (`v0`),
  KEY `idx_v1` (`v1`),
  KEY `idx_log_id` (`log_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- ----------------------------
-- Table structure for casbin_rule_log
-- ----------------------------
DROP TABLE IF EXISTS `casbin_rule_log`;
CREATE TABLE `casbin_rule_log` (
  `id` int NOT NULL AUTO_INCREMENT COMMENT 'id',
  `operate` enum('add','delete','set') NOT NULL DEFAULT 'add' COMMENT 'operate',
  `content` text NOT NULL COMMENT 'content',
  `modified_by` varchar(50) NOT NULL DEFAULT '' COMMENT 'modified_by',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT 'created_at',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- ----------------------------
-- Table structure for resource
-- ----------------------------
DROP TABLE IF EXISTS `resource`;
CREATE TABLE `resource` (
  `id` int NOT NULL AUTO_INCREMENT COMMENT 'id',
  `system_code` varchar(50) NOT NULL DEFAULT '' COMMENT 'system_code',
  `name` varchar(50) NOT NULL DEFAULT '' COMMENT 'name',
  `code` varchar(50) NOT NULL DEFAULT '' COMMENT 'code',
  `parent_code` varchar(50) NOT NULL DEFAULT '' COMMENT 'parent_code',
  `description` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT '' COMMENT 'description',
  `modified_by` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT '' COMMENT 'modified_by',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT 'created_at',
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT 'updated_at',
  `deleted_at` datetime DEFAULT NULL COMMENT 'deleted_at',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_system_code_code` (`system_code`,`code`),
  KEY `idx_system_code_name` (`system_code`,`name`),
  KEY `idx_parent_code` (`parent_code`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- ----------------------------
-- Table structure for subject
-- ----------------------------
DROP TABLE IF EXISTS `subject`;
CREATE TABLE `subject` (
  `id` int NOT NULL AUTO_INCREMENT COMMENT 'id',
  `system_code` varchar(50) NOT NULL DEFAULT '' COMMENT 'system_code',
  `type` enum('user','role') CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT 'user' COMMENT 'type',
  `name` varchar(50) NOT NULL DEFAULT '' COMMENT 'name',
  `code` varchar(50) NOT NULL DEFAULT '' COMMENT 'code',
  `description` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT '' COMMENT 'description',
  `modified_by` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT '' COMMENT 'modified_by',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT 'created_at',
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT 'updated_at',
  `deleted_at` datetime DEFAULT NULL COMMENT 'deleted_at',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_system_code_code` (`system_code`,`code`),
  KEY `idx_system_code_name_type` (`system_code`,`name`,`type`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- ----------------------------
-- Table structure for system
-- ----------------------------
DROP TABLE IF EXISTS `system`;
CREATE TABLE `system` (
  `id` int NOT NULL AUTO_INCREMENT COMMENT 'id',
  `name` varchar(50) NOT NULL DEFAULT '' COMMENT 'name',
  `code` varchar(50) NOT NULL DEFAULT '' COMMENT 'code',
  `description` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT '' COMMENT 'description',
  `modified_by` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT '' COMMENT 'modified_by',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT 'created_at',
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT 'updated_at',
  `deleted_at` datetime DEFAULT NULL COMMENT 'deleted_at',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_code` (`code`),
  KEY `idx_name` (`name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

SET FOREIGN_KEY_CHECKS = 1;
