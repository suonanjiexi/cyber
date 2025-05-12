package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/suonanjiexi/cyber"
	"github.com/suonanjiexi/cyber/example/model"
)

// UserHandler 用户处理器
type UserHandler struct {
	store *model.UserStore
}

// NewUserHandler 创建用户处理器
func NewUserHandler() *UserHandler {
	return &UserHandler{
		store: model.NewUserStore(),
	}
}

// CreateUser 创建用户
func (h *UserHandler) CreateUser(c *cyber.Context) {
	var user model.User
	if err := json.NewDecoder(c.Request.Body).Decode(&user); err != nil {
		c.Error(http.StatusBadRequest, "invalid_request", "无效的请求参数")
		return
	}

	// 验证必填字段
	if user.Username == "" || user.Email == "" || user.Password == "" {
		c.Error(http.StatusBadRequest, "missing_fields", "用户名、邮箱和密码不能为空")
		return
	}

	// 创建用户
	createdUser, err := h.store.Create(&user)
	if err != nil {
		c.Error(http.StatusBadRequest, "create_failed", err.Error())
		return
	}

	c.Success(http.StatusCreated, createdUser)
}

// GetUser 获取单个用户
func (h *UserHandler) GetUser(c *cyber.Context) {
	// 从URL参数中获取用户ID
	idStr := c.GetParam("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.Error(http.StatusBadRequest, "invalid_id", "无效的用户ID")
		return
	}

	// 获取用户
	user, err := h.store.GetByID(id)
	if err != nil {
		c.Error(http.StatusNotFound, "user_not_found", err.Error())
		return
	}

	c.Success(http.StatusOK, user)
}

// GetAllUsers 获取所有用户
func (h *UserHandler) GetAllUsers(c *cyber.Context) {
	users := h.store.GetAll()
	c.Success(http.StatusOK, users)
}

// UpdateUser 更新用户
func (h *UserHandler) UpdateUser(c *cyber.Context) {
	// 从URL参数中获取用户ID
	idStr := c.GetParam("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.Error(http.StatusBadRequest, "invalid_id", "无效的用户ID")
		return
	}

	// 解析请求体
	var user model.User
	if err := json.NewDecoder(c.Request.Body).Decode(&user); err != nil {
		c.Error(http.StatusBadRequest, "invalid_request", "无效的请求参数")
		return
	}

	// 更新用户
	updatedUser, err := h.store.Update(id, &user)
	if err != nil {
		c.Error(http.StatusBadRequest, "update_failed", err.Error())
		return
	}

	c.Success(http.StatusOK, updatedUser)
}

// DeleteUser 删除用户
func (h *UserHandler) DeleteUser(c *cyber.Context) {
	// 从URL参数中获取用户ID
	idStr := c.GetParam("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.Error(http.StatusBadRequest, "invalid_id", "无效的用户ID")
		return
	}

	// 删除用户
	if err := h.store.Delete(id); err != nil {
		c.Error(http.StatusNotFound, "delete_failed", err.Error())
		return
	}

	c.Success(http.StatusOK, map[string]string{"message": "用户删除成功"})
}
