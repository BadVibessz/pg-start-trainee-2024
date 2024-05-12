package script

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/jmoiron/sqlx"
	"math"
	"net/http"
	"net/http/httptest"
	"os/exec"
	"pg-start-trainee-2024/domain/entity"
	"pg-start-trainee-2024/internal/handler/request"
	"pg-start-trainee-2024/internal/handler/response"
	"pg-start-trainee-2024/pkg/router"
	"strconv"
	"strings"
)

func getScriptFromDB(db *sqlx.DB, id int) (*entity.Script, error) {
	result := db.QueryRowxContext(context.Background(), "SELECT * FROM script WHERE id = $1", id)

	if err := result.Err(); err != nil {
		return nil, err
	}

	var script entity.Script

	if err := result.StructScan(&script); err != nil {
		return nil, err
	}

	return &script, nil
}

func getAllScriptsFromDB(db *sqlx.DB, offset, limit int) ([]*entity.Script, error) {
	query := "SELECT * FROM script ORDER BY created_at"

	if limit == math.MaxInt64 {
		query = fmt.Sprintf(`%v OFFSET %v`, query, offset)
	} else {
		query = fmt.Sprintf(`%v LIMIT %v OFFSET %v`, query, limit, offset)
	}

	rows, err := db.QueryxContext(context.Background(), query)
	if err != nil {
		return nil, err
	}

	var scripts []*entity.Script

	for rows.Next() {
		var script entity.Script

		if err = rows.StructScan(&script); err != nil {
			return nil, err
		}

		scripts = append(scripts, &script)
	}

	return scripts, nil
}

func stopScript(cache Cache, id int) {
	if cmdContextAny, exist := cache.Get(strconv.Itoa(id)); exist {
		if cmdContext, ok := cmdContextAny.(entity.CmdContext); ok {
			cmdContext.Cancel()
			_ = cmdContext.Cmd.Wait()
		}
	}
}

func deleteScriptFromDB(db *sqlx.DB, id int) error {
	_, err := db.ExecContext(context.Background(), fmt.Sprintf("DELETE FROM script WHERE id = %v", id))
	if err != nil {
		return err
	}

	return nil
}

func (s *Suite) checkThatScriptCreatedInDB(resp *response.CreateScript) {
	script, err := getScriptFromDB(s.db, resp.ID)
	s.NoError(err)

	s.Equal(script.ID, resp.ID)
	s.Equal(script.PID, resp.PID)
	s.Equal(script.Command, resp.Command)
}

func (s *Suite) runCheckPidExistsScript(pid int, expectedOutput string) {
	cmd := exec.Command("/bin/sh", "./check_pid_exists.sh", strconv.Itoa(pid))
	output, err := cmd.Output()
	s.NoError(err)

	s.Equal(expectedOutput, strings.Trim(string(output), "\n"))
}

func (s *Suite) TestCreateEmptyScript() {
	scriptsBefore, err := getAllScriptsFromDB(s.db, 0, math.MaxInt64)
	s.NoError(err)

	command := ""

	body, err := json.Marshal(request.CreateScript{Command: command})
	s.NoError(err)

	req, err := http.NewRequest("POST", "/test/api/script", bytes.NewBuffer(body))
	s.NoError(err)

	req.Header.Set("Content-type", "application/json")

	routers := make(map[string]chi.Router)

	routers["/script"] = s.handler.Routes()

	r := router.MakeRoutes("/test/api", routers)

	recorder := httptest.NewRecorder()
	r.ServeHTTP(recorder, req)

	s.Equal(http.StatusBadRequest, recorder.Result().StatusCode)

	// check that script not added to db: for this check that count of rows in db not changed
	scriptsAfter, err := getAllScriptsFromDB(s.db, 0, math.MaxInt64)
	s.NoError(err)

	s.Equal(len(scriptsBefore), len(scriptsAfter))
}

func (s *Suite) TestCreateShortScript() {
	command := "ls -la /"

	body, err := json.Marshal(request.CreateScript{Command: command})
	s.NoError(err)

	req, err := http.NewRequest("POST", "/test/api/script", bytes.NewBuffer(body))
	s.NoError(err)

	req.Header.Set("Content-type", "application/json")

	routers := make(map[string]chi.Router)

	routers["/script"] = s.handler.Routes()

	r := router.MakeRoutes("/test/api", routers)

	recorder := httptest.NewRecorder()
	r.ServeHTTP(recorder, req)

	s.Equal(http.StatusOK, recorder.Result().StatusCode)

	var resp response.CreateScript
	s.NoError(json.Unmarshal([]byte(recorder.Body.String()), &resp))

	s.Equal(command, resp.Command)

	// check that script created in db
	s.checkThatScriptCreatedInDB(&resp)

	// check that process is not running as this is short process
	s.runCheckPidExistsScript(resp.PID, "")

	// delete script from db
	_ = deleteScriptFromDB(s.db, resp.ID)
}

func (s *Suite) TestCreateLongScript() {
	command := "ping google.com"

	body, err := json.Marshal(request.CreateScript{Command: command})
	s.NoError(err)

	req, err := http.NewRequest("POST", "/test/api/script", bytes.NewBuffer(body))
	s.NoError(err)

	req.Header.Set("Content-type", "application/json")

	routers := make(map[string]chi.Router)

	routers["/script"] = s.handler.Routes()

	r := router.MakeRoutes("/test/api", routers)

	recorder := httptest.NewRecorder()
	r.ServeHTTP(recorder, req)

	s.Equal(http.StatusOK, recorder.Result().StatusCode)

	var resp response.CreateScript
	s.NoError(json.Unmarshal([]byte(recorder.Body.String()), &resp))

	s.Equal(command, resp.Command)

	// check that script created in db
	s.checkThatScriptCreatedInDB(&resp)

	// check that process started
	s.runCheckPidExistsScript(resp.PID, "running")

	// stop created script
	stopScript(s.cache, resp.ID)

	// delete script from db
	_ = deleteScriptFromDB(s.db, resp.ID)
}

func (s *Suite) TestCreateInvalidScript() {
	command := "abcdsfgdsdafadh"

	body, err := json.Marshal(request.CreateScript{Command: command})
	s.NoError(err)

	req, err := http.NewRequest("POST", "/test/api/script", bytes.NewBuffer(body))
	s.NoError(err)

	req.Header.Set("Content-type", "application/json")

	routers := make(map[string]chi.Router)

	routers["/script"] = s.handler.Routes()

	r := router.MakeRoutes("/test/api", routers)

	recorder := httptest.NewRecorder()
	r.ServeHTTP(recorder, req)

	s.Equal(http.StatusOK, recorder.Result().StatusCode)

	var resp response.CreateScript
	s.NoError(json.Unmarshal([]byte(recorder.Body.String()), &resp))

	s.Equal(command, resp.Command)

	// check that script created in db
	s.checkThatScriptCreatedInDB(&resp)

	// check that process is not running as this is short process
	s.runCheckPidExistsScript(resp.PID, "")

	// delete script from db
	_ = deleteScriptFromDB(s.db, resp.ID)
}
