package script

import (
	"context"
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/jmoiron/sqlx"
	gocache "github.com/patrickmn/go-cache"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/suite"
	"pg-start-trainee-2024/domain/entity"
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
	dbConnectionStr string
	jwtSecret       string
)

func init() {
	dbConnectionStr = "postgresql://postgres:postgres@localhost:5433/test?sslmode=disable" // todo: in config?
}

type Suite struct {
	suite.Suite

	db    *sqlx.DB
	cache Cache

	repository Repo
	service    Service
	handler    Handler

	defaultOffset int
	defaultLimit  int
}

func TestSuite(t *testing.T) {
	suite.Run(t, new(Suite))
}

func (s *Suite) setupDB() {
	db, err := dbutils.TryToConnectToDB(dbConnectionStr, "postgres", 5, 5, logrus.New())
	if err != nil {
		s.FailNowf("cannot open database connection with connection string: %v, err: %v", dbConnectionStr, err)
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
	s.service = scriptservice.New(s.repository, s.cache, 1)
}

func (s *Suite) setupHandler() {
	logger := logrus.New()
	valid := validator.New(validator.WithRequiredStructEnabled())

	s.handler = scripthandler.New(s.service, logger, valid)
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
	s.setupDB()
	s.setupCache()
	s.setupRepo()
	s.setupService()
	s.setupHandler()
	s.loadFixturesIntoDB()

	s.defaultOffset = scripthandler.DefaultOffset
	s.defaultLimit = scripthandler.DefaultLimit
}

func (s *Suite) TearDownSuite() {
	// delete all data from db
	_, _ = s.db.Exec("DELETE FROM script WHERE true")

	// close db connection
	_ = s.db.Close()
}
