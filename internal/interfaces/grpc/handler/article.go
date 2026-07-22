// Package handler contains the gRPC service implementations.
//
// In v2 we hand-write the minimal service interface that mimics what
// protoc-gen-go-grpc would generate. The actual generated code lives in
// api/gen/go/ and is produced by \`make proto\`. Re-generating replaces this
// file's manually-rolled types with the generated ones.
package handler

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/zhanbinb/go-clean-arch_demo/internal/application/article"
	"github.com/zhanbinb/go-clean-arch_demo/pkg/errcode"
)

// ----------------------------------------------------------------------
// Manually-rolled types (mimic protoc-gen-go output).
// In production these live in api/gen/go/article/v1/article.pb.go.
// ----------------------------------------------------------------------

// Article mirrors api/proto/article/v1/article.proto.
type Article struct {
	Id         int64
	Title      string
	Content    string
	AuthorId   int64
	AuthorName string
	CreatedAt  int64 // unix seconds
	UpdatedAt  int64 // unix seconds
}

type GetArticleRequest struct {
	Id int64
}
type GetArticleResponse struct {
	Article *Article
}

type CreateArticleRequest struct {
	Title    string
	Content  string
	AuthorId int64
}
type CreateArticleResponse struct {
	Article *Article
}

type ListArticlesRequest struct {
	Cursor string
	Limit  int32
}
type ListArticlesResponse struct {
	Items      []*Article
	NextCursor string
}

// ArticleServiceServer is the gRPC service interface implemented by this package.
// protoc-gen-go-grpc would generate this exactly.
type ArticleServiceServer interface {
	GetArticle(context.Context, *GetArticleRequest) (*GetArticleResponse, error)
	CreateArticle(context.Context, *CreateArticleRequest) (*CreateArticleResponse, error)
	ListArticles(context.Context, *ListArticlesRequest) (*ListArticlesResponse, error)
}

// UnimplementedArticleServiceServer can be embedded for forward-compatibility.
type UnimplementedArticleServiceServer struct{}

func (UnimplementedArticleServiceServer) GetArticle(context.Context, *GetArticleRequest) (*GetArticleResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetArticle not implemented")
}
func (UnimplementedArticleServiceServer) CreateArticle(context.Context, *CreateArticleRequest) (*CreateArticleResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateArticle not implemented")
}
func (UnimplementedArticleServiceServer) ListArticles(context.Context, *ListArticlesRequest) (*ListArticlesResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListArticles not implemented")
}

// ----------------------------------------------------------------------
// Service implementation
// ----------------------------------------------------------------------

// ArticleServer implements ArticleServiceServer backed by the article use case.
type ArticleServer struct {
	UnimplementedArticleServiceServer
	svc *article.Service
}

// NewArticleServer wires the gRPC handler.
func NewArticleServer(svc *article.Service) *ArticleServer {
	return &ArticleServer{svc: svc}
}

// GetArticle returns a single article.
func (s *ArticleServer) GetArticle(ctx context.Context, req *GetArticleRequest) (*GetArticleResponse, error) {
	dto, err := s.svc.GetByID(ctx, req.Id)
	if err != nil {
		e := errcode.FromError(err)
		if e.Code == errcode.ErrNotFound.Code {
			return nil, status.Error(codes.NotFound, e.Message)
		}
		return nil, status.Error(codes.Internal, e.Message)
	}
	return &GetArticleResponse{Article: dtoToProto(dto)}, nil
}

// CreateArticle creates a new article.
func (s *ArticleServer) CreateArticle(ctx context.Context, req *CreateArticleRequest) (*CreateArticleResponse, error) {
	dto, err := s.svc.Create(ctx, article.CreateInput{
		Title: req.Title, Content: req.Content, AuthorID: req.AuthorId,
	})
	if err != nil {
		e := errcode.FromError(err)
		switch e.Code {
		case errcode.ErrBadRequest.Code, errcode.ErrInvalidParam.Code:
			return nil, status.Error(codes.InvalidArgument, e.Message)
		default:
			return nil, status.Error(codes.Internal, e.Message)
		}
	}
	return &CreateArticleResponse{Article: dtoToProto(dto)}, nil
}

// ListArticles returns a page of articles.
func (s *ArticleServer) ListArticles(ctx context.Context, req *ListArticlesRequest) (*ListArticlesResponse, error) {
	res, err := s.svc.List(ctx, req.Cursor, int(req.Limit))
	if err != nil {
		e := errcode.FromError(err)
		return nil, status.Error(codes.Internal, e.Message)
	}
	out := &ListArticlesResponse{Items: make([]*Article, 0, len(res.Items)), NextCursor: res.NextCursor}
	for _, d := range res.Items {
		out.Items = append(out.Items, dtoToProto(d))
	}
	return out, nil
}

// dtoToProto converts an application DTO to the gRPC message.
func dtoToProto(d *article.ArticleDTO) *Article {
	return &Article{
		Id:         d.ID,
		Title:      d.Title,
		Content:    d.Content,
		AuthorId:   d.AuthorID,
		AuthorName: d.AuthorName,
		CreatedAt:  d.CreatedAt.Unix(),
		UpdatedAt:  d.UpdatedAt.Unix(),
	}
}
