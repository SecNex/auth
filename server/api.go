package server

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/secnex/auth/database"
	"github.com/secnex/auth/middlewares/auth"
	"github.com/secnex/auth/middlewares/logger"
	"golang.org/x/crypto/acme/autocert"
)

type ApiServer struct {
	Host      *string
	Port      *int
	Databases map[string]*database.Database
	Router    *mux.Router
}

func NewApiServer(host *string, port *int, dbHost string, dbPort int, dbUser string, dbPassword string, dbName string) *ApiServer {
	databases := make(map[string]*database.Database)
	databases["auth"] = database.NewConnection(dbHost, dbPort, dbUser, dbPassword, dbName)
	return &ApiServer{
		Host:      host,
		Port:      port,
		Databases: databases,
		Router:    mux.NewRouter(),
	}
}

func (s *ApiServer) Healthz(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func (s *ApiServer) Test(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("This is a test!"))
}

func (s *ApiServer) setupRoutes() {
	// Public routes (no authentication required)
	s.Router.HandleFunc("/healthz", s.Healthz).Methods(http.MethodGet)

	// Protected routes (authentication required)
	protected := s.Router.PathPrefix("/").Subrouter()
	protected.Use(auth.AuthorizationMiddleware)
	protected.HandleFunc("/test", s.Test).Methods(http.MethodGet)
}

func (s *ApiServer) Start() {
	https := true

	domain := os.Getenv("DOMAIN")
	if domain == "" {
		log.Println("DOMAIN is not set, using HTTP only!")
		https = false
	}

	addr := ":443"
	if !https {
		addr = ":80"
	}

	if s.Host != nil && s.Port != nil {
		addr = fmt.Sprintf("%s:%d", *s.Host, *s.Port)
	}

	log.Printf("Starting api server on %s...", addr)

	// Setup routes with selective authentication
	s.setupRoutes()

	// Apply logging middleware to all routes
	handler := logger.LogHTTPRequest(s.Router)

	if https {
		s.StartHTTPS(handler, addr, domain)
	} else {
		s.StartHTTP(handler)
	}
}

func (s *ApiServer) StartHTTP(handler http.Handler) {
	server := &http.Server{
		Addr:    ":80",
		Handler: handler,
	}

	log.Fatal(server.ListenAndServe())
}

func (s *ApiServer) StartHTTPS(handler http.Handler, addr string, domain string) {
	m := &autocert.Manager{
		Prompt:     autocert.AcceptTOS,
		Cache:      autocert.DirCache("certs"),
		HostPolicy: autocert.HostWhitelist(domain),
	}

	server := &http.Server{
		Addr: addr,
		TLSConfig: &tls.Config{
			GetCertificate: m.GetCertificate,
		},
		Handler: handler,
	}

	go func() {
		http.ListenAndServe(":80", m.HTTPHandler(nil))
	}()

	log.Fatal(server.ListenAndServeTLS("", ""))
}
