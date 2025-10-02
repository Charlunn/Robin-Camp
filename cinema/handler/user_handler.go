package handler

import (
	"go_tutorial/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

// UserHandler 负责处理用户相关的 HTTP 请求。
// 它依赖于 UserService 来执行实际的业务逻辑。
type UserHandler struct {
	userService service.UserService
}

// NewUserHandler 是 UserHandler 的构造函数。
func NewUserHandler(svc service.UserService) *UserHandler {
	return &UserHandler{
		userService: svc,
	}
}

// CreateUserRequest 是用于绑定创建用户请求 JSON 的数据传输对象 (DTO)。
// 使用 DTO 可以将 API 的数据结构与内部的领域模型解耦。
type CreateUserRequest struct {
	Name  string `json:"name" binding:"required"`
	Email string `json:"email" binding:"required,email"`
}

// CreateUser 是处理创建用户请求的 Handler 方法。
// POST /users
func (h *UserHandler) CreateUser(c *gin.Context) {
	var req CreateUserRequest

	// 1. 解析和验证请求体
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 2. 调用 Service 层执行业务逻辑
	user, err := h.userService.CreateUser(req.Name, req.Email)
	if err != nil {
		// 在实际应用中，这里会根据错误类型返回不同的状态码
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建用户失败"})
		return
	}

	// 3. 返回成功的响应
	c.JSON(http.StatusCreated, user)
}

// GetUser 是处理获取用户请求的 Handler 方法。
// GET /users/:id
func (h *UserHandler) GetUser(c *gin.Context) {
	// 1. 从 URL 路径中获取参数
	id := c.Param("id")

	// 2. 调用 Service 层执行业务逻辑
	user, err := h.userService.GetUser(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取用户失败"})
		return
	}

	// 3. 处理用户不存在的情况
	if user == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "用户未找到"})
		return
	}

	// 4. 返回成功的响应
	c.JSON(http.StatusOK, user)
}
