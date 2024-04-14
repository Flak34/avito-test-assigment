package main

import (
	"avito-test-assigment/internal/config"
	"avito-test-assigment/internal/middleware"
	"avito-test-assigment/internal/repository"
	"avito-test-assigment/internal/repository/transaction_manager"
	"avito-test-assigment/internal/server"
	service2 "avito-test-assigment/internal/service"
	"context"
	"errors"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v4/pgxpool"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
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
	bannerRepository := repository.NewBannerRepository(tm)
	tagRepository := repository.NewTagRepository(tm)
	versionRepository := repository.NewVersionRepository(tm)

	service := service2.NewBannerService(bannerRepository, tm, tagRepository, versionRepository)
	bannerServer := server.NewBannerServer(service)

	authServer := server.NewAuthServer(cfg.Authentication)

	r := chi.NewRouter()
	secret := cfg.Authentication.SecretKey

	r.With(middleware.TokenAuth(secret, false)).
		Get("/user_banner", server.Handle(bannerServer.GetByTagAndFeature))

	r.With(middleware.TokenAuth(secret, true)).
		Delete("/banner/{id}", server.Handle(bannerServer.DeleteByID))

	r.With(middleware.TokenAuth(secret, true)).
		Patch("/banner/{id}", server.Handle(bannerServer.Update))

	r.With(middleware.TokenAuth(secret, true)).
		Post("/banner", server.Handle(bannerServer.Create))

	r.With(middleware.TokenAuth(secret, true)).
		Get("/banner", server.Handle(bannerServer.ListBannersByTagOrFeature))

	r.Post("/login", authServer.Login)

	r.With(middleware.TokenAuth(secret, true)).
		Delete("/banner", server.Handle(bannerServer.DeleteByTagOrFeature))

	r.With(middleware.TokenAuth(secret, true)).
		Get("/banner/{id}/version", server.Handle(bannerServer.GetBannerVersions))

	//реализация graceful shutdown для сервера
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		httpsServer := http.Server{Addr: cfg.HTTPServer.Port, Handler: r}
		idleConnsClosed := make(chan struct{})

		//ожидание сигнала завершения
		go func() {
			sigChan := make(chan os.Signal, 1)
			signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)
			<-sigChan
			if err = httpsServer.Shutdown(context.Background()); err != nil {
				log.Printf("HTTP pickupPointServer Shutdown: %v", err)
			}
			idleConnsClosed <- struct{}{}
		}()

		log.Print("Сервер http запущен и слушает подключения на порту " + cfg.HTTPServer.Port[1:])
		if err = httpsServer.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			log.Fatal(err)
		}

		<-idleConnsClosed
	}()

	wg.Wait()
	fmt.Println("Сервис завершил свою работу")
}
