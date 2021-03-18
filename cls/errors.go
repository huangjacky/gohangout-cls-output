package cls

import (
	"fmt"
	"net/http"
)

type ErrorResponse struct {
	// HTTP response that caused this error
	Response *http.Response

	ErrorCode    string `json:"errorcode"`
	ErrorMessage string `json:"errormessage"`
}

type errorCode string

const (
	Success                 errorCode = "Success"
	ErrInternalError        errorCode = "InternalError"
	ErrTopicConflict        errorCode = "TopicConflict"
	ErrTopicNotExist        errorCode = "TopicNotExist"
	ErrInvalidContentType   errorCode = "InvalidContentType"
	ErrInvalidAuthorization errorCode = "InvalidAuthorization"
	ErrInvalidContent       errorCode = "InvalidContent"
	ErrInvalidParam         errorCode = "InvalidParam"
	ErrMissingAgentIp       errorCode = "MissingAgentIp"
	ErrMissingAgentVersion  errorCode = "MissingAgentVersion"
	ErrMissingAuthorization errorCode = "MissingAuthorization"
	ErrMissingContent       errorCode = "MissingContent"
	ErrMissingContentType   errorCode = "MissingContentType"
	ErrTopicClosed          errorCode = "TopicClosed"
	ErrIndexRuleEmpty       errorCode = "IndexRuleEmpty"
	ErrLogsetNotEmpty       errorCode = "LogsetNotEmpty"
	ErrSyntaxError          errorCode = "SyntaxError"
	ErrLogsetEmpty          errorCode = "LogsetEmpty"
	ErrUnauthorized         errorCode = "Unauthorized"
	ErrLogsetExceed         errorCode = "LogsetExceed"
	ErrLogSizeExceed        errorCode = "LogSizeExceed"
	ErrMachineGroupExceed   errorCode = "MachineGroupExceed"
	ErrNotAllowed           errorCode = "NotAllowed"
	ErrTopicExceed          errorCode = "TopicExceed"
	ErrShipperExceed        errorCode = "ShipperExceed"
	ErrTaskReadOnly         errorCode = "TaskReadOnly"
	ErrCursorNotExist       errorCode = "CursorNotExist"
	ErrTaskNotExist         errorCode = "TaskNotExist"
	ErrIndexNotExist        errorCode = "IndexNotExist"
	ErrLogsetNotExist       errorCode = "LogsetNotExist"
	ErrMachineGroupNotExist errorCode = "MachineGroupNotExist"
	ErrShipperNotExist      errorCode = "ShipperNotExist"
	ErrConsumerNotExist     errorCode = "ConsumerNotExist"
	ErrNotSupported         errorCode = "NotSupported"
	ErrIndexConflict        errorCode = "IndexConflict"
	ErrLogsetConflict       errorCode = "LogsetConflict"
	ErrMachineGroupConflict errorCode = "MachineGroupConflict"
	ErrShipperConflict      errorCode = "ShipperConflict"
	ErrConsumerConflict     errorCode = "ConsumerConflict"
	ErrSpeedQuotaExceed     errorCode = "SpeedQuotaExceed"
)

func (e errorCode) String() string {
	return string(e)
}

func (r *ErrorResponse) Error() string {
	return fmt.Sprintf("%v %s: %d %s %s",
		r.Response.Request.Method, r.Response.Request.URL,
		r.Response.StatusCode, r.ErrorCode, r.ErrorMessage,
	)
}

func ErrorCode(err error) errorCode {
	if err == nil {
		return Success
	}
	val, ok := err.(*ErrorResponse)
	if !ok {
		return ErrInternalError
	}

	switch val.ErrorCode {
	case ErrInvalidContentType.String():
		return ErrInvalidContentType
	case ErrInvalidAuthorization.String():
		return ErrInvalidAuthorization
	case ErrInvalidContent.String():
		return ErrInvalidContent
	case ErrInvalidParam.String():
		return ErrInvalidParam
	case ErrMissingAgentIp.String():
		return ErrMissingAgentIp
	case ErrMissingAgentVersion.String():
		return ErrMissingAgentVersion
	case ErrMissingAuthorization.String():
		return ErrMissingAuthorization
	case ErrMissingContent.String():
		return ErrMissingContent
	case ErrMissingContentType.String():
		return ErrMissingContentType
	case ErrTopicClosed.String():
		return ErrTopicClosed
	case ErrIndexRuleEmpty.String():
		return ErrIndexRuleEmpty
	case ErrLogsetNotEmpty.String():
		return ErrLogsetNotEmpty
	case ErrSyntaxError.String():
		return ErrSyntaxError
	case ErrLogsetEmpty.String():
		return ErrLogsetEmpty
	case ErrUnauthorized.String():
		return ErrUnauthorized
	case ErrLogsetExceed.String():
		return ErrLogsetExceed
	case ErrLogSizeExceed.String():
		return ErrLogSizeExceed
	case ErrMachineGroupExceed.String():
		return ErrMachineGroupExceed
	case ErrNotAllowed.String():
		return ErrNotAllowed
	case ErrTopicExceed.String():
		return ErrTopicExceed
	case ErrShipperExceed.String():
		return ErrShipperExceed
	case ErrTaskReadOnly.String():
		return ErrTaskReadOnly
	case ErrCursorNotExist.String():
		return ErrCursorNotExist
	case ErrTaskNotExist.String():
		return ErrTaskNotExist
	case ErrIndexNotExist.String():
		return ErrIndexNotExist
	case ErrLogsetNotExist.String():
		return ErrLogsetNotExist
	case ErrMachineGroupNotExist.String():
		return ErrMachineGroupNotExist
	case ErrShipperNotExist.String():
		return ErrShipperNotExist
	case ErrConsumerNotExist.String():
		return ErrConsumerNotExist
	case ErrNotSupported.String():
		return ErrNotSupported
	case ErrMachineGroupConflict.String():
		return ErrMachineGroupConflict
	case ErrLogsetConflict.String():
		return ErrLogsetConflict
	case ErrIndexConflict.String():
		return ErrIndexConflict
	case ErrSpeedQuotaExceed.String():
		return ErrSpeedQuotaExceed
	case ErrConsumerConflict.String():
		return ErrConsumerConflict
	case ErrShipperConflict.String():
		return ErrShipperConflict
	case ErrInternalError.String():
		return ErrInternalError
	case ErrTopicNotExist.String():
		return ErrTopicNotExist
	case ErrTopicConflict.String():
		return ErrTopicConflict
	default:
		return ErrInternalError
	}
}

func IsInternalError(err error) bool {
	val, ok := err.(*ErrorResponse)
	// may be http error
	if !ok || val.ErrorCode == ErrInternalError.String() {
		return true
	}
	return false
}

func IsTopicConfictError(err error) bool {
	val, ok := err.(*ErrorResponse)
	if ok && val.ErrorCode == ErrTopicConflict.String() {
		return true
	}
	return false
}

func IsTopicNotExistError(err error) bool {
	val, ok := err.(*ErrorResponse)
	if ok && val.ErrorCode == ErrTopicNotExist.String() {
		return true
	}
	return false
}
