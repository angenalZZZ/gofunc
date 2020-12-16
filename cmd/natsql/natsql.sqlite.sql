CREATE TABLE IF NOT EXISTS subscribes (
	Name varchar(100) NOT NULL PRIMARY KEY, 
	Spec varchar(2) NOT NULL, 
	Func varchar(100) NOT NULL, 
	Content text(2097152), 
	CacheDir varchar(100) NOT NULL, 
	MsgLimit int NOT NULL, 
	BytesLimit int NOT NULL, 
	Amount int NOT NULL, 
	Bulk int NOT NULL, 
	Interval int NOT NULL, 
	Version int NOT NULL DEFAULT 1, 
	Deleted bit NOT NULL DEFAULT 0
);

CREATE TRIGGER IF NOT EXISTS trigger_update_subscribes UPDATE ON subscribes BEGIN
	UPDATE subscribes SET Version = OLD.Version + 1 WHERE Name = OLD.Name;
END;
