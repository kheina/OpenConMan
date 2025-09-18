package errors

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strings"

	"github.com/google/uuid"
	"github.com/hashicorp/go-hclog"
	"google.golang.org/grpc/status"

	pb "github.com/kheina/openconman/src/gen/pbs/api/error"
)

type ApiError struct {
	*pb.ApiError
}

func (e *ApiError) Error() string {
	return e.Message
}

// WriteHeader writes the necessary headers to the http.ResponseWriter given the current error
func (e *ApiError) WriteHeader(w http.ResponseWriter) {
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(int(e.Status))
}

func (e *ApiError) GRPCStatus() *status.Status {
	st := status.New(httpStatusToGrpcCode(e.Status), e.Error())
	st, _ = st.WithDetails(e.ApiError)
	return st
}

// New returns a new *ApiError in the stdlib error interface
//
// Supported Options: WithErrorCode
func New(status Status, op, msg string, opt ...Option) error {
	opts := getOpts(opt...)

	if opts.withErrorCode == "" {
		opts.withErrorCode = status.String()
	}

	guid := uuid.New()
	dst := make([]byte, 32)
	_ = hex.Encode(dst, guid[:])

	return &ApiError{
		&pb.ApiError{
			Status:  uint32(status),
			Code:    opts.withErrorCode,
			Message: fmt.Sprintf("%s: %s", op, msg),
			Refid:   string(dst),
		},
	}
}

type WrappedError struct {
	underlying error
	msg        string
}

func (e *WrappedError) Error() string {
	return e.msg
}

func (e *WrappedError) Unwrap() error {
	return e.underlying
}

// Wrap takes an existing error and wraps it with additional information.
//
// Supported Options: WithStatusCode, WithErrorCode
func Wrap(op string, err error, msg string, opt ...Option) error {
	apierr := &ApiError{}
	msg = fmt.Errorf("%s: %s: %w", op, msg, err).Error()
	if !errors.As(err, &apierr) {
		guid := uuid.New()
		dst := make([]byte, 32)
		_ = hex.Encode(dst, guid[:])

		opts := getOpts(opt...)
		if opts.withErrorCode == "" {
			opts.withErrorCode = opts.withStatusCode.String()
		}

		return &WrappedError{
			msg: msg,
			underlying: &ApiError{
				&pb.ApiError{
					Status:  uint32(opts.withStatusCode),
					Code:    opts.withErrorCode,
					Message: err.Error(),
					Refid:   string(dst),
				},
			},
		}
	}
	return &WrappedError{
		msg:        msg,
		underlying: err,
	}
}

var logger hclog.Logger = hclog.NewNullLogger()

func SetLogger(l hclog.Logger) {
	logger = l
}

func (e *WrappedError) GRPCStatus() *status.Status {
	var kerr *ApiError
	// GRPCStatus is called after the api handler has returned and the err is
	// getting encoded over the wire for grpc gateway. therefore, we log the
	// error in its entirety here so that none of the extra information is lost
	// before transferring it
	logger.Error(Tree(e))
	if !errors.As(e, &kerr) {
		return nil
	}
	return kerr.GRPCStatus()
}

type errv struct {
	key   string
	value any
}

func dumpErr(w io.Writer, i string, err any, fin bool) {
	v := reflect.ValueOf(err)
	for {
		switch v.Kind() {
		case reflect.Struct:
			m := v.NumField()
			errvs := []errv{}
			structs := []errv{}
			mk := 0
		fields:
			for n := range m {
				f := v.Field(n)
				k := v.Type().Field(n).Name
				switch f.Kind() {
				case reflect.Interface:
					fallthrough
				case reflect.Pointer:
					fallthrough
				case reflect.Struct:
					if !f.CanInterface() {
						continue fields
					}
					structs = append(structs, errv{
						key:   k,
						value: f.Interface(),
					})
					break fields
				}
				if !f.CanInterface() {
					break
				}
				mk = max(mk, len(k))
				errvs = append(errvs, errv{
					key:   k,
					value: f.Interface(),
				})
			}
			for n, s := range structs {
				fmt.Fprintf(w, "\n%s%s%v:", i, "├ ", s.key)
				dumpErr(w, i+"│ ", s.value, n+1 == len(structs))
			}
			for n, value := range errvs {
				j := i + "├ "
				if n+1 == len(errvs) {
					if fin && len(i) >= 4 {
						i = i[:len(i)-4] + "└ "
					}
					j = i + "└ "
				}
				fmt.Fprintf(w, "\n%s%s:%s %v", j, value.key, strings.Repeat(" ", mk-len(value.key)), value.value)
			}
			return
		case reflect.Pointer:
			v = v.Elem()
		default:
			fmt.Printf("unknown kind: %v\n", v.Kind())
			return
		}
	}
}

func tree(w io.Writer, i string, err error, fin bool) {
	// fmt.Printf("tree: %s\n", err)
	for err != nil {
		t := "├ "
		switch {
		case i == "":
			t = ""
		case fin:
			t = "└ "
		}

		switch x := err.(type) {
		case interface{ Unwrap() error }:
			e := x.Unwrap()
			if e != nil {
				fmt.Fprintf(w, "\n%s%s%s", i, t, strings.TrimSuffix(err.Error(), fmt.Sprintf(": %s", e)))
			} else {
				fmt.Fprintf(w, "\n%s%s%s", i, t, err)
			}
			dumpErr(w, i, err, false)
			err = e
		case interface{ Unwrap() []error }:
			dumpErr(w, i, err, false)
			fmt.Fprintf(w, "\n%s%s%s", i, t, err)
			errs := x.Unwrap()
			for n, e := range errs {
				tree(w, i+"  ", e, n+1 == len(errs))
			}
			err = nil
		default:
			fmt.Fprintf(w, "\n%s%s%s", i, t, err)
			j := i + "│ "
			if fin {
				j = i + "  "
			}
			dumpErr(w, j, err, false)
			err = nil
		}
		i += "  "
	}
}

func Tree(err error) string {
	var buf bytes.Buffer
	tree(&buf, "", err, true)
	return strings.Trim(buf.String(), "\n")
}
