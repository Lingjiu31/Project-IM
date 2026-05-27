package main

import (
	"Project-IM/config"
	"Project-IM/internal/handler"
	"Project-IM/internal/hub"
	"Project-IM/internal/repository"
	"Project-IM/internal/service"
	jwtpkg "Project-IM/pkg/jwt"
	"Project-IM/pkg/logger"
	"context"
	"errors"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"Project-IM/internal/router"

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

	msgRepo := repository.NewMySQLMessageRepo(db)
	userRepo := repository.NewMySQLUserRepo(db)

	imHub := hub.NewHub(msgRepo)
	go imHub.Run()

	jwtMgr := jwtpkg.NewManager(cfg.JWTSecret)
	wsHandler := handler.NewHandler(imHub, msgRepo)
	userSvc := service.NewUserService(userRepo, jwtMgr)
	userHandler := handler.NewUserHandler(userSvc)
	r := router.NewRouter(wsHandler, userHandler, jwtMgr)

	srv := &http.Server{
		Addr:    cfg.Addr,
		Handler: r,
	}

	// 启动服务
	go func() {
		zap.L().Info("IM Server starting on :8080")
		if err = srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			zap.L().Fatal("Server error", zap.Error(err))
		}
	}()

	// 优雅关闭：监听系统信号
	// signal.NotifyContext 是 Go 1.16+ 的推荐写法，比老的 signal.Notify + channel 更简洁
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
