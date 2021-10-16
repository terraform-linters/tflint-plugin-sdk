package host2plugin

import (
	"context"
	"fmt"

	"github.com/terraform-linters/tflint-plugin-sdk/logger"
	"google.golang.org/grpc"
)

type serviceType int32

const (
	hostServiceType serviceType = iota
	pluginServiceType
)

func loggingInterceptor(service serviceType) grpc.UnaryServerInterceptor {
	var direction string
	switch service {
	case hostServiceType:
		direction = "plugin2host"
	case pluginServiceType:
		direction = "host2plugin"
	default:
		panic("never happened")
	}

	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		logger.Debug(fmt.Sprintf("gRPC request (%s)", direction), "method", info.FullMethod, "req", req)
		ret, err := handler(ctx, req)
		if err != nil {
			logger.Error(fmt.Sprintf("failed to gRPC request (%s)", direction), "method", info.FullMethod, "err", err)
		}
		return ret, err
	}
}
