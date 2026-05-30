package main

import (
	"Project-IM/config"
	"Project-IM/internal/handler"
	"Project-IM/internal/hub"
	"Project-IM/internal/repository"
	"Project-IM/internal/router"
	"Project-IM/internal/service"
	jwtpkg "Project-IM/pkg/jwt"
	"Project-IM/pkg/logger"
	"Project-IM/pkg/rdb"
	"context"
	"errors"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	logger.Init()
	cfg := config.Load()

	db, err := gorm.Open(mysql.Open(cfg.DSN), &gorm.Config{})
	if err != nil {
		zap.L().Fatal("连接数据库失败", zap.Error(err))
	}
	if err = repository.AutoMigrate(db); err != nil {
		zap.L().Fatal("建表失败", zap.Error(err))
	}

	redisClient, err := rdb.NewClient(cfg.RedisAddr)
	if err != nil {
		zap.L().Fatal("连接Redis失败", zap.Error(err))
	}

	msgRepo := repository.NewMySQLMessageRepo(db)
	userRepo := repository.NewMySQLUserRepo(db)
	groupRepo := repository.NewMySQLGroupRepo(db)

	imHub := hub.NewHub(msgRepo, redisClient)
	go imHub.Run()
	go imHub.SubscribeRedis(context.Background())

	jwtMgr := jwtpkg.NewManager(cfg.JWTSecret)
	wsHandler := handler.NewHandler(imHub, msgRepo, groupRepo)
	userSvc := service.NewUserService(userRepo, jwtMgr)
	userHandler := handler.NewUserHandler(userSvc)
	groupSvc := service.NewGroupService(groupRepo)
	groupHandler := handler.NewGroupHandler(groupSvc, imHub)
	r := router.NewRouter(wsHandler, userHandler, groupHandler, jwtMgr)

	srv := &http.Server{
		Addr:    cfg.Addr,
		Handler: r,
	}

	go func() {
		zap.L().Info("IM Server starting on :8080")
		if err = srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			zap.L().Fatal("Server error", zap.Error(err))
		}
	}()

	// 优雅关闭：监听系统信号
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	<-ctx.Done()

	zap.L().Info("Shutting down server...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err = srv.Shutdown(shutdownCtx); err != nil {
		zap.L().Fatal("Server shutdown error", zap.Error(err))
	}
	zap.L().Info("Server exiting")
}
