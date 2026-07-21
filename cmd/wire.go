// Package wire assembles all runtime dependencies (the composition root).
//
// Both cmd/rest and cmd/grpc call Wire(...) to build their dependency graph.
package wire

import (
	"context"
	"database/sql"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	gormlib "gorm.io/gorm"

	"github.com/zhanbinb/go-clean-arch_demo/internal/application/article"
	"github.com/zhanbinb/go-clean-arch_demo/internal/application/auth"
	"github.com/zhanbinb/go-clean-arch_demo/internal/application/author"
	"github.com/zhanbinb/go-clean-arch_demo/internal/infrastructure/config"
	persist "github.com/zhanbinb/go-clean-arch_demo/internal/infrastructure/persistence/gorm"
	httpiface "github.com/zhanbinb/go-clean-arch_demo/internal/interfaces/http"
	"github.com/zhanbinb/go-clean-arch_demo/internal/interfaces/http/handler"
	"github.com/zhanbinb/go-clean-arch_demo/pkg/jwt"
	"github.com/zhanbinb/go-clean-arch_demo/pkg/logger"
	"github.com/zhanbinb/go-clean-arch_demo/pkg/ratelimit"
)

// Deps bundles everything a server needs at runtime.
type Deps struct {
	Cfg       *config.Config
	Log       *logger.Logger
	DB        *gormlib.DB
	SQLDB     *sql.DB
	JWT       *jwt.Manager
	Limiter   *ratelimit.Limiter
	PromReg   *prometheus.Registry

	// Application services
	ArticleSvc *article.Service
	AuthorSvc  *author.Service
	AuthSvc    *auth.Service

	// HTTP handlers bundle
	HTTPHandlers *httpiface.Handlers
}

// New builds all runtime dependencies from configuration + environment.
func New(ctx context.Context, env string) (*Deps, error) {
	cfg, err := config.Load(env)
	if err != nil {
		return nil, err
	}
	log, err := logger.New(cfg.Log.Level, cfg.Log.Format)
	if err != nil {
		return nil, err
	}

	db, err := persist.NewDB(cfg.Database)
	if err != nil {
		return nil, err
	}
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	jwtMgr, err := jwt.New(cfg.JWT.Secret, cfg.JWT.TTL, cfg.JWT.RefreshTTL)
	if err != nil {
		return nil, err
	}

	var limiter *ratelimit.Limiter
	if cfg.RateLimit.Enabled {
		limiter = ratelimit.New(
			cfg.RateLimit.RPS,
			cfg.RateLimit.Burst,
			time.Duration(cfg.RateLimit.CleanupInterval)*time.Second,
		)
	}

	promReg := prometheus.NewRegistry()
	promReg.MustRegister(prometheus.NewGoCollector())
	promReg.MustRegister(prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}))

	// Repositories
	articleRepo := persist.NewArticleRepository(db)
	authorRepo := persist.NewAuthorRepository(db)
	userRepo := persist.NewUserRepository(db)

	// Application services
	articleSvc := article.NewService(articleRepo, authorRepo, log)
	authorSvc := author.NewService(authorRepo, log)
	authSvc := auth.NewService(userRepo, jwtMgr, log)

	// Handlers
	healthH := handler.NewHealthHandler(sqlDB)
	httpHandlers := httpiface.NewHandlers(articleSvc, authorSvc, authSvc, healthH)

	return &Deps{
		Cfg:          cfg,
		Log:          log,
		DB:           db,
		SQLDB:        sqlDB,
		JWT:          jwtMgr,
		Limiter:      limiter,
		PromReg:      promReg,
		ArticleSvc:   articleSvc,
		AuthorSvc:    authorSvc,
		AuthSvc:      authSvc,
		HTTPHandlers: httpHandlers,
	}, nil
}
