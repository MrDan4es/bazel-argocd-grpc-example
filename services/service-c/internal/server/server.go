package server

import (
	"context"
	"os"

	envoyCore "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoyAuth "github.com/envoyproxy/go-control-plane/envoy/service/auth/v3"
	envoyType "github.com/envoyproxy/go-control-plane/envoy/type/v3"
	"github.com/rs/zerolog"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	HostnameHeader = "x-hostname-info"
	IPHeader       = "x-real-ip"
)

var (
	hostname string
)

func init() {
	hostname, _ = os.Hostname()
}

type Server struct {
	log zerolog.Logger
}

func New(ctx context.Context) *Server {
	log := zerolog.Ctx(ctx).
		With().
		Str("component", "server").
		Logger()

	return &Server{
		log: log,
	}
}

func (s *Server) Check(ctx context.Context, req *envoyAuth.CheckRequest) (*envoyAuth.CheckResponse, error) {
	path := req.Attributes.Request.Http.Path
	userIp := req.Attributes.Request.Http.Headers[IPHeader]

	s.log.Info().
		Str("method", "/envoy.service.auth.v3.Authorization/Check").
		Str("path", path).
		Str("user_ip", userIp).
		Send()

	return responseOk(hostname)
}

func responseOk(value string) (*envoyAuth.CheckResponse, error) {
	return &envoyAuth.CheckResponse{
		Status: status.New(codes.OK, "").Proto(),
		HttpResponse: &envoyAuth.CheckResponse_OkResponse{
			OkResponse: &envoyAuth.OkHttpResponse{
				Headers: []*envoyCore.HeaderValueOption{
					{
						Header: &envoyCore.HeaderValue{
							Key:   HostnameHeader,
							Value: value,
						},
						AppendAction: envoyCore.HeaderValueOption_APPEND_IF_EXISTS_OR_ADD,
					},
				},
			},
		},
	}, nil
}

func responseDenied(reason string) (*envoyAuth.CheckResponse, error) {
	return &envoyAuth.CheckResponse{
		Status: status.New(codes.Unauthenticated, reason).Proto(),
		HttpResponse: &envoyAuth.CheckResponse_DeniedResponse{
			DeniedResponse: &envoyAuth.DeniedHttpResponse{
				Status: &envoyType.HttpStatus{
					Code: envoyType.StatusCode_Unauthorized,
				},
				Headers: []*envoyCore.HeaderValueOption{},
			},
		},
	}, nil
}
