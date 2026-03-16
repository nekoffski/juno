package rest

// import (
// 	"context"
// 	"net/http"

// 	"github.com/labstack/echo/v4"
// 	"github.com/labstack/echo/v4/middleware"
// 	"go.uber.org/zap"
// )

// const version = "1.0.0"

// var log = zap.L().Named("httpserver")

// type Server struct {
// 	echo *echo.Echo
// 	addr string
// }

// func New(addr string) *Server {
// 	e := echo.New()
// 	e.HideBanner = true
// 	e.Use(middleware.Recover())
// 	e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
// 		LogURI:    true,
// 		LogStatus: true,
// 		LogMethod: true,
// 		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
// 			log.Info("request",
// 				zap.String("method", v.Method),
// 				zap.String("uri", v.URI),
// 				zap.Int("status", v.Status),
// 			)
// 			return nil
// 		},
// 	}))

// 	s := &Server{echo: e, addr: addr}
// 	RegisterHandlers(e, NewStrictHandler(s, nil))
// 	return s
// }

// func (s *Server) Name() string { return "httpserver" }

// func (s *Server) Init(_ context.Context) error { return nil }

// // func (s *Server) Run(ctx context.Context) error {
// // 	go func() {
// // 		<-ctx.Done()
// // 		s.Stop()
// // 	}()
// // 	if err := s.Start(); !errors.Is(err, http.ErrServerClosed) {
// // 		return err
// // 	}
// // 	return nil
// // }

// func (s *Server) GetHealth(_ context.Context, _ GetHealthRequestObject) (GetHealthResponseObject, error) {
// 	return GetHealth200JSONResponse{
// 		Status:  "ok",
// 		Version: version,
// 	}, nil
// }

// // func (s *Server) Start() error {
// // 	log.Info("listening", zap.String("addr", s.addr))
// // 	return s.echo.Start(s.addr)
// // }

// // func (s *Server) Stop() error {
// // 	log.Info("shutting down")
// // 	return s.echo.Close()
// // }

// func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
// 	s.echo.ServeHTTP(w, r)
// }
