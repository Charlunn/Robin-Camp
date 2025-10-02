package repository

import (
	"database/sql"
	"go_tutorial/model"
)

// mysqlUserRepository 是 UserRepository 接口针对 MySQL 的实现。
// 它持有一个 *sql.DB 数据库连接池对象。
type mysqlUserRepository struct {
	db *sql.DB
}

// NewMySQLUserRepository 是 mysqlUserRepository 的构造函数。
// 它接收一个数据库连接池作为参数。
func NewMySQLUserRepository(db *sql.DB) UserRepository {
	return &mysqlUserRepository{
		db: db,
	}
}

// Save 方法将一个用户插入到 MySQL 数据库中。
func (r *mysqlUserRepository) Save(user *model.User) error {
	// 我们使用 `Exec` 来执行一个不返回任何行的查询（如 INSERT, UPDATE, DELETE）。
	// `?` 是占位符，用于防止 SQL 注入。永远不要手动拼接字符串来构建查询！
	query := "INSERT INTO users (id, name, email) VALUES (?, ?, ?)"
	_, err := r.db.Exec(query, user.ID, user.Name, user.Email)
	return err
}

// FindByID 方法根据 ID 从 MySQL 数据库中查找一个用户。
func (r *mysqlUserRepository) FindByID(id string) (*model.User, error) {
	// 我们使用 `QueryRow` 来执行一个期望最多返回一行的查询。
	query := "SELECT id, name, email FROM users WHERE id = ?"
	row := r.db.QueryRow(query, id)

	user := &model.User{}

	// `row.Scan` 方法将查询结果的列值，按顺序扫描（复制）到 user 结构体的字段地址中。
	err := row.Scan(&user.ID, &user.Name, &user.Email)
	if err != nil {
		// sql.ErrNoRows 是一个标准错误，表示查询没有返回任何行。我们应该将其视为“未找到”，而不是一个程序错误。
		if err == sql.ErrNoRows {
			return nil, nil // 返回 nil, nil 来表示“未找到”
		}
		// 其他错误（如连接断开）则应该向上层报告。
		return nil, err
	}

	return user, nil
}
