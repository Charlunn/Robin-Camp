package model

// User 是我们的核心领域模型，它精确地代表了我们系统中的一个用户实体。
// 它的字段通常与数据库中的表列相对应。
type User struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}
