package mapper

import (
	"pg-start-trainee-2024/domain/entity"
	"pg-start-trainee-2024/internal/handler/request"
	"pg-start-trainee-2024/internal/handler/response"
)

func MapCreateScriptRequestToEntity(createRequest *request.CreateScript) entity.Script {
	return entity.Script{
		Command: createRequest.Command,
	}
}

func MapScriptToCreateScriptResponse(script *entity.Script) response.CreateScript {
	return response.CreateScript{
		ID:      script.ID,
		Command: script.Command,
		PID:     script.PID,
	}
}

func MapScriptToGetScriptResponse(script *entity.Script) response.GetScript {
	return response.GetScript{
		ID:        script.ID,
		Command:   script.Command,
		Output:    script.Output,
		IsRunning: script.IsRunning,
		PID:       script.PID,
		CreatedAt: script.CreatedAt,
		UpdatedAt: script.UpdatedAt,
	}
}
