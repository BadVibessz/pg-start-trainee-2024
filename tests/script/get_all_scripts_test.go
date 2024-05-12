package script

import (
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"math"
	"net/http"
	"net/http/httptest"
	"pg-start-trainee-2024/domain/entity"
	"pg-start-trainee-2024/internal/handler/response"
	"pg-start-trainee-2024/pkg/router"
	"strconv"
)

func setOffsetAndLimit(req *http.Request, offset, limit int) {
	q := req.URL.Query()

	q.Set("offset", strconv.Itoa(offset))
	q.Set("limit", strconv.Itoa(limit))

	req.URL.RawQuery = q.Encode()
}

func getScriptResponseAndScriptEqual(resp *response.GetScript, script *entity.Script) bool {
	return resp.ID == script.ID &&
		resp.Command == script.Command &&
		resp.PID == script.PID &&
		resp.IsRunning == script.IsRunning &&
		resp.Output == script.Output &&
		resp.CreatedAt == script.CreatedAt &&
		resp.UpdatedAt == script.UpdatedAt
}

func (s *Suite) getScriptResponseAndScriptSlicesEqual(resp []response.GetScript, scripts []*entity.Script) {
	if !s.True(len(resp) == len(scripts)) {
		return
	}

	for i := range resp {
		if !s.True(getScriptResponseAndScriptEqual(&resp[i], scripts[i])) {
			return
		}
	}
}

func (s *Suite) TestGetScriptsNoOffsetDefaultLimit() {
	offset := 0
	limit := 0

	req, err := http.NewRequest("GET", "/test/api/script/all", nil)
	s.NoError(err)

	req.Header.Set("Content-type", "application/json")

	setOffsetAndLimit(req, offset, limit)

	routers := make(map[string]chi.Router)

	routers["/script"] = s.handler.Routes()

	r := router.MakeRoutes("/test/api", routers)

	recorder := httptest.NewRecorder()
	r.ServeHTTP(recorder, req)

	s.Equal(http.StatusOK, recorder.Result().StatusCode)

	var resp []response.GetScript
	s.NoError(json.Unmarshal([]byte(recorder.Body.String()), &resp))

	got, err := getAllScriptsFromDB(s.db, offset, s.config.Handler.DefaultLimit)
	s.NoError(err)

	// check that got in response and from db are equal
	s.getScriptResponseAndScriptSlicesEqual(resp, got)
}

func (s *Suite) TestGetScriptsNoOffsetNoLimit() {
	offset := 0
	limit := math.MaxInt64

	req, err := http.NewRequest("GET", "/test/api/script/all", nil)
	s.NoError(err)

	req.Header.Set("Content-type", "application/json")

	setOffsetAndLimit(req, offset, limit)

	routers := make(map[string]chi.Router)

	routers["/script"] = s.handler.Routes()

	r := router.MakeRoutes("/test/api", routers)

	recorder := httptest.NewRecorder()
	r.ServeHTTP(recorder, req)

	s.Equal(http.StatusOK, recorder.Result().StatusCode)

	var resp []response.GetScript
	s.NoError(json.Unmarshal([]byte(recorder.Body.String()), &resp))

	got, err := getAllScriptsFromDB(s.db, offset, limit)
	s.NoError(err)

	// check that got in response and from db are equal
	s.getScriptResponseAndScriptSlicesEqual(resp, got)
}

func (s *Suite) TestGetScriptsNoOffsetInvalidLimit() {
	offset := 0
	limit := -1

	req, err := http.NewRequest("GET", "/test/api/script/all", nil)
	s.NoError(err)

	req.Header.Set("Content-type", "application/json")

	setOffsetAndLimit(req, offset, limit)

	routers := make(map[string]chi.Router)

	routers["/script"] = s.handler.Routes()

	r := router.MakeRoutes("/test/api", routers)

	recorder := httptest.NewRecorder()
	r.ServeHTTP(recorder, req)

	s.Equal(http.StatusBadRequest, recorder.Result().StatusCode)
}

func (s *Suite) TestGetScriptsInvalidOffsetNoLimit() {
	offset := -1
	limit := math.MaxInt64

	req, err := http.NewRequest("GET", "/test/api/script/all", nil)
	s.NoError(err)

	req.Header.Set("Content-type", "application/json")

	setOffsetAndLimit(req, offset, limit)

	routers := make(map[string]chi.Router)

	routers["/script"] = s.handler.Routes()

	r := router.MakeRoutes("/test/api", routers)

	recorder := httptest.NewRecorder()
	r.ServeHTTP(recorder, req)

	s.Equal(http.StatusBadRequest, recorder.Result().StatusCode)
}

func (s *Suite) TestGetScriptsNoOffsetLimit_1() {
	offset := 0
	limit := 1

	req, err := http.NewRequest("GET", "/test/api/script/all", nil)
	s.NoError(err)

	req.Header.Set("Content-type", "application/json")

	setOffsetAndLimit(req, offset, limit)

	routers := make(map[string]chi.Router)

	routers["/script"] = s.handler.Routes()

	r := router.MakeRoutes("/test/api", routers)

	recorder := httptest.NewRecorder()
	r.ServeHTTP(recorder, req)

	s.Equal(http.StatusOK, recorder.Result().StatusCode)

	var resp []response.GetScript
	s.NoError(json.Unmarshal([]byte(recorder.Body.String()), &resp))

	got, err := getAllScriptsFromDB(s.db, offset, limit)
	s.NoError(err)

	// check that got in response and from db are equal
	s.getScriptResponseAndScriptSlicesEqual(resp, got)
}

func (s *Suite) TestGetScriptsOffset_1_NoLimit() {
	offset := 1
	limit := math.MaxInt64

	req, err := http.NewRequest("GET", "/test/api/script/all", nil)
	s.NoError(err)

	req.Header.Set("Content-type", "application/json")

	setOffsetAndLimit(req, offset, limit)

	routers := make(map[string]chi.Router)

	routers["/script"] = s.handler.Routes()

	r := router.MakeRoutes("/test/api", routers)

	recorder := httptest.NewRecorder()
	r.ServeHTTP(recorder, req)

	s.Equal(http.StatusOK, recorder.Result().StatusCode)

	var resp []response.GetScript
	s.NoError(json.Unmarshal([]byte(recorder.Body.String()), &resp))

	got, err := getAllScriptsFromDB(s.db, offset, limit)
	s.NoError(err)

	// check that got in response and from db are equal
	s.getScriptResponseAndScriptSlicesEqual(resp, got)
}

func (s *Suite) TestGetScriptsOffset_1_Limit_1() {
	offset := 1
	limit := 1

	req, err := http.NewRequest("GET", "/test/api/script/all", nil)
	s.NoError(err)

	req.Header.Set("Content-type", "application/json")

	setOffsetAndLimit(req, offset, limit)

	routers := make(map[string]chi.Router)

	routers["/script"] = s.handler.Routes()

	r := router.MakeRoutes("/test/api", routers)

	recorder := httptest.NewRecorder()
	r.ServeHTTP(recorder, req)

	s.Equal(http.StatusOK, recorder.Result().StatusCode)

	var resp []response.GetScript
	s.NoError(json.Unmarshal([]byte(recorder.Body.String()), &resp))

	got, err := getAllScriptsFromDB(s.db, offset, limit)
	s.NoError(err)

	// check that got in response and from db are equal
	s.getScriptResponseAndScriptSlicesEqual(resp, got)
}

func (s *Suite) TestGetScriptsOffsetGreaterThanCollection_NoLimit() {
	offset := math.MaxInt64
	limit := math.MaxInt64

	req, err := http.NewRequest("GET", "/test/api/script/all", nil)
	s.NoError(err)

	req.Header.Set("Content-type", "application/json")

	setOffsetAndLimit(req, offset, limit)

	routers := make(map[string]chi.Router)

	routers["/script"] = s.handler.Routes()

	r := router.MakeRoutes("/test/api", routers)

	recorder := httptest.NewRecorder()
	r.ServeHTTP(recorder, req)

	s.Equal(http.StatusOK, recorder.Result().StatusCode)

	var resp []response.GetScript
	s.NoError(json.Unmarshal([]byte(recorder.Body.String()), &resp))

	got, err := getAllScriptsFromDB(s.db, offset, limit)
	s.NoError(err)

	// check that got in response and from db are equal
	s.getScriptResponseAndScriptSlicesEqual(resp, got)
}

func (s *Suite) TestGetScriptsNoOffset_LimitGreaterThanCollection() {
	offset := 0
	limit := 1000

	req, err := http.NewRequest("GET", "/test/api/script/all", nil)
	s.NoError(err)

	req.Header.Set("Content-type", "application/json")

	setOffsetAndLimit(req, offset, limit)

	routers := make(map[string]chi.Router)

	routers["/script"] = s.handler.Routes()

	r := router.MakeRoutes("/test/api", routers)

	recorder := httptest.NewRecorder()
	r.ServeHTTP(recorder, req)

	s.Equal(http.StatusOK, recorder.Result().StatusCode)

	var resp []response.GetScript
	s.NoError(json.Unmarshal([]byte(recorder.Body.String()), &resp))

	got, err := getAllScriptsFromDB(s.db, offset, limit)
	s.NoError(err)

	// check that got in response and from db are equal
	s.getScriptResponseAndScriptSlicesEqual(resp, got)
}
