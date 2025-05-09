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
	Unauthorized         ErrorCode = "K0004"

	// Category 01: M2M

	// Category 02: Users
	MissingCookie    ErrorCode = "K0201"
	UserNotFound     ErrorCode = "K0202"
	AlreadyLoggingIn ErrorCode = "K0203"

	// Category 03: JWKS
	InvalidServiceId ErrorCode = "K0301"

	// Category 04: Instance settings
	SettingNotFound ErrorCode = "K0401"

	// Category 05: Services
	ServiceNotFound ErrorCode = "K0501"
)

var errorDetailsMap = map[ErrorCode]ErrorDetail{
	// Category 00: General
	InternalServerError: {
		Code:        InternalServerError,
		StatusCode:  http.StatusInternalServerError,
		Description: "Internal server error",
	},
	Timeout: {
		Code:        Timeout,
		StatusCode:  http.StatusRequestTimeout,
		Description: "Request timed out",
	},
	InvalidArgumentError: {
		Code:        InvalidArgumentError,
		StatusCode:  http.StatusBadRequest,
		Description: "Invalid request body",
	},
	Unauthorized: {
		Code:        Unauthorized,
		StatusCode:  http.StatusUnauthorized,
		Description: "Unauthorized",
	},

	// Category 01: M2M

	// Category 02: Users
	MissingCookie: {
		Code:        MissingCookie,
		StatusCode:  http.StatusBadRequest,
		Description: "Missing one or more required cookies.",
	},
	UserNotFound: {
		Code:        UserNotFound,
		StatusCode:  http.StatusNotFound,
		Description: "User not found.",
	},
	AlreadyLoggingIn: {
		Code:        AlreadyLoggingIn,
		StatusCode:  http.StatusConflict,
		Description: "You're already trying to login from another device.",
	},

	// Category 03: JWKS
	InvalidServiceId: {
		Code:        InvalidServiceId,
		StatusCode:  http.StatusBadRequest,
		Description: "The service_id is not a valid uuid.",
	},

	// Category 04: Instance settings
	SettingNotFound: {
		Code:        SettingNotFound,
		StatusCode:  http.StatusNotFound,
		Description: "Setting not found.",
	},

	// Category 05: Services
	ServiceNotFound: {
		Code:        ServiceNotFound,
		StatusCode:  http.StatusNotFound,
		Description: "Service not found.",
	},
}
