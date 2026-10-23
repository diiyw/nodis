package main

import (
	"crypto/tls"
	"crypto/x509"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/diiyw/nodis"
	"github.com/diiyw/nodis/server"
)

func main() {
	address := flag.String("listen", "127.0.0.1:6380", "TCP listen address")
	dir := flag.String("data", "./data", "database directory")
	adminID := flag.String("admin", "admin", "bootstrap administrator ID")
	durable := flag.Bool("durable", true, "fsync every committed transaction")
	checkpoint := flag.Duration("checkpoint-interval", 5*time.Minute, "checkpoint interval; 0 disables automatic compaction")
	certPath := flag.String("tls-cert", "", "server TLS certificate PEM")
	keyPath := flag.String("tls-key", "", "server TLS private key PEM")
	clientCA := flag.String("tls-client-ca", "", "require client certificates signed by this PEM CA")
	flag.Parse()
	token := os.Getenv("NODIS_TOKEN")
	if token == "" {
		log.Fatal("NODIS_TOKEN must be set")
	}
	var secure *tls.Config
	if *certPath != "" || *keyPath != "" || *clientCA != "" {
		certificate, err := tls.LoadX509KeyPair(*certPath, *keyPath)
		if err != nil {
			log.Fatal(err)
		}
		secure = &tls.Config{MinVersion: tls.VersionTLS13, Certificates: []tls.Certificate{certificate}}
		if *clientCA != "" {
			pem, err := os.ReadFile(*clientCA)
			if err != nil {
				log.Fatal(err)
			}
			roots := x509.NewCertPool()
			if !roots.AppendCertsFromPEM(pem) {
				log.Fatal("tls-client-ca contains no certificates")
			}
			secure.ClientAuth, secure.ClientCAs = tls.RequireAndVerifyClientCert, roots
		}
	}
	db, err := nodis.Open(nodis.Options{Dir: *dir, Admin: *adminID, Durable: *durable, CheckpointInterval: *checkpoint})
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	principal := nodis.Principal{ID: *adminID, Roles: []string{"admin"}}
	s := server.New(db, server.StaticTokens{token: principal})
	stopped := make(chan os.Signal, 1)
	signal.Notify(stopped, os.Interrupt, syscall.SIGTERM)
	go func() { <-stopped; _ = s.Close() }()
	if secure != nil {
		err = s.ListenAndServeTLS(*address, secure)
	} else {
		err = s.ListenAndServe(*address)
	}
	if err != nil {
		log.Fatal(err)
	}
}
