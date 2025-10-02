package repository

import (
	"go_tutorial/model"
	"sync"
)

// UserRepository 接口定义了与用户数据存储相关的操作。
// 通过定义接口，我们将业务逻辑与具体的数据实现（内存、MySQL、PostgreSQL等）解耦。
type UserRepository interface {
	Save(user *model.User) error
	FindByID(id string) (*model.User, error)
}

// inMemoryUserRepository 是 UserRepository 接口的一个内存实现。
// 它使用一个 map 来模拟数据库存储。`sync.RWMutex` 用于保证并发安全。
type inMemoryUserRepository struct {
	mtx   sync.RWMutex
	users map[string]*model.User
}

// NewInMemoryUserRepository 是 inMemoryUserRepository 的构造函数。
// 依赖注入时，我们总是通过构造函数来创建实例。
func NewInMemoryUserRepository() UserRepository {
	return &inMemoryUserRepository{
		users: make(map[string]*model.User),
	}
}

// Save 方法将一个用户保存到内存 map 中。
func (r *inMemoryUserRepository) Save(user *model.User) error {
	r.mtx.Lock()         // 写操作前加锁
	defer r.mtx.Unlock() // 函数结束时解锁

	r.users[user.ID] = user
	return nil
}

// FindByID 方法根据 ID 从内存 map 中查找一个用户。
func (r *inMemoryUserRepository) FindByID(id string) (*model.User, error) {
	r.mtx.RLock()         // 读操作前加读锁
	defer r.mtx.RUnlock() // 函数结束时解锁

	if user, ok := r.users[id]; ok {
		return user, nil
	}

	// 在实际应用中，这里会返回一个标准的 "not found" 错误
	return nil, nil
}
