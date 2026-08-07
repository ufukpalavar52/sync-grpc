package util

import (
	"context"

	"google.golang.org/grpc/metadata"
)

const TransactionIDMetaKey = "Transaction-Id"

func GetMetadata(ctx context.Context) metadata.MD {
	md, _ := metadata.FromIncomingContext(ctx)
	return md
}

func GetFirstMetaVal(md metadata.MD, key string) any {
	if md == nil {
		return ""
	}

	val := md.Get(key)
	if len(val) == 0 {
		return ""
	}

	return val[0]
}

func GetMetaTransactionId(md metadata.MD) any {
	return GetFirstMetaVal(md, TransactionIDMetaKey)
}
