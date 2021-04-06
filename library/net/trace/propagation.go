package trace

import (
	"errors"
	"net/http"
)

var (
	// ErrUnsupportedFormat occurs when the `format` passed to Tracer.Inject() or
	// Tracer.Extract is not recognized by the Tracer implementation.
	ErrUnsupportedFormat = errors.New("trace: Unknown or unsupported Inject/Extract format")

	// ErrTraceNotFound occurs when the `carrier` passed to Tracer.Extract() is
	// valid and uncorrupted but has insufficient information to extract a Trace.
	ErrTraceNotFound = errors.New("trace: Trace not found in Extract carrier")

	// ErrInvalidTrace errors occur when Tracer.Inject() is asked to operate on a Trace
	// which it is not prepared to handle (eg:
	ErrInvalidTrace = errors.New("trace: Trace type incompatible with tracer")

	// ErrInvalidCarrier errors occur when Tracer.Inject() or Tracer.Extract()
	// implementations expect a different type of `carrier` than they are
	// given.
	ErrInvalidCarrier = errors.New("trace: Invalid Inject/Extract carrier")

	// ErrTraceCorrupted occurs when the `carrier` passed to
	// Tracer.Extract() is of the expected type but is corrupted.
	ErrTraceCorrupted = errors.New("trace: Trace data corrupted in Extract carrier")
)

type BuiltinFormat byte

// support format list
const (
	// HTTPFormat represent Trace as HTTP header string pairs.
	// the HTTPFormat format requires that the keys and values be
	// valid as HTTP headers as-is (eg character casing may be
	// unstable and special characters are disallowed in keys,
	// values should be URL-escaped, etc).
	// the carrier must be a `http.Header`.
	HTTPFormat BuiltinFormat = iota

	// GRPCFormat represents Trace as gRPC metadata.
	// the carrier must be a `google.golang.org/grpc/metadata.MD`.
	GRPCFormat
)

type Carrier interface {
	Set(key, val string)
	Get(key string) string
}

// propagator is responsible for injecting and extracting `Trace`
// instances from a format-specific "carrier"
type propagator interface {
	Inject(carrier interface{}) (Carrier, error)
	Extract(carrier interface{}) (Carrier, error)
}

type httpPropagator struct{}

type httpCarrier http.Header

func (h httpCarrier) Set(key, val string) {
	http.Header(h).Set(key, val)
}

func (h httpCarrier) Get(key string) string {
	return http.Header(h).Get(key)
}

func (httpPropagator) Inject(carrier interface{}) (Carrier, error) {
	header, ok := carrier.(http.Header)
	if !ok {
		return nil, ErrInvalidCarrier
	}
	if header == nil {
		return nil, ErrInvalidTrace
	}
	return httpCarrier(header), nil
}

func (httpPropagator) Extract(carrier interface{}) (Carrier, error) {
	header, ok := carrier.(http.Header)
	if !ok {
		return nil, ErrInvalidCarrier
	}
	if header == nil {
		return nil, ErrTraceNotFound
	}
	return httpCarrier(header), nil
}
