package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-chi/chi/v5"
	chimiddlewares "github.com/go-chi/chi/v5/middleware"
	"github.com/go-playground/validator/v10"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
	gocache "github.com/patrickmn/go-cache"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	httpswagger "github.com/swaggo/http-swagger"
	"net/http"
	"os"
	"os/signal"
	_ "pg-start-trainee-2024/docs"
	"pg-start-trainee-2024/internal/config"
	scripthandler "pg-start-trainee-2024/internal/handler/script"
	scriprepo "pg-start-trainee-2024/internal/repository/postgres/script"
	scriptservice "pg-start-trainee-2024/internal/service/script"
	"pg-start-trainee-2024/pkg/router"
	dbutils "pg-start-trainee-2024/pkg/utils/db"
	"strconv"
	"syscall"
)

const (
	configPath = "./config"
	baseUri    = "/pg-start-trainee/api/"
	apiVersion = "v1"
)

func initConfig() (*config.Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")

	viper.AddConfigPath(configPath)

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	var conf config.Config
	if err := viper.Unmarshal(&conf); err != nil {
		return nil, err
	}

	// env variables
	if err := godotenv.Load(configPath + "/.env"); err != nil {
		return nil, err
	}

	viper.SetEnvPrefix("pg_start_trainee")
	viper.AutomaticEnv()

	return &conf, nil
}

func shutdownScripts(ctx context.Context, scriptService *scriptservice.Service, cache *gocache.Cache, logger *logrus.Logger) {
	for k := range cache.Items() {
		if id, err := strconv.Atoi(k); err == nil {
			err = scriptService.StopScript(ctx, id)
			if err != nil {
				logger.Errorf("error occurred stopping script: %v", err)
			}
		}
	}
}

func main() {
	logger := logrus.New()
	valid := validator.New(validator.WithRequiredStructEnabled())
	cache := gocache.New(gocache.NoExpiration, 0)

	ctx, cancel := context.WithCancel(context.Background())

	conf, err := initConfig()
	if err != nil {
		logger.Fatalf("cannot init config: %v", err)
	}

	db, err := dbutils.TryToConnectToDB(conf.Postgres.ConnectionURL(), "postgres", conf.Postgres.Retries, conf.Postgres.Interval, logger)
	if err != nil {
		logger.Fatalf("cannot connect to db: %v", err)
	}

	scriptRepo := scriprepo.New(db)
	scriptService := scriptservice.New(scriptRepo, cache, conf.Service.OutputBufferLength)
	scriptHandler := scripthandler.New(scriptService, logger, valid, conf.Handler.DefaultOffset, conf.Handler.DefaultLimit)

	routers := make(map[string]chi.Router)

	routers["/script"] = scriptHandler.Routes()

	middlewares := []router.Middleware{
		chimiddlewares.Recoverer,
		chimiddlewares.Logger,
	}

	r := router.MakeRoutes(baseUri+apiVersion, routers, middlewares...)

	server := http.Server{
		Addr:    fmt.Sprintf(":%v", conf.Server.Port),
		Handler: r,
	}

	// add swagger middleware
	r.Get("/swagger/*", httpswagger.Handler(
		httpswagger.URL(fmt.Sprintf("http://localhost:%v/swagger/doc.json", conf.Server.Port)),
	))

	logger.Infof("server started at port %v", server.Addr)

	go func() {
		if listenErr := server.ListenAndServe(); listenErr != nil && !errors.Is(listenErr, http.ErrServerClosed) {
			logger.WithError(listenErr).Fatalf("server can't listen requests")
		}
	}()

	logger.Infof("documentation available on: http://localhost:%v/swagger/index.html", conf.Server.Port)

	// graceful shutdown
	interrupt := make(chan os.Signal, 1)

	signal.Ignore(syscall.SIGHUP, syscall.SIGPIPE)
	signal.Notify(interrupt, syscall.SIGINT)

	go func() {
		<-interrupt

		logger.Info("interrupt signal caught: shutting server down")

		// stop all running scripts
		shutdownScripts(ctx, scriptService, cache, logger)

		// shutdown http server
		if shutdownErr := server.Shutdown(ctx); err != nil {
			logger.WithError(shutdownErr).Fatalf("can't close server listening on '%s'", server.Addr)
		}

		cancel()
	}()

	<-ctx.Done()
}
