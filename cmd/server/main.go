package main

import (
	srvr "VK/internal/gateways/grpc"
	"VK/internal/repository/inmemory"
	"context"
	"errors"
	"golang.org/x/sync/errgroup"
	"gopkg.in/yaml.v3"
	"log"
	"os"
	"os/signal"
	"syscall"
)

type Config struct {
	Server struct {
		Host string `yaml:"host"`
		Port uint16 `yaml:"port"`
	}
}

func readConfigYaml(cfg *Config) error {
	errCh := make(chan error)

	file, err := os.Open("cmd/server/config.yml")
	if err != nil {
		return err
	}

	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			errCh <- err
		}
	}(file)

	decoder := yaml.NewDecoder(file)
	err = decoder.Decode(cfg)
	if err != nil {
		return err
	}

	close(errCh)

	for err := range errCh {
		if err != nil {
			return err
		}
	}

	return nil
}

func main() {
	var cfg Config
	if err := readConfigYaml(&cfg); err != nil {
		log.Fatal(err)
		return
	}

	host := cfg.Server.Host
	port := cfg.Server.Port

	if host == "" {
		host = "localhost"
	}
	if port == 0 {
		port = 8080
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	eg, ctx := errgroup.WithContext(ctx)
	sigQuit := make(chan os.Signal, 1)
	signal.Notify(sigQuit, syscall.SIGINT, syscall.SIGTERM)

	eg.Go(func() error {
		select {
		case <-sigQuit:
			cancel()
		case <-ctx.Done():
		}
		return nil
	})

	spServer := srvr.NewSubPubServer(inmemory.NewSubPub())
	s := srvr.NewServer(spServer, srvr.WithHost(host), srvr.WithPort(port))

	eg.Go(func() error {
		return s.Run(ctx)
	})

	if err := eg.Wait(); err != nil && !errors.Is(err, context.Canceled) {
		log.Printf("shutdown with error: %v", err)
		return
	}
}
