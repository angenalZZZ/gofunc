package main

func Prod() {
	defaultService = &cacheServiceImpl{}

	switch *flagSvc {
	case flagSvcGRPC:
		ProdGRPC() // gRPC server
	case flagSvcTCP:
		ProdTCP() // TCP server
	case flagSvcWS:
		ProdWS() // WS server
	}
}
