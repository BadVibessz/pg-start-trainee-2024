package script

import (
	"context"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"

	"pg-start-trainee-2024/domain/entity"
	"pg-start-trainee-2024/internal/handler/mapper"
	"pg-start-trainee-2024/internal/handler/request"

	handlerinternalutils "pg-start-trainee-2024/internal/pkg/utils/handler"
	handlerutils "pg-start-trainee-2024/pkg/utils/handler"
	sliceutils "pg-start-trainee-2024/pkg/utils/slice"
)

type Service interface {
	CreateScript(ctx context.Context, script entity.Script) (*entity.Script, error)
	StopScript(ctx context.Context, id int) error
	GetScript(ctx context.Context, id int) (*entity.Script, error)
	GetAllScripts(ctx context.Context, offset, limit int) ([]*entity.Script, error)
	DeleteScript(ctx context.Context, id int) error
}

type Middleware = func(http.Handler) http.Handler

type Handler struct {
	Service     Service
	Middlewares []Middleware

	logger        *logrus.Logger
	validator     *validator.Validate
	defaultOffset int
	defaultLimit  int
}

func New(service Service, logger *logrus.Logger, validator *validator.Validate, defaultOffset, defaultLimit int, middlewares ...Middleware) *Handler {
	return &Handler{
		Service:       service,
		Middlewares:   middlewares,
		logger:        logger,
		validator:     validator,
		defaultOffset: defaultOffset,
		defaultLimit:  defaultLimit,
	}
}

func (h *Handler) Routes() *chi.Mux {
	router := chi.NewRouter()

	router.Group(func(r chi.Router) {
		r.Use(h.Middlewares...)

		r.Post("/", h.CreateScript)
		r.Patch("/", h.StopScript)
		r.Get("/", h.GetScript)
		r.Get("/all", h.GetAllScripts)
		r.Delete("/", h.DeleteScript)
	})

	return router
}

// CreateScript godoc
//
//	@Summary		Create and run new script
//	@Description	Create and run new script
//	@Tags			Script
//	@Accept			json
//	@Produce		json
//	@Param			input	body		request.CreateScript	true	"create script schema"
//	@Success		200		{object}	response.CreateScript
//	@Failure		401		{string}	Unauthorized
//	@Failure		400		{string}	invalid		request
//	@Failure		500		{string}	internal	error
//	@Router			/pg-start-trainee/api/v1/script [post]
func (h *Handler) CreateScript(rw http.ResponseWriter, req *http.Request) {
	var scriptReq request.CreateScript

	if err := render.DecodeJSON(req.Body, &scriptReq); err != nil {
		msg := fmt.Sprintf("error occurred decoding request body to CreateScript request: %v", err)

		handlerutils.WriteErrResponseAndLog(rw, h.logger, http.StatusBadRequest, msg, msg)

		return
	}

	if err := scriptReq.Validate(h.validator); err != nil {
		msg := fmt.Sprintf("error occurred validating CreateScript request: %v", err)

		handlerutils.WriteErrResponseAndLog(rw, h.logger, http.StatusBadRequest, msg, msg)

		return
	}

	created, err := h.Service.CreateScript(req.Context(), mapper.MapCreateScriptRequestToEntity(&scriptReq))
	if err != nil {
		msg := fmt.Sprintf("error occurred creating script: %v", err)

		handlerutils.WriteErrResponseAndLog(rw, h.logger, http.StatusBadRequest, msg, msg)
		return
	}

	render.JSON(rw, req, mapper.MapScriptToCreateScriptResponse(created))
	rw.WriteHeader(http.StatusOK)
}

// StopScript godoc
//
//	@Summary		Stop running script
//	@Description	Stop running script
//	@Tags			Script
//	@Accept			json
//	@Produce		json
//	@Param			id	header	int	true	"script ID"
//	@Success		200
//	@Failure		401	{string}	Unauthorized
//	@Failure		400	{string}	invalid		request
//	@Failure		500	{string}	internal	error
//	@Router			/pg-start-trainee/api/v1/script [patch]
func (h *Handler) StopScript(rw http.ResponseWriter, req *http.Request) {
	id, err := handlerutils.GetIntHeaderByKey(req, "id")
	if err != nil {
		msg := fmt.Sprintf("no id header provided: %v", err)

		handlerutils.WriteErrResponseAndLog(rw, h.logger, http.StatusBadRequest, msg, msg)
		return
	}

	if err = h.Service.StopScript(req.Context(), id); err != nil {
		msg := fmt.Sprintf("error occurred stopping script: %v", err)

		handlerutils.WriteErrResponseAndLog(rw, h.logger, http.StatusBadRequest, msg, msg)
		return
	}

	rw.WriteHeader(http.StatusOK)

	_, err = rw.Write([]byte("script successfully stopped."))
	if err != nil {
		h.logger.Errorf("error occurred writing response: %v", err)
	}
}

// GetScript godoc
//
//	@Summary		Get script
//	@Description	Get script
//	@Tags			Script
//	@Accept			json
//	@Produce		json
//	@Param			id	header		int	true	"script ID"
//	@Success		200	{object}	response.GetScript
//	@Failure		401	{string}	Unauthorized
//	@Failure		400	{string}	invalid		request
//	@Failure		500	{string}	internal	error
//	@Router			/pg-start-trainee/api/v1/script [get]
func (h *Handler) GetScript(rw http.ResponseWriter, req *http.Request) {
	id, err := handlerutils.GetIntHeaderByKey(req, "id")
	if err != nil {
		msg := fmt.Sprintf("no id header provided: %v", err)

		handlerutils.WriteErrResponseAndLog(rw, h.logger, http.StatusBadRequest, msg, msg)
		return
	}

	script, err := h.Service.GetScript(req.Context(), id)
	if err != nil {
		msg := fmt.Sprintf("error occurred fetching script: %v", err)

		handlerutils.WriteErrResponseAndLog(rw, h.logger, http.StatusBadRequest, msg, msg)
		return
	}

	render.JSON(rw, req, mapper.MapScriptToGetScriptResponse(script))
	rw.WriteHeader(http.StatusOK)
}

// GetAllScripts godoc
//
//	@Summary		Get all scripts
//	@Description	Get all scripts
//	@Tags			Script
//	@Accept			json
//	@Produce		json
//	@Param			offset	query		int	false	"Offset"
//	@Param			limit	query		int	false	"Limit"
//	@Success		200		{object}	[]response.GetScript
//	@Failure		400		{string}	invalid		request
//	@Failure		500		{string}	internal	error
//	@Router			/pg-start-trainee/api/v1/script/all [get]
func (h *Handler) GetAllScripts(rw http.ResponseWriter, req *http.Request) {
	paginationOpts := handlerinternalutils.GetPaginationOptsFromQuery(req, h.defaultOffset, h.defaultLimit)

	if err := paginationOpts.Validate(h.validator); err != nil {
		msg := fmt.Sprintf("invalid pagination options provided: %v", err)

		handlerutils.WriteErrResponseAndLog(rw, h.logger, http.StatusBadRequest, msg, msg)

		return
	}

	scripts, err := h.Service.GetAllScripts(req.Context(), paginationOpts.Offset, paginationOpts.Limit)
	if err != nil {
		msg := fmt.Sprintf("error occurred fetching scripts: %v", err)

		handlerutils.WriteErrResponseAndLog(rw, h.logger, http.StatusBadRequest, msg, msg)
		return
	}

	render.JSON(rw, req, sliceutils.Map(scripts, mapper.MapScriptToGetScriptResponse))
	rw.WriteHeader(http.StatusOK)
}

// DeleteScript godoc
//
//	@Summary		Delete script by ID
//	@Description	Delete script by ID
//	@Tags			Script
//	@Produce		json
//	@Param			id	header	int	true	"script ID"
//	@Success		200
//	@Failure		400	{string}	invalid		request
//	@Failure		500	{string}	internal	error
//	@Router			/pg-start-trainee/api/v1/script [delete]
func (h *Handler) DeleteScript(rw http.ResponseWriter, req *http.Request) {
	id, err := handlerutils.GetIntHeaderByKey(req, "id")
	if err != nil {
		msg := fmt.Sprintf("no id header provided: %v", err)

		handlerutils.WriteErrResponseAndLog(rw, h.logger, http.StatusBadRequest, msg, msg)
		return
	}

	err = h.Service.DeleteScript(req.Context(), id)
	if err != nil {
		msg := fmt.Sprintf("error occurred deleting script: %v", err)

		handlerutils.WriteErrResponseAndLog(rw, h.logger, http.StatusBadRequest, msg, msg)
		return
	}

	rw.WriteHeader(http.StatusOK)

	_, err = rw.Write([]byte("script successfully deleted."))
	if err != nil {
		h.logger.Errorf("error occurred writing response: %v", err)
	}
}
