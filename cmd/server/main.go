package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"matriz/internal/config"
	"matriz/internal/httpapi"
	"matriz/internal/messaging"
	"matriz/internal/repository"
)

func main() {
	cfg := config.Load()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(cfg.MongoURI))
	if err != nil {
		log.Fatal(err)
	}

	repo, err := repository.NewMongoEmpresaRepo(client, cfg.MongoDB, cfg.MongoCollection)
	if err != nil {
		log.Fatal(err)
	}

	pub, err := messaging.NewPublisher(cfg.RabbitURL, "logs.empresas")
	if err != nil {
		log.Printf("RabbitMQ indisponível: %v", err)
	}
	defer func() {
		if pub != nil {
			pub.Close()
		}
	}()

	api := httpapi.NewServer(repo, pub)
	r := chi.NewRouter()
	r.Mount("/api", api.Routes())

	srv := &http.Server{Addr: ":" + cfg.Port, Handler: r}
	go func() {
		log.Printf("HTTP server on :%s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	ctxShut, cancelShut := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelShut()
	_ = srv.Shutdown(ctxShut)
}
