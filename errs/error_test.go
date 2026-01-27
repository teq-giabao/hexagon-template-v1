package errs_test

import (
	"errors"
	"testing"

	"hexagon/errs"
)

func TestError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *errs.Error
		expected string
	}{
		{
			name: "basic error",
			err: &errs.Error{
				Code:    errs.EINVALID,
				Message: "invalid input",
			},
			expected: "application error: code=invalid message=invalid input",
		},
		{
			name: "conflict error",
			err: &errs.Error{
				Code:    errs.ECONFLICT,
				Message: "resource already exists",
			},
			expected: "application error: code=conflict message=resource already exists",
		},
		{
			name: "empty message",
			err: &errs.Error{
				Code:    errs.EINTERNAL,
				Message: "",
			},
			expected: "application error: code=internal message=",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.err.Error()
			if got != tt.expected {
				t.Errorf("Error() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestErrorCode(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected string
	}{
		{
			name:     "nil error returns empty string",
			err:      nil,
			expected: "",
		},
		{
			name: "application error returns its code",
			err: &errs.Error{
				Code:    errs.EINVALID,
				Message: "invalid input",
			},
			expected: errs.EINVALID,
		},
		{
			name: "not found error",
			err: &errs.Error{
				Code:    errs.ENOTFOUND,
				Message: "resource not found",
			},
			expected: errs.ENOTFOUND,
		},
		{
			name: "conflict error",
			err: &errs.Error{
				Code:    errs.ECONFLICT,
				Message: "already exists",
			},
			expected: errs.ECONFLICT,
		},
		{
			name: "unauthorized error",
			err: &errs.Error{
				Code:    errs.EUNAUTHORIZED,
				Message: "not authorized",
			},
			expected: errs.EUNAUTHORIZED,
		},
		{
			name: "not implemented error",
			err: &errs.Error{
				Code:    errs.ENOTIMPLEMENTED,
				Message: "feature not available",
			},
			expected: errs.ENOTIMPLEMENTED,
		},
		{
			name: "internal error",
			err: &errs.Error{
				Code:    errs.EINTERNAL,
				Message: "internal error",
			},
			expected: errs.EINTERNAL,
		},
		{
			name:     "non-application error returns EINTERNAL",
			err:      errors.New("standard error"),
			expected: errs.EINTERNAL,
		},
		{
			name:     "wrapped application error",
			err:      errors.Join(&errs.Error{Code: errs.EINVALID, Message: "bad request"}),
			expected: errs.EINVALID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := errs.ErrorCode(tt.err)
			if got != tt.expected {
				t.Errorf("ErrorCode() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestErrorMessage(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected string
	}{
		{
			name:     "nil error returns empty string",
			err:      nil,
			expected: "",
		},
		{
			name: "application error returns its message",
			err: &errs.Error{
				Code:    errs.EINVALID,
				Message: "invalid input provided",
			},
			expected: "invalid input provided",
		},
		{
			name: "error with empty message",
			err: &errs.Error{
				Code:    errs.EINTERNAL,
				Message: "",
			},
			expected: "",
		},
		{
			name: "error with multi-line message",
			err: &errs.Error{
				Code:    errs.EINVALID,
				Message: "validation failed:\n- field1 is required\n- field2 is invalid",
			},
			expected: "validation failed:\n- field1 is required\n- field2 is invalid",
		},
		{
			name:     "non-application error returns Internal error",
			err:      errors.New("disk write error"),
			expected: "Internal error.",
		},
		{
			name:     "wrapped application error",
			err:      errors.Join(&errs.Error{Code: errs.ENOTFOUND, Message: "user not found"}),
			expected: "user not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := errs.ErrorMessage(tt.err)
			if got != tt.expected {
				t.Errorf("ErrorMessage() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestErrorf(t *testing.T) {
	tests := []struct {
		name          string
		code          string
		format        string
		args          []interface{}
		expectedCode  string
		expectedMsg   string
		expectedError string
	}{
		{
			name:          "simple message without formatting",
			code:          errs.EINVALID,
			format:        "invalid request",
			args:          nil,
			expectedCode:  errs.EINVALID,
			expectedMsg:   "invalid request",
			expectedError: "application error: code=invalid message=invalid request",
		},
		{
			name:          "formatted message with single argument",
			code:          errs.ENOTFOUND,
			format:        "user %s not found",
			args:          []interface{}{"john"},
			expectedCode:  errs.ENOTFOUND,
			expectedMsg:   "user john not found",
			expectedError: "application error: code=not_found message=user john not found",
		},
		{
			name:          "formatted message with multiple arguments",
			code:          errs.ECONFLICT,
			format:        "duplicate entry: id=%d, name=%s",
			args:          []interface{}{123, "test"},
			expectedCode:  errs.ECONFLICT,
			expectedMsg:   "duplicate entry: id=123, name=test",
			expectedError: "application error: code=conflict message=duplicate entry: id=123, name=test",
		},
		{
			name:          "internal error code",
			code:          errs.EINTERNAL,
			format:        "database connection failed",
			args:          nil,
			expectedCode:  errs.EINTERNAL,
			expectedMsg:   "database connection failed",
			expectedError: "application error: code=internal message=database connection failed",
		},
		{
			name:          "unauthorized error code",
			code:          errs.EUNAUTHORIZED,
			format:        "token expired at %s",
			args:          []interface{}{"2024-01-01"},
			expectedCode:  errs.EUNAUTHORIZED,
			expectedMsg:   "token expired at 2024-01-01",
			expectedError: "application error: code=unauthorized message=token expired at 2024-01-01",
		},
		{
			name:          "not implemented error code",
			code:          errs.ENOTIMPLEMENTED,
			format:        "feature %q is not available",
			args:          []interface{}{"export"},
			expectedCode:  errs.ENOTIMPLEMENTED,
			expectedMsg:   "feature \"export\" is not available",
			expectedError: "application error: code=not_implemented message=feature \"export\" is not available",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := errs.Errorf(tt.code, tt.format, tt.args...)

			if err.Code != tt.expectedCode {
				t.Errorf("Errorf().Code = %q, want %q", err.Code, tt.expectedCode)
			}

			if err.Message != tt.expectedMsg {
				t.Errorf("Errorf().Message = %q, want %q", err.Message, tt.expectedMsg)
			}

			if err.Error() != tt.expectedError {
				t.Errorf("Errorf().Error() = %q, want %q", err.Error(), tt.expectedError)
			}
		})
	}
}

func TestErrorCodes(t *testing.T) {
	// Test that all error code constants are defined correctly
	codes := map[string]string{
		"ECONFLICT":       errs.ECONFLICT,
		"EINTERNAL":       errs.EINTERNAL,
		"EINVALID":        errs.EINVALID,
		"ENOTFOUND":       errs.ENOTFOUND,
		"ENOTIMPLEMENTED": errs.ENOTIMPLEMENTED,
		"EUNAUTHORIZED":   errs.EUNAUTHORIZED,
	}

	expected := map[string]string{
		"ECONFLICT":       "conflict",
		"EINTERNAL":       "internal",
		"EINVALID":        "invalid",
		"ENOTFOUND":       "not_found",
		"ENOTIMPLEMENTED": "not_implemented",
		"EUNAUTHORIZED":   "unauthorized",
	}

	for name, code := range codes {
		if code != expected[name] {
			t.Errorf("constant %s = %q, want %q", name, code, expected[name])
		}
	}
}
