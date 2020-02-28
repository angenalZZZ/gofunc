package f

// OpensslWebServerCert generate webserver certificates against a private certificate authority,
// input: companyName of Certificate Authority Name, hostname of TLS server to install the private cert/key,
// output: ca.cer ca.key sv.cer sv.key files.
func OpensslWebServerCert(companyName string, hostname string) error {
	// Create Private CA
	_, err := ExecCommandOutput("openssl", "genrsa", "-out", "ca.key", "2048")
	if err != nil {
		return err // log.Fatal("Could not create private Certificate Authority key")
	}
	_, err = ExecCommandOutput("openssl", "req", "-x509", "-new", "-key", "ca.key", "-out", "ca.cer", "-days", "3650", "-subj", "/CN=\""+companyName+"\"")
	if err != nil {
		return err // log.Fatal("Could not create private Certificate Authority certificate")
	}
	// Create Server Cert Key
	_, err = ExecCommandOutput("openssl", "genrsa", "-out", "sv.key", "2048")
	if err != nil {
		return err // log.Fatal("Could not create private server key")
	}
	_, err = ExecCommandOutput("openssl", "req", "-new", "-out", "sv.req", "-key", "sv.key", "-subj", "/CN="+hostname)
	if err != nil {
		return err // log.Fatal("Could not create private server certificate signing request")
	}
	_, err = ExecCommandOutput("openssl", "x509", "-req", "-in", "sv.req", "-out", "sv.cer", "-CAkey", "ca.key", "-CA", "ca.cer", "-days", "3650", "-CAcreateserial", "-CAserial", "serial")
	if err != nil {
		return err // log.Fatal("Could not create private server certificate")
	}
	return nil
}
