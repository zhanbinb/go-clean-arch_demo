package interfaces

import (
	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"github.com/zhanbinb/go-clean-arch_demo/internal/application/article"
	"github.com/zhanbinb/go-clean-arch_demo/internal/application/auth"
	"github.com/zhanbinb/go-clean-arch_demo/internal/application/author"
	"github.com/zhanbinb/go-clean-arch_demo/internal/infrastructure/config"
	"github.com/zhanbinb/go-clean-arch_demo/internal/interfaces/http/handler"
	"github.com/zhanbinb/go-clean-arch_demo/internal/interfaces/http/middleware"
	"github.com/zhanbinb/go-clean-arch_demo/pkg/jwt"
	"github.com/zhanbinb/go-clean-arch_demo/pkg/logger"
	"github.com/zhanbinb/go-clean-arch_demo/pkg/ratelimit"
)

// Handlers bundles all HTTP handlers for DI.
type Handlers struct {
	Health  *handler.HealthHandler
	Auth    *handler.AuthHandler
	Article *handler.ArticleHandler
	Author  *handler.AuthorHandler
}

// NewHandlers constructs all handlers from their services.
func NewHandlers(
	articleSvc *article.Service,
	authorSvc *author.Service,
	authSvc *auth.Service,
	health *handler.HealthHandler,
) *Handlers {
	return &Handlers{
		Health:  health,
		Auth:    handler.NewAuthHandler(authSvc),
		Article: handler.NewArticleHandler(articleSvc),
		Author:  handler.NewAuthorHandler(authorSvc),
	}
}

// NewRouter assembles the full Gin engine with all middleware and routes.
func NewRouter(
	cfg *config.Config,
	h *Handlers,
	log *logger.Logger,
	jwtMgr *jwt.Manager,
	limiter *ratelimit.Limiter,
	promReg *prometheus.Registry,
) *gin.Engine {
	if cfg.Server.Mode != "" {
		gin.SetMode(cfg.Server.Mode)
	}
	r := gin.New()

	// Global middleware (order matters!)
	r.Use(middleware.RequestID())
	r.Use(middleware.Logger(log))
	r.Use(middleware.Recovery(log))
	r.Use(middleware.CORS(cfg.CORS))
	r.Use(middleware.Metrics(promReg))

	// Health probes (no auth, no rate limit)
	r.GET("/healthz", h.Health.Liveness)
	r.GET("/readyz", h.Health.Readiness)
	r.GET("/metrics", gin.WrapH(promhttp.HandlerFor(promReg, promhttp.HandlerOpts{})))

	// Swagger UI (only when enabled)
	if cfg.Swagger.Enabled {
		swPath := cfg.Swagger.Path
		if swPath == "" {
			swPath = "/swagger"
		}
		r.GET(swPath+"/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}

	// pprof (only in debug mode)
	if cfg.Server.Mode == "debug" {
		pprof.Register(r)
	}

	// API v1
	v1 := r.Group("/api/v1")

	// Public auth endpoints
	authGroup := v1.Group("/auth")
	authGroup.POST("/login", h.Auth.Login)
	authGroup.POST("/refresh", h.Auth.Refresh)
	authGroup.POST("/register", h.Auth.Register)

	// Authenticated endpoints
	auth := v1.Group("")
	auth.Use(middleware.JWT(jwtMgr))
	if cfg.RateLimit.Enabled && limiter != nil {
		auth.Use(middleware.RateLimit(limiter, cfg.RateLimit.Dimension))
	}
	{
		auth.POST("/articles", h.Article.Create)
		auth.GET("/articles", h.Article.List)
		auth.GET("/articles/:id", h.Article.GetByID)
		auth.PUT("/articles/:id", h.Article.Update)
		auth.DELETE("/articles/:id", h.Article.Delete)

		auth.POST("/authors", h.Author.Create)
		auth.GET("/authors", h.Author.List)
		auth.GET("/authors/:id", h.Author.GetByID)
		auth.PUT("/authors/:id", h.Author.Update)
		auth.DELETE("/authors/:id", h.Author.Delete)
	}

	return r
}
