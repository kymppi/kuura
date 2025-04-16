package errcode

import "net/http"

type ErrorCode string

type ErrorDetail struct {
	Code        ErrorCode
	StatusCode  int    // HTTP status code
	Description string // Short description of the error
}

func GetErrorDetail(code ErrorCode) ErrorDetail {
	if detail, exists := errorDetailsMap[code]; exists {
		return detail
	}

	return ErrorDetail{
		Code:        "K0000",
		StatusCode:  500,
		Description: "Unknown error",
	}
}

const (
	// Format: MHXXYYYY
	// XX = Category
	// YYYY = Error

	// Category 00: General
	InternalServerError  ErrorCode = "K0001"
	Timeout              ErrorCode = "K0002"
	InvalidArgumentError ErrorCode = "K0003"

	// Category 01: M2M

	// Category 02: Users
	MissingCookie ErrorCode = "K0201"

	// Category 03: JWKS
	InvalidServiceId ErrorCode = "K0301"
)

var errorDetailsMap = map[ErrorCode]ErrorDetail{
	// Category 00: General
	InternalServerError: {
		Code:        InternalServerError,
		StatusCode:  500,
		Description: "Internal server error",
	},
	Timeout: {
		Code:        Timeout,
		StatusCode:  408,
		Description: "Request timed out",
	},
	InvalidArgumentError: {
		Code:        InvalidArgumentError,
		StatusCode:  400,
		Description: "Invalid request body",
	},

	// Category 01: M2M

	// Category 02:
	MissingCookie: {
		Code:        MissingCookie,
		StatusCode:  http.StatusBadRequest,
		Description: "Missing one or more required cookies.",
	},

	// Category 03: JWKS
	InvalidServiceId: {
		Code:        InvalidServiceId,
		StatusCode:  http.StatusBadRequest,
		Description: "The service_id is not a valid uuid.",
	},
}
