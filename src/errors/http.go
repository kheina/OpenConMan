package errors

import "google.golang.org/grpc/codes"

type Status uint32

const (
	BadRequest Status = 400
	Conflict          = 409

	Internal = 500
)

func (s Status) String() string {
	return httpStatusToErrorCode(uint32(s))
}

func httpStatusToErrorCode(status uint32) string {
	switch status {
	case 100:
		return "Continue"
	case 101:
		return "SwitchingProtocols"
	case 102:
		return "Processing"
	case 103:
		return "EarlyHints"

	case 200:
		return "OK"
	case 201:
		return "Created"
	case 202:
		return "Accepted"
	case 203:
		return "NonAuthoritativeInfo"
	case 204:
		return "NoContent"
	case 205:
		return "ResetContent"
	case 206:
		return "PartialContent"
	case 207:
		return "MultiStatus"
	case 208:
		return "AlreadyReported"
	case 226:
		return "IMUsed"

	case 300:
		return "MultipleChoices"
	case 301:
		return "MovedPermanently"
	case 302:
		return "Found"
	case 303:
		return "SeeOther"
	case 304:
		return "NotModified"
	case 305:
		return "UseProxy"
	case 306:
		return "Unused: RFC 9110, 15.4.7"
	case 307:
		return "TemporaryRedirect"
	case 308:
		return "PermanentRedirect"

	case 400:
		return "BadRequest"
	case 401:
		return "Unauthorized"
	case 402:
		return "PaymentRequired"
	case 403:
		return "Forbidden"
	case 404:
		return "NotFound"
	case 405:
		return "MethodNotAllowed"
	case 406:
		return "NotAcceptable"
	case 407:
		return "ProxyAuthRequired"
	case 408:
		return "RequestTimeout"
	case 409:
		return "Conflict"
	case 410:
		return "Gone"
	case 411:
		return "LengthRequired"
	case 412:
		return "PreconditionFailed"
	case 413:
		return "RequestEntityTooLarge"
	case 414:
		return "RequestURITooLong"
	case 415:
		return "UnsupportedMediaType"
	case 416:
		return "RequestedRangeNotSatisfiable"
	case 417:
		return "ExpectationFailed"
	case 418:
		return "Teapot"
	case 421:
		return "MisdirectedRequest"
	case 422:
		return "UnprocessableEntity"
	case 423:
		return "Locked"
	case 424:
		return "FailedDependency"
	case 425:
		return "TooEarly"
	case 426:
		return "UpgradeRequired"
	case 428:
		return "PreconditionRequired"
	case 429:
		return "TooManyRequests"
	case 431:
		return "RequestHeaderFieldsTooLarge"
	case 451:
		return "UnavailableForLegalReasons"

	case 500:
		return "InternalServerError"
	case 501:
		return "NotImplemented"
	case 502:
		return "BadGateway"
	case 503:
		return "ServiceUnavailable"
	case 504:
		return "GatewayTimeout"
	case 505:
		return "HTTPVersionNotSupported"
	case 506:
		return "VariantAlsoNegotiates"
	case 507:
		return "InsufficientStorage"
	case 508:
		return "LoopDetected"
	case 510:
		return "NotExtended"
	case 511:
		return "NetworkAuthenticationRequired"

	default:
		return "UnknownStatusCode"
	}
}

func httpStatusToGrpcCode(status uint32) codes.Code {
	// https://github.com/grpc/grpc/blob/e5d6bda22ea78337c0b39df4844632fc967b61e5/doc/http-grpc-status-mapping.md
	// 400 Bad Request          INTERNAL
	// 401 Unauthorized         UNAUTHENTICATED
	// 403 Forbidden            PERMISSION_DENIED
	// 404 Not Found            UNIMPLEMENTED
	// 429 Too Many Requests    UNAVAILABLE
	// 502 Bad Gateway          UNAVAILABLE
	// 503 Service Unavailable  UNAVAILABLE
	// 504 Gateway Timeout      UNAVAILABLE
	// All other codes          UNKNOWN
	switch status {
	case 400:
		return codes.Internal
	case 401:
		return codes.Unauthenticated
	case 403:
		return codes.PermissionDenied
	case 404:
		return codes.Unimplemented
	case 429:
		return codes.Unavailable

	case 502:
		return codes.Unavailable
	case 503:
		return codes.Unavailable
	case 504:
		return codes.Unavailable

	default:
		return codes.Unknown
	}
}

func grpcCodeToHttpStatus(code codes.Code) uint32 {
	switch code {
	case 0:
		// Code: OK
		// Not an error; returned on success.
		return 200
	case 1:
		// Code: CANCELLED
		// The operation was cancelled, typically by the caller.
		break
	case 2:
		// Code: UNKNOWN
		// Unknown error. For example, this error may be returned when a Status
		// value received from another address space belongs to an error space
		// that is not known in this address space. Also errors raised by APIs
		// that do not return enough error information may be converted to this
		// error.
		return 500
	case 3:
		// Code: INVALID_ARGUMENT
		// The client specified an invalid argument. Note that this differs from
		// FAILED_PRECONDITION. INVALID_ARGUMENT indicates arguments that are
		// problematic regardless of the state of the system (e.g., a malformed
		// file name).
		return 400 // Bad Request
	case 4:
		// Code: DEADLINE_EXCEEDED
		// The deadline expired before the operation could complete. For operations
		// that change the state of the system, this error may be returned even
		// if the operation has completed successfully. For example, a successful
		// response from a server could have been delayed long enough for the
		// deadline to expire.
		return 408 // Request Timeout
	case 5:
		// Code: NOT_FOUND
		// Some requested entity (e.g., file or directory) was not found. Note
		// to server developers: if a request is denied for an entire class of
		// users, such as gradual feature rollout or undocumented allowlist,
		// NOT_FOUND may be used. If a request is denied for some users within
		// a class of users, such as user-based access control, PERMISSION_DENIED
		// must be used.
		return 404 // Not Found
	case 6:
		// Code: ALREADY_EXISTS
		// The entity that a client attempted to create (e.g., file or directory)
		// already exists.
		return 409 // Conflict
	case 7:
		// Code: PERMISSION_DENIED
		// The caller does not have permission to execute the specified operation.
		// PERMISSION_DENIED must not be used for rejections caused by exhausting some resource (use RESOURCE_EXHAUSTED instead for those errors). PERMISSION_DENIED must not be used if the caller can not be identified (use UNAUTHENTICATED instead for those errors). This error code does not imply the request is valid or the requested entity exists or satisfies other pre-conditions.
		return 403 // Forbidden
	case 8:
		// Code: RESOURCE_EXHAUSTED
		// Some resource has been exhausted, perhaps a per-user quota, or perhaps
		// the entire file system is out of space.
		break
	case 9:
		// Code: FAILED_PRECONDITION
		// The operation was rejected because the system is not in a state required
		// for the operation’s execution. For example, the directory to be deleted
		// is non-empty, an rmdir operation is applied to a non-directory, etc.
		// Service implementors can use the following guidelines to decide between
		// FAILED_PRECONDITION, ABORTED, and UNAVAILABLE: (a) Use UNAVAILABLE
		// if the client can retry just the failing call. (b) Use ABORTED if the
		// client should retry at a higher level (e.g., when a client-specified
		// test-and-set fails, indicating the client should restart a read-modify-write
		// sequence). (c) Use FAILED_PRECONDITION if the client should not retry
		// until the system state has been explicitly fixed. E.g., if an “rmdir”
		// fails because the directory is non-empty, FAILED_PRECONDITION should
		// be returned since the client should not retry unless the files are
		// deleted from the directory.
		return 412 // Precondition Failed
	case 10:
		// Code: ABORTED
		// The operation was aborted, typically due to a concurrency issue such
		// as a sequencer check failure or transaction abort. See the guidelines
		// above for deciding between FAILED_PRECONDITION, ABORTED, and UNAVAILABLE.
		break
	case 11:
		// Code: OUT_OF_RANGE
		// The operation was attempted past the valid range. E.g., seeking or
		// reading past end-of-file. Unlike INVALID_ARGUMENT, this error indicates
		// a problem that may be fixed if the system state changes. For example,
		// a 32-bit file system will generate INVALID_ARGUMENT if asked to read
		// at an offset that is not in the range [0,2^32-1], but it will generate
		// OUT_OF_RANGE if asked to read from an offset past the current file
		// size. There is a fair bit of overlap between FAILED_PRECONDITION and
		// OUT_OF_RANGE. We recommend using OUT_OF_RANGE (the more specific error)
		// when it applies so that callers who are iterating through a space can
		// easily look for an OUT_OF_RANGE error to detect when they are done.
		return 416 // Range Not Satisfiable
	case 12:
		// Code: UNIMPLEMENTED
		// The operation is not implemented or is not supported/enabled in this
		// service.
		return 501 // Not Implemented
	case 13:
		// Code: INTERNAL
		// Internal errors. This means that some invariants expected by the
		// underlying system have been broken. This error code is reserved for
		// serious errors.
		return 500
	case 14:
		// Code: UNAVAILABLE
		// The service is currently unavailable. This is most likely a transient
		// condition, which can be corrected by retrying with a backoff. Note
		// that it is not always safe to retry non-idempotent operations.
		return 503 // Service Unavailable
	case 15:
		// Code: DATA_LOSS
		// Unrecoverable data loss or corruption.
		break
	case 16:
		// Code: UNAUTHENTICATED
		// The request does not have valid authentication credentials for the
		// operation.
		return 401 // Unauthorized
	}
	return 500
}
