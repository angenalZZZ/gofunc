package main

func Prod() {
	switch *flagSvc {
	case flagSvcGRPC:
		ProdGRPC() // gRPC server
	case flagSvcTCP:
		ProdTCP() // TCP server
	}
}
