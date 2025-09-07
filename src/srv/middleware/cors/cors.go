package cors

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-secure-stdlib/strutil"

	"github.com/kheina/openconman/src/errors"
)

type KhCorsMiddleware struct {
	allowedOrigins   []string
	allowedProtocols []string
	allowedHeaders   []string
	allowedMethods   []string
	exposedHeaders   []string
	allowCredentials bool
	maxAge           int

	logger hclog.Logger
}

func New(logger hclog.Logger) *KhCorsMiddleware {
	return &KhCorsMiddleware{
		allowedOrigins: []string{
			"localhost",
			"127.0.0.1",
		},
		allowedProtocols: []string{
			"http",
			"https",
		},
		allowedHeaders: append([]string{
			"access-control-request-method",
			"origin",
		}, []string{
			"accept",
			"accept-language",
			"authorization",
			"cache-control",
			"content-encoding",
			"content-language",
			"content-length",
			"content-security-policy",
			"content-type",
			"cookie",
			"host",
			"location",
			"referer",
			"referrer-policy",
			"set-cookie",
			"user-agent",
			"www-authenticate",
			"kh-trace",
			"x-frame-options",
			"x-xss-protection",
		}...),
		allowedMethods: []string{
			http.MethodGet,
			http.MethodPut,
			http.MethodPost,
			http.MethodPatch,
			http.MethodDelete,
		},
		exposedHeaders: []string{
			"authorization",
			"cache-control",
			"content-type",
			"cookie",
			"set-cookie",
			"www-authenticate",
			"kh-trace",
			"kh-hash",
		},
		allowCredentials: true,
		maxAge:           86400,

		logger: logger,
	}
}

func (c *KhCorsMiddleware) WrapHandler(h http.Handler) http.Handler {
	const op = "cors.(KhCorsMiddleware).WrapHandler"
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if o := req.Header.Get("origin"); o != "" {
			c.logger.Trace(fmt.Sprintf("%s: origin: %s", op, o))
			origin, err := url.Parse(o)

			if err != nil {
				w.Header().Set("content-type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				err := errors.Wrap(op, err, "failed to parse origin url")
				c.logger.Debug(fmt.Sprintf("%s: cors error: %s", op, err))
				enc := json.NewEncoder(w)
				if err = enc.Encode(err); err != nil {
					c.logger.Error(fmt.Sprintf("%s: failed to send api error: %s", op, err))
				}
				return
			}

			switch {
			case strutil.StrListContains(c.allowedOrigins, "*"):
			case strutil.StrListContains(c.allowedProtocols, origin.Scheme) &&
				strutil.StrListContains(c.allowedOrigins, origin.Hostname()):
			default:
				w.Header().Set("content-type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				err := errors.New(http.StatusBadRequest, op, "origin not allowed", errors.WithErrorCode("OriginNotAllowed"))
				c.logger.Debug(fmt.Sprintf("%s: cors error: %s", op, err))
				enc := json.NewEncoder(w)
				if err = enc.Encode(err); err != nil {
					c.logger.Error(fmt.Sprintf("%s: failed to send api error: %s", op, err))
				}
				c.logger.Trace(fmt.Sprintf("%s: cors request failed: origin not allowed", op))
				return
			}

			// this is the only logical difference between this implementation and the python version
			// of KhCorsMiddleware, this is technically more correct
			if req.Method == http.MethodOptions && !strutil.StrListContains(c.allowedMethods, req.Header.Get("access-control-request-method")) {
				w.Header().Set("content-type", "application/json")
				w.WriteHeader(http.StatusMethodNotAllowed)
				err := errors.New(http.StatusMethodNotAllowed, op, "request method not allowed")
				c.logger.Debug(fmt.Sprintf("%s: cors error: %s", op, err))
				enc := json.NewEncoder(w)
				if err = enc.Encode(err); err != nil {
					c.logger.Error(fmt.Sprintf("%s: failed to send api error: %s", op, err))
				}
				c.logger.Trace(fmt.Sprintf("%s: cors request failed: method not allowed", op))
				return
			}

			w.Header().Set("access-control-allow-origin", origin.String())
			w.Header().Set("access-control-allow-methods", strings.Join(c.allowedMethods, ","))
			w.Header().Set("access-control-allow-headers", strings.Join(c.allowedHeaders, ","))
			w.Header().Set("access-control-allow-credentials", strconv.FormatBool(c.allowCredentials))
			w.Header().Set("access-control-max-age", strconv.FormatUint(uint64(c.maxAge), 10))
			w.Header().Set("access-control-expose-headers", strings.Join(c.exposedHeaders, ","))

			if req.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}
		} else {
			c.logger.Trace(fmt.Sprintf("%s: no origin, skipping cors", op))
		}

		h.ServeHTTP(w, req)
	})
}
