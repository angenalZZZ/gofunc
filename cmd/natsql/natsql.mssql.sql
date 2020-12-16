CREATE TABLE [dbo].[subscribes] (
	[Id] varchar(36) NOT NULL PRIMARY KEY CLUSTERED, 
	[Name] varchar(100) NOT NULL, 
	[Spec] varchar(2) NOT NULL, 
	[Func] varchar(100) NOT NULL, 
	[Content] text, 
	[CacheDir] varchar(100) NOT NULL, 
	[MsgLimit] int NOT NULL, 
	[BytesLimit] int NOT NULL, 
	[Amount] int NOT NULL, 
	[Bulk] int NOT NULL, 
	[Interval] int NOT NULL, 
	[Version] int NOT NULL DEFAULT ((1)), 
	[Deleted] bit NOT NULL DEFAULT ((0))
) ON [PRIMARY];

CREATE TRIGGER [dbo].[trigger_update_subscribes] ON [subscribes] AFTER UPDATE AS BEGIN
	SET NOCOUNT ON;
	UPDATE [subscribes] SET [Version] = [Version] + 1 WHERE [Id] IN (SELECT [Id] FROM inserted);
END;
