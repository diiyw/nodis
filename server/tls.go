package server

import (
	"crypto/tls"
	"errors"
	"net"
)

// ListenAndServeTLS encrypts both commands and synchronization connections.
// Supply ClientAuth and ClientCAs in config to enable mutual TLS.
func (s *Server) ListenAndServeTLS(address string, config *tls.Config) error {
	secure, err := serverTLSConfig(config)
	if err != nil {
		return err
	}
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return err
	}
	return s.Serve(tls.NewListener(listener, secure))
}

// ServeTLS takes ownership of listener, including when configuration is invalid.
func (s *Server) ServeTLS(listener net.Listener, config *tls.Config) error {
	secure, err := serverTLSConfig(config)
	if err != nil {
		_ = listener.Close()
		return err
	}
	return s.Serve(tls.NewListener(listener, secure))
}

func serverTLSConfig(config *tls.Config) (*tls.Config, error) {
	if config == nil {
		return nil, errors.New("nodis server: TLS config is required")
	}
	secure := config.Clone()
	if secure.MinVersion < tls.VersionTLS13 {
		secure.MinVersion = tls.VersionTLS13
	}
	if len(secure.Certificates) == 0 && secure.GetCertificate == nil && secure.GetConfigForClient == nil {
		return nil, errors.New("nodis server: TLS certificate is required")
	}
	return secure, nil
}
