package service

import (
	"fmt"
	"imapsync-grpc/config"
	"imapsync-grpc/internal/util"
	"imapsync-grpc/logger"
	"imapsync-grpc/pkg/pb/sync_log"
	"io"
	"path/filepath"
	"strings"
	"time"

	"github.com/nxadm/tail"
	"google.golang.org/grpc"
)

const batchSize = 100
const idleSeconds = 500
const failureSeconds = 50

type SyncLogService struct{}

func NewSyncLogService() *SyncLogService {
	return &SyncLogService{}
}

func (s *SyncLogService) Tail(req *sync_log.SyncLogRequest, stream grpc.ServerStreamingServer[sync_log.SyncLogChunk]) error {
	logPath := fmt.Sprintf("%s/%s.log", config.SyncLogPath, req.GetTransactionId())
	t, err := tail.TailFile(logPath, tail.Config{
		Follow:    true,
		ReOpen:    true,
		MustExist: true,
		Location:  &tail.SeekInfo{Offset: 0, Whence: io.SeekStart},
	})

	isFinish := make(chan int, 1)

	if err != nil {
		return err
	}
	defer t.Cleanup()

	var buffer string
	var lineCount int
	transactionId := req.GetTransactionId()

	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()

	idleDuration := time.Duration(idleSeconds) * time.Second
	if s.isProcessComplete(req.GetStatus()) {
		idleDuration = 10 * time.Second
	}

	idleTimer := time.NewTimer(idleDuration)
	defer idleTimer.Stop()

	for {
		select {
		case <-idleTimer.C:
			fPath := filepath.Base(logPath)
			logMsg := fmt.Sprintf("Connection was closed because data didn't come from %s for %d seconds.", fPath, int(idleDuration.Seconds()))
			logger.InfoL(logMsg, util.GetMetadata(stream.Context()))
			_ = t.Stop()

			if lineCount > 0 {
				logger.DebugL("Added chunk data", util.GetMetadata(stream.Context()), logger.DT{"transactionId": transactionId, "data": buffer})
				_ = stream.Send(&sync_log.SyncLogChunk{Content: buffer})
			}

			return nil

		case <-stream.Context().Done():
			_ = t.Stop()
			return nil

		case line, ok := <-t.Lines:
			if !ok || line.Err != nil {
				continue
			}

			idleTimer.Reset(idleDuration)
			buffer += line.Text + "\n"
			lineCount++

			if lineCount >= batchSize {
				if err = s.SendData(buffer, t, stream); err != nil {
					return err
				}

				checkFinish := s.checkFinish(buffer)
				if checkFinish {
					go func() {
						time.Sleep(time.Second * 5)
						isFinish <- 1
					}()
				}

				if !checkFinish && !s.isProcessComplete(req.GetStatus()) && s.checkFailure(buffer) {
					go func() {
						time.Sleep(time.Second * failureSeconds)
						isFinish <- 1
					}()
				}

				logger.DebugL("Added chuck data", util.GetMetadata(stream.Context()), logger.DT{"transactionId": transactionId, "data": buffer})

				buffer = ""
				lineCount = 0
			}

		case <-ticker.C:
			if lineCount > 0 {
				if err = s.SendData(buffer, t, stream); err != nil {
					return err
				}

				checkFinish := s.checkFinish(buffer)
				if checkFinish {
					go func() {
						time.Sleep(time.Second * 5)
						isFinish <- 1
					}()
				}
				if !checkFinish && !s.isProcessComplete(req.GetStatus()) && s.checkFailure(buffer) {
					go func() {
						time.Sleep(time.Second * failureSeconds)
						isFinish <- 1
					}()
				}
				logger.DebugL("Added chuck data", util.GetMetadata(stream.Context()), logger.DT{"transactionId": transactionId, "data": buffer})

				buffer = ""
				lineCount = 0
			}
		case <-isFinish:
			logger.DebugL("Finishing stream flow..", util.GetMetadata(stream.Context()), logger.DT{"transactionId": transactionId, "data": buffer})
			if lineCount > 0 {
				if err = stream.Send(&sync_log.SyncLogChunk{Content: buffer}); err != nil {
					_ = t.Stop()
					return err
				}
				_ = t.Stop()
				return nil
			}
			_ = t.Stop()
			return nil
		}
	}
}

func (s *SyncLogService) SendData(buffer string, t *tail.Tail, stream grpc.ServerStreamingServer[sync_log.SyncLogChunk]) error {
	if err := stream.Send(&sync_log.SyncLogChunk{Content: buffer}); err != nil {
		_ = t.Stop()
		return err
	}
	return nil
}

func (s *SyncLogService) checkFinish(str string) bool {
	return strings.Contains(str, "EX_OK")
}

func (s *SyncLogService) checkFailure(str string) bool {
	return strings.Contains(str, "EXIT_") && strings.Contains(str, "_FAILURE_")
}

func (s *SyncLogService) isProcessComplete(status string) bool {
	return status != util.ImapInProgress && status != util.ImapPending
}
