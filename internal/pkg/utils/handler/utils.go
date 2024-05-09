package handler

import (
	"net/http"

	"pg-start-trainee-2024/internal/handler/request"

	handlerutils "pg-start-trainee-2024/pkg/utils/handler"
)

func GetPaginationOptsFromQuery(req *http.Request, defaultOffset int, defaultLimit int) request.PaginationOptions {
	offset, err := handlerutils.GetIntParamFromQuery(req, "offset")
	if offset == 0 || err != nil {
		offset = defaultOffset
	}

	limit, err := handlerutils.GetIntParamFromQuery(req, "limit")
	if limit == 0 || err != nil {
		limit = defaultLimit
	}

	paginationOpts := request.PaginationOptions{
		Offset: offset,
		Limit:  limit,
	}

	return paginationOpts
}
