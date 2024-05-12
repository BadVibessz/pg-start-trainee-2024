package script

import (
	"context"
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	gocache "github.com/patrickmn/go-cache"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/suite"
	"pg-start-trainee-2024/domain/entity"
	"pg-start-trainee-2024/internal/config"
	scripthandler "pg-start-trainee-2024/internal/handler/script"
	scriptservice "pg-start-trainee-2024/internal/service/script"
	dbutils "pg-start-trainee-2024/pkg/utils/db"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	scriptrepo "pg-start-trainee-2024/internal/repository/postgres/script"
)

type Repo interface {
	CreateScript(ctx context.Context, script entity.Script) (*entity.Script, error)
	UpdateScriptOutput(ctx context.Context, id int, output string) (*entity.Script, error)
	DeleteScript(ctx context.Context, id int) (*entity.Script, error)
	UpdateScriptPIDAndRunningState(ctx context.Context, id, pid int, isRunning bool) (*entity.Script, error)
	UpdateScriptRunningState(ctx context.Context, id int, isRunning bool) (*entity.Script, error)
	GetScript(ctx context.Context, id int) (*entity.Script, error)
	GetAllScripts(ctx context.Context, offset, limit int) ([]*entity.Script, error)
}

type Cache interface {
	Set(key string, value any, duration time.Duration)
	Get(key string) (any, bool)
	Delete(key string)
}

type Service interface {
	CreateScript(ctx context.Context, script entity.Script) (*entity.Script, error)
	StopScript(ctx context.Context, id int) error
	GetScript(ctx context.Context, id int) (*entity.Script, error)
	GetAllScripts(ctx context.Context, offset, limit int) ([]*entity.Script, error)
	DeleteScript(ctx context.Context, id int) error
}

type Handler interface {
	Routes() *chi.Mux
}

var (
	configPath = "../../config/"
)

func initConfig() (*config.Config, error) {
	viper.SetConfigName("testing_config")
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

	viper.SetEnvPrefix("pg_start_trainee_test")
	viper.AutomaticEnv()

	return &conf, nil
}

type Suite struct {
	suite.Suite

	config *config.Config

	db    *sqlx.DB
	cache Cache

	repository Repo
	service    Service
	handler    Handler
}

func TestSuite(t *testing.T) {
	suite.Run(t, new(Suite))
}

func (s *Suite) setupConfig() {
	conf, err := initConfig()
	if err != nil {
		s.FailNowf(err.Error(), err.Error())
	}

	s.config = conf
}

func (s *Suite) setupDB() {
	db, err := dbutils.TryToConnectToDB(s.config.Postgres.ConnectionURL(), "postgres", 5, 5, logrus.New())
	if err != nil {
		s.FailNowf("cannot open database connection with connection string: %v, err: %v", s.config.Postgres.ConnectionURL(), err)
	}

	s.db = db
}

func (s *Suite) setupCache() {
	s.cache = gocache.New(gocache.NoExpiration, gocache.NoExpiration)
}

func (s *Suite) setupRepo() {
	s.repository = scriptrepo.New(s.db)
}

func (s *Suite) setupService() {
	s.service = scriptservice.New(s.repository, s.cache, s.config.Service.OutputBufferLength)
}

func (s *Suite) setupHandler() {
	logger := logrus.New()
	valid := validator.New(validator.WithRequiredStructEnabled())

	s.handler = scripthandler.New(s.service, logger, valid, s.config.DefaultOffset, s.config.DefaultLimit)
}

func (s *Suite) loadFixturesIntoDB() {
	for _, script := range scripts {
		_, err := s.repository.CreateScript(context.Background(), script)
		if err != nil {
			s.FailNowf(err.Error(), err.Error())
		}
	}
}

func (s *Suite) SetupSuite() {
	s.setupConfig()
	s.setupDB()
	s.setupCache()
	s.setupRepo()
	s.setupService()
	s.setupHandler()
	s.loadFixturesIntoDB()
}

func (s *Suite) TearDownSuite() {
	// delete all data from db
	_, _ = s.db.Exec("DELETE FROM script WHERE true")

	// close db connection
	_ = s.db.Close()
}
