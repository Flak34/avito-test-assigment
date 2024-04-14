//go:build integration

package test

import (
	"avito-test-assigment/internal/config"
	"avito-test-assigment/internal/middleware"
	"avito-test-assigment/internal/repository"
	"avito-test-assigment/internal/repository/transaction_manager"
	"avito-test-assigment/internal/server"
	service2 "avito-test-assigment/internal/service"
	"context"
	"encoding/base64"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt"
	"github.com/jackc/pgx/v4/pgxpool"
	"log"
	"net/http"
	"os"
	"testing"
	"time"
)

var (
	cfg              *config.Config
	ctx              context.Context
	cancel           context.CancelFunc
	bannerRepository *repository.BannerRepository
	pool             *pgxpool.Pool
	token            string
	service          *service2.BannerService
)

func TestMain(m *testing.M) {
	os.Setenv("CONFIG_PATH", "./config/test_config.yaml")
	getwd, _ := os.Getwd()
	fmt.Println(getwd)

	err := setUp()
	if err != nil {
		log.Fatal(err.Error())
	}

	m.Run()

	tearDown()
}

func setUp() error {
	var err error
	cfg, err = config.LoadConfig()

	if err != nil {
		log.Fatal("ошибка загрузки конфига: ", err)
	}

	ctx, cancel = context.WithCancel(context.Background())
	defer cancel()

	pool, err = pgxpool.Connect(ctx, generateDsn(cfg.DB))
	if err != nil {
		log.Fatal("ошибка подключения к базе данных: ", err)
	}

	tm := transaction_manager.NewTransactionManager(pool)
	bannerRepository = repository.NewBannerRepository(tm)
	tagRepository := repository.NewTagRepository(tm)
	versionRepository := repository.NewVersionRepository(tm)

	service = service2.NewBannerService(bannerRepository, tm, tagRepository, versionRepository)
	bannerServer := server.NewBannerServer(service)

	r := chi.NewRouter()
	secret := cfg.Authentication.SecretKey

	r.With(middleware.TokenAuth(secret, false)).
		Get("/user_banner", server.Handle(bannerServer.GetByTagAndFeature))

	secretKey, err := base64.StdEncoding.DecodeString(cfg.Authentication.SecretKey)
	if err != nil {
		log.Fatal("ошибка при декодировании секретного ключа")
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256,
		jwt.MapClaims{
			"role": "user",
			"exp":  time.Now().Add(time.Hour * 24).Unix(),
		})

	token, err = t.SignedString(secretKey)
	if err != nil {
		log.Fatal("ошибка при декодировании создании токена")
	}

	go func() {
		http.ListenAndServe(cfg.HTTPServer.Port, r)
	}()

	return nil
}

func tearDown() {
	pool.Close()
	cancel()
}

func generateDsn(db config.DB) string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", db.Host, db.Port, db.User, db.Password, db.DbName)
}
