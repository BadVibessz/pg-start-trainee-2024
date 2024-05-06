package script

import (
	"context"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
	"net/http"
	"pg-start-trainee-2024/domain/entity"
	"pg-start-trainee-2024/internal/handler/mapper"
	"pg-start-trainee-2024/internal/handler/request"

	"github.com/go-chi/render"

	handlerutils "pg-start-trainee-2024/pkg/utils/handler"
)

type Service interface {
	CreateScript(ctx context.Context, script entity.Script) (*entity.Script, error)
}

type Middleware = func(http.Handler) http.Handler

type Handler struct {
	Service     Service
	Middlewares []Middleware

	logger    *logrus.Logger
	validator *validator.Validate
}

func New(service Service, logger *logrus.Logger, validator *validator.Validate, middlewares ...Middleware) *Handler {
	return &Handler{
		Service:     service,
		Middlewares: middlewares,
		logger:      logger,
		validator:   validator,
	}
}

func (h *Handler) Routes() *chi.Mux {
	router := chi.NewRouter()

	router.Group(func(r chi.Router) {
		r.Use(h.Middlewares...)

		r.Post("/", h.CreateScript)
	})

	return router
}

// CreateScript godoc
//
//	@Summary		Create and run new script
//	@Description	Create and run new script
//	@Security		JWT
//	@Tags			Banner
//	@Accept			json
//	@Produce		json
//	@Param			input	body		request.CreateScript	true	"create script schema"
//	@Success		200		{object}	response.CreateScript
//	@Failure		401		{string}	Unauthorized
//	@Failure		403		{string}	Forbidden
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

	// TODO: req context is cancelled when req is processed -> osutils.RunCommand is cancelled -> script also cancelled, for this func we need another context
	created, err := h.Service.CreateScript(req.Context(), mapper.MapCreateScriptRequestToEntity(&scriptReq))
	if err != nil {
		msg := fmt.Sprintf("error occurred creating script: %v", err)

		handlerutils.WriteErrResponseAndLog(rw, h.logger, http.StatusBadRequest, msg, msg)

		return
	}

	render.JSON(rw, req, mapper.MapScriptToCreateScriptResponse(created))
	rw.WriteHeader(http.StatusCreated)
}
