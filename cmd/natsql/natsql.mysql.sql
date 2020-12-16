CREATE TABLE `subscribes` (
  `Id` varchar(36) NOT NULL,
  `Name` varchar(100) NOT NULL,
  `Spec` varchar(2) NOT NULL,
  `Func` varchar(100) NOT NULL,
  `Content` mediumtext,
  `CacheDir` varchar(100) NOT NULL,
  `MsgLimit` int(11) NOT NULL,
  `BytesLimit` int(11) NOT NULL,
  `Amount` int(11) NOT NULL,
  `Bulk` int(11) NOT NULL,
  `Interval` int(11) NOT NULL,
  `Version` int(11) NOT NULL DEFAULT '1',
  `Deleted` bit(1) NOT NULL DEFAULT b'0',
  PRIMARY KEY (`Id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TRIGGER `trigger_update_subscribes` AFTER UPDATE ON `subscribes` FOR EACH ROW BEGIN
	UPDATE `subscribes` SET `Version` = OLD.`Version` + 1 WHERE `Id` = OLD.`Id`;
END;
