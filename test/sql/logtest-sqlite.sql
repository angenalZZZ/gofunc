
CREATE TABLE [logtest] (
	[Id] integer NOT NULL PRIMARY KEY AUTOINCREMENT,
	[Code] varchar(50) NOT NULL,
	[Type] int NOT NULL,
	[Message] nvarchar(2000) NOT NULL,
	[Account] varchar(36),
	[CreateTime] datetime NOT NULL,
	[CreateUser] varchar(36)
);
