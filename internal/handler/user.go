package handler

import (
	"Project-IM/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	svc *service.UserService
}

func NewUserHandler(svc *service.UserService) *UserHandler {
	return &UserHandler{
		svc: svc,
	}
}

type UserReq struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (h *UserHandler) Register(ctx *gin.Context) {
	var req UserReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.svc.Register(ctx, req.Username, req.Password); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusCreated, gin.H{"message": "注册成功"})
}
func (h *UserHandler) Login(ctx *gin.Context) {
	var req UserReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	tokenStr, err := h.svc.Login(ctx, req.Username, req.Password)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"token": tokenStr, "message": "登录成功"})
}
