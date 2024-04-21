package pkg

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"fmt"
	"io"
	"log"
	"math/big"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

type FileServer struct {
	Server      *http.Server
	Port        int
	RootPath    string
	StoragePath string
	AuthToken   string
	CertFile    string
	KeyFile     string
	SelfSigned  bool
}

type FileServerOptions struct {
	Port        int
	RootPath    string
	StoragePath string
	AuthToken   string
	CertFile    string
	KeyFile     string
	SelfSigned  bool
}

/*
# http
curl -X POST \
  http://localhost:8080/upload \
  -H 'Authorization: Bearer flargy' \
  -F "file=@WillCodeForFood.png"

# https with valid cert
curl -X POST \
  https://localhost:8080/upload \
  -H 'Authorization: Bearer flargy' \
  -F "file=@WillCodeForFood.png"

# https with self-signed cert
curl -k -X POST \
  https://localhost:8080/upload \
  -H 'Authorization: Bearer flargy' \
  -F "file=@WillCodeForFood.png"
*/

func StartFileServer(options FileServerOptions) (*FileServer, error) {
	// ensure root path exists
	if _, err := os.Stat(options.RootPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("root path does not exist: %s", options.RootPath)
	}

	// if storage path is set, ensure it exists
	if options.StoragePath != "" {
		if _, err := os.Stat(options.StoragePath); os.IsNotExist(err) {
			return nil, fmt.Errorf("storage path does not exist: %s", options.StoragePath)
		}
	}

	fs := &FileServer{
		Port:        options.Port,
		RootPath:    options.RootPath,
		StoragePath: options.StoragePath,
		AuthToken:   options.AuthToken,
		CertFile:    options.CertFile,
		KeyFile:     options.KeyFile,
		SelfSigned:  options.SelfSigned,
	}

	addr := fmt.Sprintf(":%d", fs.Port)
	fs.Server = &http.Server{Addr: addr}

	http.Handle("/", fs.authMiddleware(http.FileServer(http.Dir(fs.RootPath))))
	if fs.StoragePath != "" {
		http.Handle("/upload", fs.authMiddleware(http.HandlerFunc(fs.uploadHandler)))
	}

	go func() {
		log.Printf("Starting file server on port %d...", fs.Port)
		var err error
		if fs.SelfSigned {
			cert, key, err := generateSelfSignedCert()
			if err != nil {
				log.Printf("Failed to generate self-signed certificate: %s\n", err)
				return
			}
			fs.Server.TLSConfig = &tls.Config{
				Certificates: []tls.Certificate{
					{
						Certificate: [][]byte{cert},
						PrivateKey:  key,
					},
				},
			}
			err = fs.Server.ListenAndServeTLS("", "")
		} else if fs.CertFile != "" && fs.KeyFile != "" {
			err = fs.Server.ListenAndServeTLS(fs.CertFile, fs.KeyFile)
		} else {
			err = fs.Server.ListenAndServe()
		}
		if err != nil && err != http.ErrServerClosed {
			log.Printf("listen: %s\n", err)
		}
	}()

	return fs, nil
}

func generateSelfSignedCert() ([]byte, *rsa.PrivateKey, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, err
	}

	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"Self-Signed Certificate"},
		},
		NotBefore: time.Now(),
		NotAfter:  time.Now().Add(365 * 24 * time.Hour),

		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		return nil, nil, err
	}

	return derBytes, privateKey, nil
}

func (fs *FileServer) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if fs.AuthToken != "" {
			token := r.Header.Get("Authorization")
			if token != fs.AuthToken {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}

func (fs *FileServer) uploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Invalid request method", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	dst, err := os.Create(filepath.Join(fs.StoragePath, header.Filename))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	_, err = io.Copy(dst, file)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write([]byte("File uploaded successfully"))
}

func StopFileServer(fs *FileServer) error {
	return fs.Server.Shutdown(nil)
}
