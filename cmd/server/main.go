package main

import (
	"Project-IM/config"
	"Project-IM/internal/handler"
	"Project-IM/internal/hub"
	"Project-IM/internal/repository"
	"Project-IM/internal/service"
	"Project-IM/internal/ws"
	"context"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"Project-IM/internal/router"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	cfg := config.Load()
	db, err := gorm.Open(mysql.Open(cfg.DSN), &gorm.Config{})
	if err != nil {
		log.Fatalf("连接数据库失败: %v", err)
	}

	msgRepo := repository.NewMySQLMessageRepo(db)
	if err = msgRepo.InitTable(); err != nil {
		log.Fatalf("建表失败: %v", err)
	}
	userRepo := repository.NewMySQLUserRepo(db)
	if err = userRepo.InitTable(); err != nil {
		log.Fatalf("建表失败: %v", err)
	}

	imHub := hub.NewHub(msgRepo)
	go imHub.Run()

	wsHandler := ws.NewHandler(imHub)
	userSvc := service.NewUserService(userRepo)
	userHandler := handler.NewUserHandler(userSvc)
	r := router.NewRouter(wsHandler, userHandler)

	srv := &http.Server{
		Addr:    cfg.Addr,
		Handler: r,
	}

	// 启动服务
	go func() {
		log.Println("IM Server starting on :8080")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// 优雅关闭：监听系统信号
	// signal.NotifyContext 是 Go 1.16+ 的推荐写法，比老的 signal.Notify + channel 更简洁
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	<-ctx.Done()
	log.Println("Shutting down server...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Forced shutdown: %v", err)
	}

	log.Println("Server exited cleanly")
}
