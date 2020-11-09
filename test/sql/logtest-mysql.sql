
CREATE TABLE IF NOT EXISTS `logtest` (
  `Id` bigint(20) NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增ID',
  `Code` varchar(50) DEFAULT NULL,
  `Type` int(11) DEFAULT NULL,
  `Message` varchar(4000) NOT NULL,
  `Account` varchar(36) DEFAULT NULL,
  `CreateTime` datetime NOT NULL,
  `CreateUser` varchar(36) DEFAULT NULL
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='测试记录';
