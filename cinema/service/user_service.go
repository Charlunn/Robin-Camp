package service

import (
	"go_tutorial/model"
	"go_tutorial/repository"

	"github.com/google/uuid"
)

// UserService 接口定义了用户相关的业务逻辑。
// 它是 handler 层和 repository 层之间的桥梁。	ype UserService interface {
	CreateUser(name, email string) (*model.User, error)
	GetUser(id string) (*model.User, error)
}

// userServiceImpl 是 UserService 接口的实现。
// 它包含业务逻辑，并依赖于 UserRepository 来进行数据持久化。
type userServiceImpl struct {
	userRepo repository.UserRepository
}

// NewUserService 是 userServiceImpl 的构造函数。
// 它接收一个 UserRepository 接口作为参数，这就是“依赖注入”。
// 通过这种方式，我们可以在不修改业务逻辑代码的情况下，切换底层的数据存储实现。
func NewUserService(repo repository.UserRepository) UserService {
	return &userServiceImpl{
		userRepo: repo,
	}
}

// CreateUser 方法实现了创建用户的业务逻辑。
func (s *userServiceImpl) CreateUser(name, email string) (*model.User, error) {
	// 1. 业务逻辑: 生成一个新的唯一 ID
	userID := uuid.New().String()

	// 2. 创建领域模型对象
	newUser := &model.User{
		ID:    userID,
		Name:  name,
		Email: email,
	}

	// 3. 调用 Repository 层进行数据持久化
	if err := s.userRepo.Save(newUser); err != nil {
		// 在实际应用中，这里可能会记录日志或包装错误
		return nil, err
	}

	// 4. 返回创建好的用户对象
	return newUser, nil
}

// GetUser 方法实现了获取用户的业务逻辑。
func (s *userServiceImpl) GetUser(id string) (*model.User, error) {
	// 对于简单的“读”操作，Service 层可能只是直接调用 Repository 层
	return s.userRepo.FindByID(id)
}
