package model

import (
	"errors"
	"sync"
	"time"
)

// User 用户模型
type User struct {
	ID        int64     `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Password  string    `json:"-"` // 密码不返回给客户端
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// UserStore 用户存储接口
type UserStore struct {
	mutex sync.RWMutex
	users map[int64]*User
	seq   int64 // 自增ID
}

// NewUserStore 创建用户存储
func NewUserStore() *UserStore {
	return &UserStore{
		users: make(map[int64]*User),
	}
}

// Create 创建用户
func (s *UserStore) Create(user *User) (*User, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// 检查用户名是否存在
	for _, u := range s.users {
		if u.Username == user.Username {
			return nil, errors.New("用户名已存在")
		}
		if u.Email == user.Email {
			return nil, errors.New("邮箱已存在")
		}
	}

	// 自增ID
	s.seq++
	user.ID = s.seq
	now := time.Now()
	user.CreatedAt = now
	user.UpdatedAt = now

	// 存储用户
	s.users[user.ID] = user
	return user, nil
}

// GetByID 根据ID获取用户
func (s *UserStore) GetByID(id int64) (*User, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	user, ok := s.users[id]
	if !ok {
		return nil, errors.New("用户不存在")
	}
	return user, nil
}

// GetAll 获取所有用户
func (s *UserStore) GetAll() []*User {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	users := make([]*User, 0, len(s.users))
	for _, user := range s.users {
		users = append(users, user)
	}
	return users
}

// Update 更新用户
func (s *UserStore) Update(id int64, update *User) (*User, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	user, ok := s.users[id]
	if !ok {
		return nil, errors.New("用户不存在")
	}

	// 检查用户名和邮箱是否已被其他用户使用
	for _, u := range s.users {
		if u.ID == id {
			continue
		}
		if u.Username == update.Username {
			return nil, errors.New("用户名已存在")
		}
		if u.Email == update.Email {
			return nil, errors.New("邮箱已存在")
		}
	}

	// 更新用户信息
	if update.Username != "" {
		user.Username = update.Username
	}
	if update.Email != "" {
		user.Email = update.Email
	}
	if update.Password != "" {
		user.Password = update.Password
	}
	user.UpdatedAt = time.Now()

	return user, nil
}

// Delete 删除用户
func (s *UserStore) Delete(id int64) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if _, ok := s.users[id]; !ok {
		return errors.New("用户不存在")
	}

	delete(s.users, id)
	return nil
}
