package script

import (
	"context"
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"net/http"
	"net/http/httptest"
	"pg-start-trainee-2024/domain/entity"
	"pg-start-trainee-2024/internal/handler/response"
	"pg-start-trainee-2024/pkg/router"
	"strconv"
	"time"
	"unicode/utf8"
)

func (s *Suite) createScript(command string) *entity.Script {
	created, err := s.service.CreateScript(context.Background(), entity.Script{Command: command})
	s.NoError(err)

	return created
}

func (s *Suite) scriptAndResponseEqual(script *entity.Script, resp *response.GetScript) {
	s.Equal(script.ID, resp.ID)
	s.Equal(script.Command, resp.Command)
	s.Equal(script.CreatedAt, resp.CreatedAt)
}

func (s *Suite) TestGetNotExistingScript() {
	req, err := http.NewRequest("GET", "/test/api/script", nil)
	s.NoError(err)

	req.Header.Set("Content-type", "application/json")
	req.Header.Set("id", strconv.Itoa(-1))

	routers := make(map[string]chi.Router)

	routers["/script"] = s.handler.Routes()

	r := router.MakeRoutes("/test/api", routers)

	recorder := httptest.NewRecorder()
	r.ServeHTTP(recorder, req)

	s.Equal(http.StatusBadRequest, recorder.Result().StatusCode)
}

func (s *Suite) TestGetShortScript() {
	// create script for testing
	created := s.createScript("ls -la /")

	// wait some time for process to exit
	time.Sleep(1 * time.Second)

	// test
	req, err := http.NewRequest("GET", "/test/api/script", nil)
	s.NoError(err)

	req.Header.Set("Content-type", "application/json")
	req.Header.Set("id", strconv.Itoa(created.ID))

	routers := make(map[string]chi.Router)

	routers["/script"] = s.handler.Routes()

	r := router.MakeRoutes("/test/api", routers)

	recorder := httptest.NewRecorder()
	r.ServeHTTP(recorder, req)

	s.Equal(http.StatusOK, recorder.Result().StatusCode)

	var resp response.GetScript
	s.NoError(json.Unmarshal([]byte(recorder.Body.String()), &resp))

	s.scriptAndResponseEqual(created, &resp)
	s.Equal(false, resp.IsRunning)
}

func (s *Suite) TestGetRunningLongScript() {
	// create script for testing
	created := s.createScript("ping google.com")

	// test
	req, err := http.NewRequest("GET", "/test/api/script", nil)
	s.NoError(err)

	req.Header.Set("Content-type", "application/json")
	req.Header.Set("id", strconv.Itoa(created.ID))

	routers := make(map[string]chi.Router)

	routers["/script"] = s.handler.Routes()

	r := router.MakeRoutes("/test/api", routers)

	recorder := httptest.NewRecorder()
	r.ServeHTTP(recorder, req)

	s.Equal(http.StatusOK, recorder.Result().StatusCode)

	var resp response.GetScript
	s.NoError(json.Unmarshal([]byte(recorder.Body.String()), &resp))

	s.scriptAndResponseEqual(created, &resp)
	s.Equal(true, resp.IsRunning)

	// check that process is running
	s.runCheckPidExistsScript(resp.PID, "running")

	// wait some time and check if output is changed
	time.Sleep(1 * time.Second)

	got, err := getScriptFromDB(s.db, created.ID)
	s.NoError(err)

	s.True(utf8.RuneCountInString(got.Output) > utf8.RuneCountInString(resp.Output))

	// stop created script
	stopScript(s.cache, created.ID)
}

func (s *Suite) TestGetStoppedLongScript() {
	// create script for testing
	created := s.createScript("ping google.com")

	// wait some time and then stop it
	time.Sleep(1 * time.Second)
	s.NoError(s.service.StopScript(context.Background(), created.ID))

	// test
	req, err := http.NewRequest("GET", "/test/api/script", nil)
	s.NoError(err)

	req.Header.Set("Content-type", "application/json")
	req.Header.Set("id", strconv.Itoa(created.ID))

	routers := make(map[string]chi.Router)

	routers["/script"] = s.handler.Routes()

	r := router.MakeRoutes("/test/api", routers)

	recorder := httptest.NewRecorder()
	r.ServeHTTP(recorder, req)

	s.Equal(http.StatusOK, recorder.Result().StatusCode)

	var resp response.GetScript
	s.NoError(json.Unmarshal([]byte(recorder.Body.String()), &resp))

	s.scriptAndResponseEqual(created, &resp)
	s.Equal(false, resp.IsRunning)

	// wait some time and check if output is changed
	time.Sleep(1 * time.Second)

	got, err := getScriptFromDB(s.db, created.ID)
	s.NoError(err)

	s.True(got.Output == resp.Output)
}
