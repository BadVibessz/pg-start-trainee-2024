package script

import (
	"github.com/go-chi/chi/v5"
	"net/http"
	"net/http/httptest"
	"pg-start-trainee-2024/pkg/router"
	"strconv"
	"time"
)

func (s *Suite) TestStopNotExistingScript() {
	id := -1

	req, err := http.NewRequest("PATCH", "/test/api/script", nil)
	s.NoError(err)

	req.Header.Set("Content-type", "application/json")
	req.Header.Set("id", strconv.Itoa(id))

	routers := make(map[string]chi.Router)

	routers["/script"] = s.handler.Routes()

	r := router.MakeRoutes("/test/api", routers)

	recorder := httptest.NewRecorder()
	r.ServeHTTP(recorder, req)

	s.Equal(http.StatusBadRequest, recorder.Result().StatusCode)
}

func (s *Suite) TestStopExistingNotRunningScript() {
	// create script for testing
	created := s.createScript("ls -la /")

	// wait some time for script to exit
	time.Sleep(1 * time.Second)

	// test
	id := created.ID

	req, err := http.NewRequest("PATCH", "/test/api/script", nil)
	s.NoError(err)

	req.Header.Set("Content-type", "application/json")
	req.Header.Set("id", strconv.Itoa(id))

	routers := make(map[string]chi.Router)

	routers["/script"] = s.handler.Routes()

	r := router.MakeRoutes("/test/api", routers)

	recorder := httptest.NewRecorder()
	r.ServeHTTP(recorder, req)

	s.Equal(http.StatusBadRequest, recorder.Result().StatusCode)
}

func (s *Suite) TestStopExistingRunningScript() {
	// create script for testing
	created := s.createScript("ping google.com")

	// test
	id := created.ID

	req, err := http.NewRequest("PATCH", "/test/api/script", nil)
	s.NoError(err)

	req.Header.Set("Content-type", "application/json")
	req.Header.Set("id", strconv.Itoa(id))

	routers := make(map[string]chi.Router)

	routers["/script"] = s.handler.Routes()

	r := router.MakeRoutes("/test/api", routers)

	recorder := httptest.NewRecorder()
	r.ServeHTTP(recorder, req)

	s.Equal(http.StatusOK, recorder.Result().StatusCode)

	// check that process stopped
	s.runCheckPidExistsScript(created.PID, "")
}
