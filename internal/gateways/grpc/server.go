package grpc

import (
	gen "VK/internal/gateways/generated"
	"context"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"log"
	"net"
	"strconv"
)

type Server struct {
	host   string
	port   uint16
	server *grpc.Server
	subpub *SubPubServer
}

func NewServer(sp *SubPubServer, options ...func(*Server)) *Server {
	s := grpc.NewServer()
	server := &Server{
		host:   "localhost",
		port:   8080,
		server: s,
		subpub: sp,
	}

	for _, option := range options {
		option(server)
	}

	return server
}

func WithHost(host string) func(*Server) {
	return func(server *Server) {
		server.host = host
	}
}

func WithPort(port uint16) func(*Server) {
	return func(server *Server) {
		server.port = port
	}
}

func (s *Server) Run(ctx context.Context) error {
	listener, err := net.Listen("tcp", s.host+":"+strconv.FormatUint(uint64(s.port), 10))
	if err != nil {
		log.Fatalf("Failed to listen on %s:%d: %v", s.host, s.port, err)
		return err
	}

	defer listener.Close()

	gen.RegisterPubSubServer(s.server, s.subpub)
	reflection.Register(s.server)

	eg, egCtx := errgroup.WithContext(ctx)
	eg.Go(func() error {
		log.Printf("Listening on %s:%d", s.host, s.port)
		return s.server.Serve(listener)
	})

	go func() {
		<-egCtx.Done()
		log.Printf("Shutting down server on %s:%d", s.host, s.port)

		if err := s.subpub.subPub.Close(context.Background()); err != nil {
			log.Printf("Error closing SubPub: %v", err)
		}
		s.server.Stop()
	}()

	return eg.Wait()
}
