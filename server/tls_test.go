package server_test

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"math/big"
	"net"
	"testing"
	"time"

	"github.com/diiyw/nodis"
	"github.com/diiyw/nodis/client"
	"github.com/diiyw/nodis/server"
)

type certificateSet struct {
	server tls.Certificate
	client tls.Certificate
	roots  *x509.CertPool
}

func testCertificates(t *testing.T) certificateSet {
	t.Helper()
	caKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	ca := &x509.Certificate{SerialNumber: big.NewInt(1), Subject: pkix.Name{CommonName: "Nodis test CA"}, NotBefore: time.Now().Add(-time.Hour), NotAfter: time.Now().Add(time.Hour), IsCA: true, BasicConstraintsValid: true, KeyUsage: x509.KeyUsageCertSign}
	der, err := x509.CreateCertificate(rand.Reader, ca, ca, &caKey.PublicKey, caKey)
	if err != nil {
		t.Fatal(err)
	}
	parsedCA, err := x509.ParseCertificate(der)
	if err != nil {
		t.Fatal(err)
	}
	roots := x509.NewCertPool()
	roots.AddCert(parsedCA)
	leaf := func(serial int64, usage x509.ExtKeyUsage) tls.Certificate {
		key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		if err != nil {
			t.Fatal(err)
		}
		template := &x509.Certificate{SerialNumber: big.NewInt(serial), Subject: pkix.Name{CommonName: "localhost"}, DNSNames: []string{"localhost"}, IPAddresses: []net.IP{net.ParseIP("127.0.0.1")}, NotBefore: time.Now().Add(-time.Hour), NotAfter: time.Now().Add(time.Hour), KeyUsage: x509.KeyUsageDigitalSignature, ExtKeyUsage: []x509.ExtKeyUsage{usage}}
		der, err := x509.CreateCertificate(rand.Reader, template, parsedCA, &key.PublicKey, caKey)
		if err != nil {
			t.Fatal(err)
		}
		return tls.Certificate{Certificate: [][]byte{der}, PrivateKey: key}
	}
	return certificateSet{server: leaf(2, x509.ExtKeyUsageServerAuth), client: leaf(3, x509.ExtKeyUsageClientAuth), roots: roots}
}

func TestTLSCRUDSyncAndCertificateVerification(t *testing.T) {
	certs := testCertificates(t)
	for _, mutual := range []bool{false, true} {
		t.Run(map[bool]string{false: "serverTLS", true: "mutualTLS"}[mutual], func(t *testing.T) {
			admin := nodis.Principal{ID: "admin"}
			user := nodis.Principal{ID: "reader"}
			db, err := nodis.Open(nodis.Options{})
			if err != nil {
				t.Fatal(err)
			}
			defer db.Close()
			doc, _ := db.CreateDocument(admin, "secure")
			if err := db.Grant(admin, nodis.Rule{SubjectKind: nodis.SubjectUser, Subject: user.ID, Resource: nodis.Resource{Scope: nodis.ScopeDocument, Document: doc.ID()}, Actions: nodis.ActionAll}); err != nil {
				t.Fatal(err)
			}
			listener, err := net.Listen("tcp", "127.0.0.1:0")
			if err != nil {
				t.Fatal(err)
			}
			s := server.New(db, server.StaticTokens{"token": user})
			config := &tls.Config{Certificates: []tls.Certificate{certs.server}}
			if mutual {
				config.ClientAuth, config.ClientCAs = tls.RequireAndVerifyClientCert, certs.roots
			}
			done := make(chan error, 1)
			go func() { done <- s.ServeTLS(listener, config) }()
			defer func() {
				_ = s.Close()
				if err := <-done; err != nil {
					t.Error(err)
				}
			}()
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			clientConfig := &tls.Config{RootCAs: certs.roots, ServerName: "localhost"}
			if mutual {
				clientConfig.Certificates = []tls.Certificate{certs.client}
			}
			remote, err := client.DialTLS(ctx, listener.Addr().String(), []byte("token"), nodis.SyncScope{Documents: []nodis.DocumentID{doc.ID()}}, clientConfig)
			if err != nil {
				t.Fatal(err)
			}
			defer remote.Close()
			id, err := remote.CreateAsync(ctx, doc.ID(), map[string]any{"name": "Ada", "secret": "private"}).Wait(ctx)
			if err != nil {
				t.Fatal(err)
			}
			object, err := remote.Get(ctx, doc.ID(), id)
			if err != nil {
				t.Fatal(err)
			}
			if name, _ := object.String("name"); name != "Ada" {
				t.Fatal("TLS result changed")
			}
			filter, err := nodis.Filter("name", nodis.Equal, "Ada")
			if err != nil {
				t.Fatal(err)
			}
			page, err := remote.FindAsync(ctx, nodis.Query{Document: doc.ID(), Where: []nodis.Condition{filter}, Select: []string{"name"}}).Wait(ctx)
			if err != nil || len(page.Objects) != 1 || page.Objects[0].ID() != id {
				t.Fatal("TLS query failed", err)
			}
			secret, _ := doc.FieldID("secret")
			if err := db.Deny(admin, nodis.Rule{SubjectKind: nodis.SubjectUser, Subject: user.ID, Resource: nodis.Resource{Scope: nodis.ScopeField, Document: doc.ID(), Field: secret}, Actions: nodis.ActionRead | nodis.ActionSync}); err != nil {
				t.Fatal(err)
			}
			deadline := time.Now().Add(3 * time.Second)
			for {
				if cached, ok := remote.Cached(doc.ID(), id); ok {
					if _, hidden := cached.String("secret"); !hidden {
						break
					}
				}
				if time.Now().After(deadline) {
					t.Fatal("TLS resubscription did not apply revoked permissions")
				}
				time.Sleep(time.Millisecond)
			}
			badName := clientConfig.Clone()
			badName.ServerName = "not-localhost"
			if bad, err := client.DialTLS(ctx, listener.Addr().String(), []byte("token"), nodis.SyncScope{}, badName); err == nil {
				bad.Close()
				t.Fatal("accepted wrong server name")
			}
			untrusted := clientConfig.Clone()
			untrusted.RootCAs = x509.NewCertPool()
			if bad, err := client.DialTLS(ctx, listener.Addr().String(), []byte("token"), nodis.SyncScope{}, untrusted); err == nil {
				bad.Close()
				t.Fatal("accepted untrusted CA")
			}
			if mutual {
				missing := clientConfig.Clone()
				missing.Certificates = nil
				if bad, err := client.DialTLS(ctx, listener.Addr().String(), []byte("token"), nodis.SyncScope{}, missing); err == nil {
					bad.Close()
					t.Fatal("accepted absent client certificate")
				}
			}
			legacy := clientConfig.Clone()
			legacy.MinVersion, legacy.MaxVersion = tls.VersionTLS12, tls.VersionTLS12
			legacyConn, err := (&tls.Dialer{Config: legacy}).DialContext(ctx, "tcp", listener.Addr().String())
			if err == nil {
				_ = legacyConn.Close()
				t.Fatal("accepted TLS 1.2")
			}
			if config.MinVersion != 0 || clientConfig.MinVersion != 0 {
				t.Fatal("TLS API modified caller configuration")
			}
		})
	}
}

func TestTLSRejectsConfigurationBeforeListening(t *testing.T) {
	db, _ := nodis.Open(nodis.Options{})
	defer db.Close()
	s := server.New(db, server.StaticTokens{})
	if err := s.ListenAndServeTLS("127.0.0.1:0", nil); err == nil {
		t.Fatal("accepted nil TLS config")
	}
	if err := s.ListenAndServeTLS("127.0.0.1:0", &tls.Config{}); err == nil {
		t.Fatal("accepted missing server certificate")
	}
}
