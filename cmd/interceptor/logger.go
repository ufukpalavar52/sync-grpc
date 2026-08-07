package interceptor

import (
	"context"
	"imapsync-user/internal/util"
	"imapsync-user/logger"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

func LoggingInterceptor(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	md, _ := metadata.FromIncomingContext(ctx)
	startTime := time.Now()
	res, err := handler(ctx, req)

	duration := time.Since(startTime)

	if err != nil {
		logger.ErrorL("gRpc request was failed with error.", md, CreateErrorLog(err, md, ctx, info.FullMethod, duration))
		return res, err
	}

	logger.InfoL("gRpc request was successful.", md, CreateSuccessLog(md, ctx, info.FullMethod, duration))

	return res, err
}

func StreamLoggerInterceptor(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	startTime := time.Now()
	md, _ := metadata.FromIncomingContext(ss.Context())
	err := handler(srv, ss)
	duration := time.Since(startTime)

	if err != nil {
		logger.ErrorL("gRpc stream request was failed with error.", md, CreateErrorLog(err, md, ss.Context(), info.FullMethod, duration))
		return err
	}

	logger.InfoL("gRpc stream request was successful.", md, CreateSuccessLog(md, ss.Context(), info.FullMethod, duration))
	return nil
}

func CreateErrorLog(err error, md metadata.MD, ctx context.Context, fullMethod string, duration time.Duration) logger.DT {
	clientIP := ""
	if p, ok := peer.FromContext(ctx); ok {
		clientIP = p.Addr.String()
	}

	transactionId := util.GetFirstMetaVal(md, util.TransactionHeaderKey)
	return logger.DT{
		"clientIp":      clientIP,
		"transactionId": transactionId,
		"duration":      duration.Seconds(),
		"code":          status.Code(err).String(),
		"method":        fullMethod,
		"error":         err.Error(),
	}
}

func CreateSuccessLog(md metadata.MD, ctx context.Context, fullMethod string, duration time.Duration) logger.DT {
	clientIP := ""
	if p, ok := peer.FromContext(ctx); ok {
		clientIP = p.Addr.String()
	}
	transactionId := util.GetFirstMetaVal(md, util.TransactionHeaderKey)

	return logger.DT{
		"clientIp":      clientIP,
		"transactionId": transactionId,
		"duration":      duration.Seconds(),
		"code":          "OK",
		"method":        fullMethod,
	}
}
