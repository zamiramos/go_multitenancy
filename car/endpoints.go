package car

import (
	"context"
	"github.com/go-kit/kit/endpoint"
	httptransport "github.com/go-kit/kit/transport/http"
	"net/url"
	"strings"
)

type Endpoints struct {
	PostCarEndpoint   endpoint.Endpoint
	GetCarEndpoint    endpoint.Endpoint
	DeleteCarEndpoint endpoint.Endpoint
}

// PostCar implements Service. Primarily useful in a client.
func (e Endpoints) PostCar(ctx context.Context, car Car) error {
	request := postCarRequest{Car: car}
	response, err := e.PostCarEndpoint(ctx, request)
	if err != nil {
		return err
	}
	resp := response.(postCarResponse)
	return resp.Err
}

// GetProfile implements Service. Primarily useful in a client.
func (e Endpoints) GetProfile(ctx context.Context, id string) (Car, error) {
	request := getCarRequest{ID: id}
	response, err := e.GetCarEndpoint(ctx, request)
	if err != nil {
		return Car{}, err
	}
	resp := response.(getCarResponse)
	return resp.Car, resp.Err
}

func MakeServerEndpoints(s Service) Endpoints {
	return Endpoints{
		PostCarEndpoint:   MakePostCarEndpoint(s),
		GetCarEndpoint:    MakeGetCarEndpoint(s),
		DeleteCarEndpoint: MakeDeleteCarEndpoint(s),
	}
}

// MakeClientEndpoints returns an Endpoints struct where each endpoint invokes
// the corresponding method on the remote instance, via a transport/http.Client.
func MakeClientEndpoints(instance string) (Endpoints, error) {
	if !strings.HasPrefix(instance, "http") {
		instance = "http://" + instance
	}
	tgt, err := url.Parse(instance)
	if err != nil {
		return Endpoints{}, err
	}
	tgt.Path = ""

	options := []httptransport.ClientOption{}

	// Note that the request encoders need to modify the request URL, changing
	// the path. That's fine: we simply need to provide specific encoders for
	// each endpoint.

	return Endpoints{
		PostCarEndpoint:   httptransport.NewClient("POST", tgt, encodePostCarRequest, decodePostCarResponse, options...).Endpoint(),
		GetCarEndpoint:    httptransport.NewClient("GET", tgt, encodeGetCarRequest, decodeGetCarResponse, options...).Endpoint(),
		DeleteCarEndpoint: httptransport.NewClient("DELETE", tgt, encodeDeleteCarRequest, decodeDeleteCarResponse, options...).Endpoint(),
	}, nil
}

// MakePostCarEndpoint returns an endpoint via the passed service.
// Primarily useful in a server.
func MakePostCarEndpoint(s Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(postCarRequest)
		e := s.PostCar(ctx, req.Car)
		return postCarResponse{Err: e}, nil
	}
}

// MakeGetCarEndpoint returns an endpoint via the passed service.
// Primarily useful in a server.
func MakeGetCarEndpoint(s Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(getCarRequest)
		car, e := s.GetCar(ctx, req.ID)
		return getCarResponse{Car: car, Err: e}, nil
	}
}

// MakeDeleteCarEndpoint returns an endpoint via the passed service.
// Primarily useful in a server.
func MakeDeleteCarEndpoint(s Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(deleteCarRequest)
		e := s.DeleteCar(ctx, req.ID)
		return deleteCarResponse{Err: e}, nil
	}
}

type postCarRequest struct {
	Car Car
}

type postCarResponse struct {
	Err error `json:"err,omitempty"`
}

func (r postCarResponse) error() error { return r.Err }

type getCarRequest struct {
	ID string
}

type getCarResponse struct {
	Car Car   `json:"profile,omitempty"`
	Err error `json:"err,omitempty"`
}

func (r getCarResponse) error() error { return r.Err }

type deleteCarRequest struct {
	ID string
}

type deleteCarResponse struct {
	Err error `json:"err,omitempty"`
}

func (r deleteCarResponse) error() error { return r.Err }
