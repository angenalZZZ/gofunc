package nsq

var (
	// Default NSQ TCP Address
	NSQdTCPAddr     = "127.0.0.1:4150"
	NSQdTCPAddrList = []string{"127.0.0.2:4150"}

	// Default LOOKUP HTTP Address
	LOOKUPdHTTPAddr     = "127.0.0.1:4161"
	LOOKUPdHTTPAddrList = []string{"127.0.0.1:4161"}

	// Test Topic and Channel
	TestTopic   = "TestTopic"
	TestChannel = "TestChannel"
)
