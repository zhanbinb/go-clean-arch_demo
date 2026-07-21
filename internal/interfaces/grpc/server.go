// Package grpc wires the gRPC server: service registration and reflection.
package grpc

import (
	"google.golang.org/grpc"

	"github.com/zhanbinb/go-clean-arch_demo/internal/application/article"
	articlegrpc "github.com/zhanbinb/go-clean-arch_demo/internal/interfaces/grpc/handler"
)

// Handlers bundles gRPC handlers for DI.
type Handlers struct {
	Article *articlegrpc.ArticleServer
}

// NewHandlers constructs gRPC handlers from services.
func NewHandlers(articleSvc *article.Service) *Handlers {
	return &Handlers{
		Article: articlegrpc.NewArticleServer(articleSvc),
	}
}

// Register attaches all gRPC service implementations to the server.
//
// Once `make proto` runs, replace this with the generated
//   articlev1.RegisterArticleServiceServer(s, h.Article)
func (h *Handlers) Register(s *grpc.Server) {
	s.RegisterService(&articlegrpc.ArticleService_ServiceDesc, h.Article)
}
