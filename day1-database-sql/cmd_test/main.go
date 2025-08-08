/*
-- 模块职责简图 --

   [ main.go ]
       │
       ▼
   [ geeorm.Engine ]
       ├── 初始化数据库连接
       └── 创建 Session
            │
            ▼
   [ session.Session ]
       ├── Raw(sql, args...)   → 构建 SQL
       ├── Exec / QueryRow / QueryRows → 执行 SQL
       └── 自动清空状态 + 日志记录
*/

package main

/*
   这个 main.go 文件被放在 day1-database-sql/cmd_test/ 目录，而不是项目顶层，是为了：
   作为测试/演示用的可执行程序，和框架核心代码隔离开来，保持项目结构清晰、职责分明。
*/

import (
	"geeorm" // 引入我们自己实现的 geeorm 包
	// 引入日志模块（虽然此处未使用，但会输出日志）
	"fmt" // 用于打印结果到控制台

	_ "github.com/mattn/go-sqlite3" // 导入 SQLite3 驱动（注册 init()，但不直接引用）
)

func main() {
	// 创建数据库引擎（连接数据库）
	// 使用 SQLite3 数据库，数据库文件为 gee.db
	engine, _ := geeorm.NewEngine("sqlite3", "gee.db")
	defer engine.Close() // main 函数结束前关闭数据库连接

	// 创建一个新的 Session，用于执行 SQL
	s := engine.NewSession()

	// 如果存在 User 表，则删除之（为了保证幂等性，每次运行都干净开始）
	_, _ = s.Raw("DROP TABLE IF EXISTS User;").Exec()

	// 创建 User 表，只包含一个 Name 字段
	_, _ = s.Raw("CREATE TABLE User(Name text);").Exec()

	// 再次尝试创建 User 表（实际上是重复操作，用于测试会不会报错）
	_, _ = s.Raw("CREATE TABLE User(Name text);").Exec()

	// 插入两条数据，使用占位符传参，防止 SQL 注入
	result, _ := s.Raw("INSERT INTO User(`Name`) values (?), (?)", "Tom", "Sam").Exec()

	// 获取受影响的行数
	count, _ := result.RowsAffected()

	// 打印结果
	fmt.Printf("Exec success, %d affected\n", count)
}

// 输出：
/*
[info ] 2025/08/06 19:39:43 geeorm.go:48: Connect database success
[info ] 2025/08/06 19:39:43 raw.go:46: DROP TABLE IF EXISTS User;  []
[info ] 2025/08/06 19:39:43 raw.go:46: CREATE TABLE User(Name text);  []
[info ] 2025/08/06 19:39:43 raw.go:46: CREATE TABLE User(Name text);  []
[error] 2025/08/06 19:39:43 raw.go:50: table User already exists
[info ] 2025/08/06 19:39:43 raw.go:46: INSERT INTO User(`Name`) values (?), (?)  [Tom Sam]
Exec success, 2 affected
[info ] 2025/08/06 19:39:43 geeorm.go:57: Close database success
*/
