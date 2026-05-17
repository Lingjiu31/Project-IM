package main

import (
	"Project-IM/internal/hub"
	"Project-IM/internal/ws"
	"context"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"Project-IM/internal/router"
)

func main() {
	h := hub.NewHub()
	go h.Run()

	w := ws.NewHandler(h)
	r := router.NewRouter(w)

	srv := &http.Server{
		Addr:    ":8080",
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
