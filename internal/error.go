package internal

import (
	"fmt"
	"net/http"
	"runtime"

	"github.com/go-playground/validator"
	"github.com/labstack/echo/v4"
)

type (
	errorLocation struct {
		File string
		Line int
	}
	customError struct {
		Location *errorLocation
		Message  string
		Original error
	}
	Error customError
)

const (
	ErrBEEmail           = "Error occurred while sending the invite email"
	ErrBEHashSalt        = "Error occurred while generating salt for hashing the given password"
	ErrBEInvalidInvite   = "The given invite token is no longer available"
	ErrBEInvalidPassword = "The given email or password do not match any available data"
	ErrBEMongoIDCast     = "Error occurred while casting MongoDB ObjectID"
	ErrBEMongoIDEmpty    = "The given MongoDB ObjectID is empty"
	ErrBENotActive       = "The account trying to log in is not activated"
	ErrBENotAdmin        = "This logged in account does not have permissions in this section"
	ErrBETimeConversion  = "Error occurred while converting a time field"
	ErrBEUserExists      = "A user account is already associated to this email"

	ErrBEQPInvalidChartType    = "The current request has an invalid or empty chartType query parameter"
	ErrBEQPInvalidDateTime     = "The current request has an invalid or empty date query parameter"
	ErrBEQPInvalidIsInside     = "The current request has an invalid or empty isInside query parameter"
	ErrBEQPInvalidIntervalType = "The current request has an invalid or empty intervalType query parameter"
	ErrBEQPInvalidLocation     = "The current request has an invalid or empty location query parameter"
	ErrBEQPInvalidMobile       = "The current request has an invalid or empty mobile query parameter"
	ErrBEQPInvalidTimezone     = "The current request has an invalid or empty timezone query parameter"
	ErrBEQPMissing             = "The current request is missing one or more query parameters"
	ErrBEQPNoRawOnGate         = "The current request is trying to retrieve non-existing raw data on gates"

	ErrDBCursorClose   = "Error occurred while closing the MongoDB cursor"
	ErrDBCursorIterate = "Error occurred while iterating over the MongoDB cursor"
	ErrDBDecode        = "Error occurred while decoding MongoDB documents"
	ErrDBDelete        = "Error occurred while deleting MongoDB documents"
	ErrDBInsert        = "Error occurred while inserting MongoDB documents"
	ErrDBQuery         = "Error occurred while querying MongoDB documents"
	ErrDBUpdate        = "Error occurred while updating MongoDB documents"
	ErrDBNoData        = "No data found to be grabbed in MongoDB query"
	ErrDBNoUpdate      = "No data found to be updated in MongoDB query"
)

// Create a new backend Error
func NewError(message string, original error, skip int) *Error {
	// Generate the error location
	_, file, line, _ := runtime.Caller(skip)

	return &Error{
		Location: &errorLocation{
			File: file,
			Line: line,
		},
		Message:  message,
		Original: original,
	}
}

// Error function for error interface
func (e *Error) Error() string {
	return e.Message
}

// Error handler with a custom response for the Echo instance
func ErrorHandler(err error, c echo.Context) {
	// Check if response not already sent to requester
	if !c.Response().Committed {
		// Create the default values for response code and message
		code := http.StatusInternalServerError
		message := "Unhandled error"

		switch e := err.(type) {
		case *echo.HTTPError: // Handle an Echo error
			code = e.Code
			message = e.Message.(string)

		case validator.ValidationErrors: // Handle a validator error
			code = http.StatusBadRequest
			message = ""

			// Customize the response message
			for _, v := range e {
				message += fmt.Sprintf("field validation for '%s' failed on the '%s' tag;", v.Field(), v.ActualTag())
			}

		case *Error: // Handle a backend error
			// Log the error
			c.Logger().Errorf(
				"BackendError :: File:%s - Line:%d :: %s -> %v", e.Location.File, e.Location.Line, e.Message, e.Original,
			)

			// Construct the response
			message = e.Message

			switch message {
			case ErrDBNoData, ErrDBNoUpdate:
				code = http.StatusNotFound
			case ErrBEInvalidPassword, ErrBENotAdmin:
				code = http.StatusUnauthorized
			case ErrBEEmail, ErrBEHashSalt, ErrBEMongoIDCast, ErrBETimeConversion,
				ErrDBCursorClose, ErrDBCursorIterate, ErrDBDecode, ErrDBDelete, ErrDBInsert, ErrDBQuery, ErrDBUpdate:
				code = http.StatusInternalServerError
			case ErrBEInvalidInvite, ErrBEMongoIDEmpty, ErrBEUserExists,
				ErrBEQPInvalidChartType, ErrBEQPInvalidDateTime, ErrBEQPInvalidIsInside, ErrBEQPInvalidIntervalType,
				ErrBEQPInvalidLocation, ErrBEQPInvalidMobile, ErrBEQPInvalidTimezone, ErrBEQPMissing, ErrBEQPNoRawOnGate:
				code = http.StatusBadRequest
			}
		}

		// Send the error response
		if err := c.JSON(code, map[string]interface{}{
			"error":   true,
			"message": message,
			"data":    nil,
		}); err != nil {
			c.Logger().Error(err)
		}
	}
}
