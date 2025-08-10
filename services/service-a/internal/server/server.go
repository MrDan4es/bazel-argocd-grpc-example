package server

import (
	"context"
	"fmt"
	"os"
	"runtime"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	pb "github.com/mrdan4es/bazel-argocd-grpc-example/services/service-a/api/v1"
)

type Server struct {
	pb.UnimplementedServiceAServer
}

func New() *Server {
	return &Server{}
}

func (s *Server) GetSystemInfo(ctx context.Context, _ *pb.GetSystemInfoRequest) (*pb.GetSystemInfoResponse, error) {
	r := &pb.GetSystemInfoResponse{
		Os:           runtime.GOOS,
		Arch:         runtime.GOARCH,
		GoVersion:    runtime.Version(),
		CpuCores:     int32(runtime.NumCPU()),
		K8SPod:       getK8sValue("POD_NAME", "pod"),
		K8SNode:      getK8sValue("NODE_NAME", "node"),
		K8SNamespace: getK8sValue("POD_NAMESPACE", "namespace"),
		K8SPodIp:     getK8sValue("POD_IP", "IP"),
	}

	if hostname, err := os.Hostname(); err == nil {
		r.Hostname = hostname
	}

	// CPU and memory info (Linux specific)
	if runtime.GOOS == "linux" {
		r.CpuInfo = readFileIfExists("/proc/cpuinfo")
		r.MemInfo = readFileIfExists("/proc/meminfo")
	}

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Internal, "get metadata from incoming context")
	}

	r.AuthzHostname = "undefined"
	authzHostNames := md["x-hostname-info"]
	if len(authzHostNames) >= 1 {
		r.AuthzHostname = authzHostNames[0]
	}

	return r, nil
}

func getK8sValue(envVar, description string) string {
	if val := os.Getenv(envVar); val != "" {
		return val
	}
	return "Not available (not in k8s or " + description + " not set)"
}

func readFileIfExists(path string) string {
	if _, err := os.Stat(path); err == nil {
		if content, err := os.ReadFile(path); err == nil {
			return fmt.Sprintf("\nContents of %s:\n%s\n", path, string(content))
		}
	}
	return ""
}
