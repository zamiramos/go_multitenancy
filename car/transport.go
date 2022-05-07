package car

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/transport"
	httptransport "github.com/go-kit/kit/transport/http"
	"github.com/gorilla/mux"
	"io/ioutil"
	"net/http"
	"net/url"
)

var (
	// ErrBadRouting is returned when an expected path variable is missing.
	// It always indicates programmer error.
	ErrBadRouting = errors.New("inconsistent mapping between route and handler (programmer error)")
)

// MakeHTTPHandler mounts all of the service endpoints into an http.Handler.
func MakeHTTPHandler(s Service, logger log.Logger, mwf ...mux.MiddlewareFunc) http.Handler {
	r := mux.NewRouter()
	r.Use(mwf...)
	e := MakeServerEndpoints(s)
	options := []httptransport.ServerOption{
		httptransport.ServerErrorHandler(transport.NewLogErrorHandler(logger)),
		httptransport.ServerErrorEncoder(encodeError),
	}

	// POST    /cars/                          adds another car
	// GET     /cars/:id                       retrieves the given car by id
	// DELETE  /cars/:id                       remove the given car

	r.Methods("POST").Path("/cars/").Handler(httptransport.NewServer(
		e.PostCarEndpoint,
		decodePostCarRequest,
		encodeResponse,
		options...,
	))
	r.Methods("GET").Path("/cars/{id}").Handler(httptransport.NewServer(
		e.GetCarEndpoint,
		decodeGetCarRequest,
		encodeResponse,
		options...,
	))
	r.Methods("DELETE").Path("/cars/{id}").Handler(httptransport.NewServer(
		e.DeleteCarEndpoint,
		decodeDeleteCarRequest,
		encodeResponse,
		options...,
	))
	return r
}

func decodePostCarRequest(_ context.Context, r *http.Request) (request interface{}, err error) {
	var req postCarRequest
	if e := json.NewDecoder(r.Body).Decode(&req.Car); e != nil {
		return nil, e
	}
	return req, nil
}

func decodeGetCarRequest(_ context.Context, r *http.Request) (request interface{}, err error) {
	vars := mux.Vars(r)
	id, ok := vars["id"]
	if !ok {
		return nil, ErrBadRouting
	}
	return getCarRequest{ID: id}, nil
}

func decodeDeleteCarRequest(_ context.Context, r *http.Request) (request interface{}, err error) {
	vars := mux.Vars(r)
	id, ok := vars["id"]
	if !ok {
		return nil, ErrBadRouting
	}
	return deleteCarRequest{ID: id}, nil
}

func decodePostCarResponse(_ context.Context, resp *http.Response) (interface{}, error) {
	var response postCarResponse
	err := json.NewDecoder(resp.Body).Decode(&response)
	return response, err
}

func decodeGetCarResponse(_ context.Context, resp *http.Response) (interface{}, error) {
	var response getCarResponse
	err := json.NewDecoder(resp.Body).Decode(&response)
	return response, err
}

func decodeDeleteCarResponse(_ context.Context, resp *http.Response) (interface{}, error) {
	var response deleteCarResponse
	err := json.NewDecoder(resp.Body).Decode(&response)
	return response, err
}

func encodePostCarRequest(ctx context.Context, req *http.Request, request interface{}) error {
	// r.Methods("POST").Path("/cars/")
	req.URL.Path = "/cars/"
	return encodeRequest(ctx, req, request)
}

func encodeGetCarRequest(ctx context.Context, req *http.Request, request interface{}) error {
	// r.Methods("GET").Path("/cars/{id}")
	r := request.(getCarRequest)
	carID := url.QueryEscape(r.ID)
	req.URL.Path = "/cars/" + carID
	return encodeRequest(ctx, req, request)
}

func encodeDeleteCarRequest(ctx context.Context, req *http.Request, request interface{}) error {
	// r.Methods("DELETE").Path("/cars/{id}")
	r := request.(deleteCarRequest)
	carID := url.QueryEscape(r.ID)
	req.URL.Path = "/cars/" + carID
	return encodeRequest(ctx, req, request)
}

// encodeResponse is the common method to encode all response types to the
// client. I chose to do it this way because, since we're using JSON, there's no
// reason to provide anything more specific. It's certainly possible to
// specialize on a per-response (per-method) basis.
func encodeResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	if e, ok := response.(errorer); ok && e.error() != nil {
		// Not a Go kit transport error, but a business-logic error.
		// Provide those as HTTP errors.
		encodeError(ctx, e.error(), w)
		return nil
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	return json.NewEncoder(w).Encode(response)
}

// errorer is implemented by all concrete response types that may contain
// errors. It allows us to change the HTTP response code without needing to
// trigger an endpoint (transport-level) error. For more information, read the
// big comment in endpoints.go.
type errorer interface {
	error() error
}

// encodeRequest likewise JSON-encodes the request to the HTTP request body.
// Don't use it directly as a transport/http.Client EncodeRequestFunc:
// go_multitenancy endpoints require mutating the HTTP method and request path.
func encodeRequest(_ context.Context, req *http.Request, request interface{}) error {
	var buf bytes.Buffer
	err := json.NewEncoder(&buf).Encode(request)
	if err != nil {
		return err
	}
	req.Body = ioutil.NopCloser(&buf)
	return nil
}

func encodeError(_ context.Context, err error, w http.ResponseWriter) {
	if err == nil {
		panic("encodeError with nil error")
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(codeFrom(err))
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": err.Error(),
	})
}

func codeFrom(err error) int {
	switch err {
	case ErrNotFound:
		return http.StatusNotFound
	case ErrAlreadyExists, ErrInconsistentIDs:
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}
