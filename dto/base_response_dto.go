package dto

type BaseResponseSuccess struct {
	Success bool   `json:"success" example:"true"`
	Message string `json:"message"`
}

type BaseResponseSuccessWithData struct {
	Success bool   `json:"success" example:"true"`
	Message string `json:"message"`
	Data    any    `json:"data"`
}

type BaseResponseError struct {
	Success bool   `json:"success" example:"false"`
	Message string `json:"message"`
}
