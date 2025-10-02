package db

import (
	"database/sql"
	"log"

	_ "github.com/go-sql-driver/mysql" // 数据库驱动的导入现在也移到了这里！
)

// NewConnection 创建并返回一个新的数据库连接池。
// 这是数据库初始化的唯一入口点。
func NewConnection() (*sql.DB, error) {
	// 在真实的应用中，DSN 字符串应该来自配置文件或环境变量，而不是硬编码。
	dsn := "go_user:your_strong_password@tcp(127.0.0.1:3306)/go_tutorial_db?parseTime=true"

	// `sql.Open` 只是验证参数，不会建立连接。
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		// 如果 DSN 格式错误，这里会立即失败。
		return nil, err
	}

	// `db.Ping()` 尝试建立一个连接，以验证数据库是否可达。
	if err := db.Ping(); err != nil {
		// 如果数据库连不上（密码错误、地址错误等），这里会失败。
		db.Close() // 如果 ping 失败，最好关闭句柄以释放资源
		return nil, err
	}

	log.Println("Successfully connected to the database!")

	// 返回创建好的数据库连接池
	return db, nil
}
