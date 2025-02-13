package monitoring

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

type Server struct {
	logger *zap.Logger
	port   string
}

func NewServer(logger *zap.Logger, port string) *Server {
	return &Server{
		logger: logger,
		port:   port,
	}
}

func (s *Server) Start() error {
	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/health", s.healthCheck)

	s.logger.Info("Starting monitoring server", zap.String("port", s.port))
	return http.ListenAndServe(":"+s.port, nil)
}

func (s *Server) healthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}
