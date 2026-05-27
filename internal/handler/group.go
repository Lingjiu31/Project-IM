package handler

import (
	"Project-IM/internal/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type GroupHandler struct {
	svc *service.GroupService
}

func NewGroupHandler(svc *service.GroupService) *GroupHandler {
	return &GroupHandler{svc: svc}
}

type CreateGroupReq struct {
	Name string `json:"name"`
}

func (h *GroupHandler) CreateGroup(ctx *gin.Context) {
	var req CreateGroupReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	userID := ctx.GetInt64("user_id")
	group, err := h.svc.CreateGroup(ctx, req.Name, userID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"group_id": group.ID,
		"name":     group.Name,
		"message":  "创建成功",
	})
}

func (h *GroupHandler) JoinGroup(ctx *gin.Context) {
	idStr := ctx.Param("id")
	groupID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	userID := ctx.GetInt64("user_id")
	if err = h.svc.JoinGroup(ctx, groupID, userID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusCreated, gin.H{
		"message": "加入成功",
	})
}
