package main

import (
	"avito-test-assigment/banner_app/internal/config"
	"avito-test-assigment/banner_app/internal/repository"
	"avito-test-assigment/banner_app/internal/repository/transaction_manager"
	"avito-test-assigment/banner_app/internal/server"
	service2 "avito-test-assigment/banner_app/internal/service"
	"context"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v4/pgxpool"
	"log"
	"net/http"
)

func generateDsn(db config.DB) string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", db.Host, db.Port, db.User, db.Password, db.DbName)
}

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal("ошибка загрузки конфига: ", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pool, err := pgxpool.Connect(ctx, generateDsn(cfg.DB))
	if err != nil {
		log.Fatal("ошибка подключения к базе данных: ", err)
	}
	tm := transaction_manager.NewTransactionManager(pool)
	repo := repository.NewBannerRepository(tm)
	service := service2.NewBannerService(repo, tm)
	serv := server.NewBannerServer(service)

	r := chi.NewRouter()
	r.Get("/user_banner", server.Handle(serv.GetByTagAndFeature))
	r.Delete("/banner/{id}", server.Handle(serv.DeleteByID))
	r.Patch("/banner/{id}", server.Handle(serv.Update))
	r.Post("/banner", server.Handle(serv.Create))
	r.Get("/banner", server.Handle(serv.ListBannersByTagOrFeature))

	httpServer := http.Server{Addr: ":8080", Handler: r}
	httpServer.ListenAndServe()
}
