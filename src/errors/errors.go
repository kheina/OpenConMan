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

// Wrap takes an existing error and wraps it with additional information. any additional args are passed directly to fmt.Errorf
func Wrap(op string, err error, msg string, arg ...any) error {
	apierr := &ApiError{}
	if !errors.As(err, &apierr) {
		guid := uuid.New()
		dst := make([]byte, 32)
		_ = hex.Encode(dst, guid[:])

		return &WrappedError{
			msg: fmt.Errorf("%s: %w: %w", op, fmt.Errorf(msg, arg...), err).Error(),
			underlying: &ApiError{
				&pb.ApiError{
					Status:  500,
					Code:    httpStatusToErrorCode(500),
					Message: err.Error(),
					Refid:   string(dst),
				},
			},
		}
	}
	return &WrappedError{
		msg:        fmt.Errorf("%s: %s: %w", op, msg, err).Error(),
		underlying: err,
	}
}

func (e *WrappedError) GRPCStatus() *status.Status {
	var kerr *ApiError
	if !errors.As(e, &kerr) {
		return nil
	}
	return kerr.GRPCStatus()
}

func dumpErr(w io.Writer, i string, err any) {
	v := reflect.ValueOf(err)
	for {
		switch v.Kind() {
		case reflect.Struct:
		fields:
			for n := range v.NumField() {
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
					fmt.Fprintf(w, "\n%s%s%v:", i, "├ ", k)
					dumpErr(w, i+"│ ", f.Interface())
					return
				}
				if !f.CanInterface() {
					return
				}
				fmt.Fprintf(w, "\n%s%s%v: %v", i, "├ ", k, f.Interface())
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

func Tree(err error) string {
	var buf bytes.Buffer
	i := ""
	for err != nil {
		switch x := err.(type) {
		case interface{ Unwrap() error }:
			e := x.Unwrap()
			// json.Marshal()
			dumpErr(&buf, i, err)
			t := "└ "
			if i == "" {
				t = ""
			}
			if e != nil {
				fmt.Fprintf(&buf, "\n%s%s%s", i, t, strings.TrimSuffix(err.Error(), fmt.Sprintf(": %s", e)))
			} else {
				fmt.Fprintf(&buf, "\n%s%s%s", i, t, err)
			}
			err = e
		// case interface{ Unwrap() []error }:
		// 	for _, err = range x.Unwrap() {
		// 		tree(node, err)
		// 	}
		// 	err = nil
		default:
			dumpErr(&buf, i, err)
			t := "└ "
			if i == "" {
				t = ""
			}
			fmt.Fprintf(&buf, "\n%s%s%s", i, t, err)
			err = nil
		}
		i += "  "
	}
	return strings.Trim(buf.String(), "\n")
}

// func Tree(err error) string {
// 	root := textree.NewNode(err.Error())
// 	tree(root, err)
// 	var buf bytes.Buffer
// 	opts := textree.NewRenderOptions()
// 	opts.ChildrenMarginBottom = 0
// 	opts.ChildrenMarginTop = 0
// 	root.Render(&buf, opts)
// 	return strings.Trim(buf.String(), "\n\r")
// }

// func tree(node *textree.Node, err error) {
// 	for err != nil {
// 		switch x := err.(type) {
// 		case interface{ Unwrap() error }:
// 			ie := x.Unwrap()
// 			if ie != nil {
// 				newnode := textree.NewNode(strings.TrimSuffix(err.Error(), fmt.Sprintf(": %s", ie)))
// 				node.Append(newnode)
// 				node = newnode
// 			}
// 			err = ie
// 		case interface{ Unwrap() []error }:
// 			for _, err = range x.Unwrap() {
// 				tree(node, err)
// 			}
// 			err = nil
// 		default:
// 			node.Append(textree.NewNode(err.Error()))
// 			err = nil
// 		}
// 	}
// }
