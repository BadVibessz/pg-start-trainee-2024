package script

import (
	"github.com/go-chi/chi/v5"
	"net/http"
	"net/http/httptest"
	"pg-start-trainee-2024/pkg/router"
	"strconv"
)

func (s *Suite) TestDeleteNotExistingScript() {
	id := -1

	req, err := http.NewRequest("DELETE", "/test/api/script", nil)
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

func (s *Suite) TestDeleteExistingScript() {
	// create script for testing
	created := s.createScript("ls -la /") // no need to stop this process as command short

	// test
	id := created.ID

	req, err := http.NewRequest("DELETE", "/test/api/script", nil)
	s.NoError(err)

	req.Header.Set("Content-type", "application/json")
	req.Header.Set("id", strconv.Itoa(id))

	routers := make(map[string]chi.Router)

	routers["/script"] = s.handler.Routes()

	r := router.MakeRoutes("/test/api", routers)

	recorder := httptest.NewRecorder()
	r.ServeHTTP(recorder, req)

	s.Equal(http.StatusOK, recorder.Result().StatusCode)

	// check that scrip deleted from db
	_, err = getScriptFromDB(s.db, id)
	s.Error(err)
}

func (s *Suite) TestDeleteRunningLongScript() {
	// create script for testing
	created := s.createScript("ping google.com")

	// test
	id := created.ID

	req, err := http.NewRequest("DELETE", "/test/api/script", nil)
	s.NoError(err)

	req.Header.Set("Content-type", "application/json")
	req.Header.Set("id", strconv.Itoa(id))

	routers := make(map[string]chi.Router)

	routers["/script"] = s.handler.Routes()

	r := router.MakeRoutes("/test/api", routers)

	recorder := httptest.NewRecorder()
	r.ServeHTTP(recorder, req)

	s.Equal(http.StatusOK, recorder.Result().StatusCode)

	// check that scrip deleted from db
	_, err = getScriptFromDB(s.db, id)
	s.Error(err)

	// check that process isn't stopped  // TODO:
}
