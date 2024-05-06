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
	"syscall"
)

const (
	configPath = "./config" // todo: env vars?
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

//func main() {
//	logger := logrus.New()
//	ctx, cancel := context.WithCancel(context.Background())
//
//	conf, err := initConfig()
//	if err != nil {
//		logger.Fatalf("cannot init config: %v", err)
//	}
//
//	db, err := dbutils.TryToConnectToDB(conf.Postgres.ConnectionURL(), "postgres", conf.Postgres.Retries, conf.Postgres.Interval, logger)
//	if err != nil {
//		logger.Fatalf("cannot connect to db: %v", err)
//	}
//
//	scriptRepo := scriprepo.New(db)
//	scriptService := scriptservice.New(scriptRepo)
//
//	script := entity.Script{
//		Command: "ping google.com",
//	}
//
//	created, err := scriptService.CreateScript(ctx, script)
//	if err != nil {
//		logger.Fatalf("cannot create script: %v", err)
//	}
//
//	logger.Infof("%+v", created)
//
//	// graceful shutdown
//	interrupt := make(chan os.Signal, 1)
//
//	signal.Ignore(syscall.SIGHUP, syscall.SIGPIPE)
//	signal.Notify(interrupt, syscall.SIGINT)
//
//	go func() {
//		<-interrupt
//
//		logger.Info("interrupt signal caught: shutting server down")
//
//		cancel()
//
//		time.Sleep(2 * time.Second) // todo: need some time for gracefully shutdown, TODO: remove sleep!
//	}()
//
//	<-ctx.Done()
//}

func main() {
	logger := logrus.New()
	valid := validator.New(validator.WithRequiredStructEnabled())

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
	scriptService := scriptservice.New(scriptRepo)
	scriptHandler := scripthandler.New(scriptService, logger, valid)

	routers := make(map[string]chi.Router)

	routers["/script"] = scriptHandler.Routes()

	middlewares := []router.Middleware{
		chimiddlewares.Recoverer,
		chimiddlewares.Logger,
	}

	r := router.MakeRoutes("/pg-start-trainee/api/v1", routers, middlewares...) // todo: to conf or const

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

		if shutdownErr := server.Shutdown(ctx); err != nil {
			logger.WithError(shutdownErr).Fatalf("can't close server listening on '%s'", server.Addr)
		}

		cancel()

		// time.Sleep(2 * time.Second) // todo: need some time for gracefully shutdown, TODO: remove sleep!
	}()

	<-ctx.Done()

	// todo: gracefully shutdown, maybe save created filenames in inmemdb and on shutdown check if all created files are deleted?
}
