package errors

import (
	"context"
	"encoding/hex"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/hashicorp/go-hclog"
	"google.golang.org/grpc/status"

	pb "github.com/kheina/openconman/src/gen/pbs/api/error"
)

func refid() string {
	guid := uuid.New()
	dst := make([]byte, 32)
	_ = hex.Encode(dst, guid[:])
	return string(dst)
}

func ApiErrorHandler(l hclog.Logger) runtime.ErrorHandlerFunc {
	return func(ctx context.Context, mux *runtime.ServeMux, m runtime.Marshaler, w http.ResponseWriter, req *http.Request, err error) {
		var kerr *ApiError
		if sterr, ok := err.(interface{ GRPCStatus() *status.Status }); ok {
			for _, d := range sterr.GRPCStatus().Details() {
				fmt.Printf("==> sterr.Detail: %#v\n", d)
				if und, ok := d.(*pb.ApiError); ok {
					kerr = &ApiError{
						ApiError: und,
					}
					break
				}
			}
			if kerr == nil {
				code := sterr.GRPCStatus().Code()
				kerr = &ApiError{
					ApiError: &pb.ApiError{
						Status:  grpcCodeToHttpStatus(code),
						Code:    code.String(),
						Message: err.Error(),
						Refid:   refid(),
					},
				}
			}
		} else {
			kerr = &ApiError{
				ApiError: &pb.ApiError{
					Status:  500,
					Code:    httpStatusToErrorCode(500),
					Message: err.Error(),
					Refid:   refid(),
				},
			}
		}
		l.Error(err.Error(), "status", kerr.Status, "code", kerr.Code, "refid", kerr.Refid)
		kerr.WriteHeader(w)
		b, err := m.Marshal(kerr)
		if err != nil {
			w.Write([]byte("something very very wrong has happpened, and an error failed to be written"))
			return
		}
		w.Write(b)
	}
}
