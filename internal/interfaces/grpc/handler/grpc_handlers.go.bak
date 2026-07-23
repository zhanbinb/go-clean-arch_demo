package handler

import (
	"context"

	"google.golang.org/grpc"
)

// _ArticleService_*_Handler functions mimic the shims emitted by
// protoc-gen-go-grpc. They decode the request bytes, call the service method,
// and encode the response.

func _ArticleService_GetArticle_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetArticleRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ArticleServiceServer).GetArticle(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/article.v1.ArticleService/GetArticle",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ArticleServiceServer).GetArticle(ctx, req.(*GetArticleRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ArticleService_CreateArticle_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CreateArticleRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ArticleServiceServer).CreateArticle(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/article.v1.ArticleService/CreateArticle",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ArticleServiceServer).CreateArticle(ctx, req.(*CreateArticleRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ArticleService_ListArticles_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ListArticlesRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ArticleServiceServer).ListArticles(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/article.v1.ArticleService/ListArticles",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ArticleServiceServer).ListArticles(ctx, req.(*ListArticlesRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// ArticleService_ServiceDesc is the gRPC service descriptor (mirrors what
// protoc-gen-go-grpc would generate).
var ArticleService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "article.v1.ArticleService",
	HandlerType: (*ArticleServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{MethodName: "GetArticle", Handler: _ArticleService_GetArticle_Handler},
		{MethodName: "CreateArticle", Handler: _ArticleService_CreateArticle_Handler},
		{MethodName: "ListArticles", Handler: _ArticleService_ListArticles_Handler},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "article/v1/article.proto",
}

// init registers the default proto codec so the manually-rolled handlers work.
