
CREATE TABLE [logtest] (
	[Id] bigint NOT NULL IDENTITY(1,1) PRIMARY KEY CLUSTERED,
	[Code] varchar(50),
	[Type] int,
	[Message] varchar(4000) NOT NULL,
	[Account] varchar(36),
	[CreateTime] datetime NOT NULL,
	[CreateUser] varchar(36)
) ON [PRIMARY];
EXEC sp_addextendedproperty @name=N'MS_Description',@value=N'自增ID',@level0type=N'Schema',@level0name=N'dbo',@level1type=N'Table',@level1name=N'logtest',@level2type=N'Column',@level2name=N'Id';
EXEC sp_addextendedproperty @name=N'MS_Description',@value=N'测试记录',@level0type=N'Schema',@level0name=N'dbo',@level1type=N'Table',@level1name=N'logtest';
