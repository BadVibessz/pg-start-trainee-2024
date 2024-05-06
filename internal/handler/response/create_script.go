package response

type CreateScript struct {
	ID      int    `json:"id"`
	Command string `json:"command"`
	PID     int    `json:"pid"`
}
