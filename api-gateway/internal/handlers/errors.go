package handlers

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// HandleGrpcError 将gRPC错误转换为HTTP响应
func HandleGrpcError(c *gin.Context, err error, defaultMessage string) {
	log.Printf("gRPC Error: %v", err)
	st, ok := status.FromError(err)
	if ok {
		// 根据gRPC状态码映射到HTTP状态码
		httpCode := http.StatusInternalServerError // 默认为500
		switch st.Code() {
		case codes.InvalidArgument:
			httpCode = http.StatusBadRequest
		case codes.Unauthenticated:
			httpCode = http.StatusUnauthorized
		case codes.PermissionDenied:
			httpCode = http.StatusForbidden
		case codes.NotFound:
			httpCode = http.StatusNotFound
		case codes.AlreadyExists:
			httpCode = http.StatusConflict
		case codes.DeadlineExceeded:
			httpCode = http.StatusGatewayTimeout
		}
		c.JSON(httpCode, gin.H{"error": st.Message()})
	} else {
		// 如果不是标准的gRPC错误
		c.JSON(http.StatusInternalServerError, gin.H{"error": defaultMessage, "details": err.Error()})
	}
}
