package main

import (
	"fmt"
	"imapsync-grpc/cmd/handler"
	"imapsync-grpc/cmd/interceptor"
	"imapsync-grpc/config"
	"imapsync-grpc/internal/global"
	"imapsync-grpc/internal/service"
	"imapsync-grpc/logger"
	"imapsync-grpc/pkg/pb/analysis"
	"imapsync-grpc/pkg/pb/encrypt"
	"imapsync-grpc/pkg/pb/sync"
	"imapsync-grpc/pkg/pb/sync_log"
	"imapsync-grpc/pkg/pb/user"
	"log"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"gorm.io/gorm"
)

var db *gorm.DB

func init() {
	db = config.GetDB()
	log.Println("Initializing Encryption Key...")
	encService := service.NewEncryptionService(db)
	key, err := encService.GetOrCreateFirstKey(config.EncryptionKeyVersion)
	if err != nil {
		log.Fatalf("Failed to get or create encryption key: %v", err)
	}
	global.EncKey = key
}

func main() {
	logger.InitLog(config.LogPath, config.LogLevel)
	defer logger.Sync()
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", config.AppPort))
	if err != nil {
		logger.FatalF("Port didn't listening: %v", err)
	}

	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(interceptor.LoggingInterceptor),
		grpc.StreamInterceptor(interceptor.StreamLoggerInterceptor),
	)

	sync.RegisterSyncServiceServer(grpcServer, handler.NewSyncServer(db))
	user.RegisterUserServiceServer(grpcServer, handler.NewUserServer(db))
	encrypt.RegisterEncryptServiceServer(grpcServer, handler.NewEncryptServer(db))
	sync_log.RegisterSyncLogServiceServer(grpcServer, handler.NewSyncLogService())
	analysis.RegisterAnalysisServiceServer(grpcServer, handler.NewAnalysisServer(db))

	reflection.Register(grpcServer)

	logger.InfoF("gRPC server listening at %v", lis.Addr())

	if err = grpcServer.Serve(lis); err != nil {
		logger.FatalF("gRPC server failed to serve: %v", err)
	}
}
