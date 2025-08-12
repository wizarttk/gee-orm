/*
geeorm.go 是整个框架的入口点和核心。它负责数据库连接的建立、测试、管理和关闭，并提供了一个方法来创建用于具体数据库操作的 Session 实例。
    - 初始化数据库连接；
    - 封装 *sql.DB 为 Engine；
    - 提供创建 Session 的方法；
    - 管理数据库连接的生命周期（打开、关闭）；
*/

package geeorm

import (
	"database/sql"   // Go 标准数据库驱动接口
	"geeorm/log"     // 自定义日志模块
	"geeorm/session" // 会话封装模块
)

// Engine 是 ORM 的核心结构体，
// 主要职责是管理数据库连接（db）并创建 Session。
type Engine struct {
	db *sql.DB // 这个字段持有一个数据库连接池的指针，所有后续的数据库操作都将通过它进行。
}

// NewEngine 用于初始化一个 Engine 实例，建立数据库连接。
//
// 参数：
//   - driver: 驱动名称（如 "sqlite3", "mysql"）
//   - source: 数据库连接字符串（如 SQLite 文件路径、MySQL DSN）
//
// 返回：
//   - e: 初始化后的 *Engine 实例
//   - err: 错误信息，如果连接失败
func NewEngine(driver, source string) (e *Engine, err error) {
	// 第一步：打开数据库连接（不代表立刻建立连接）
	db, err := sql.Open(driver, source)
	if err != nil {
		log.Error(err)
		return
	}

	// 第二步：使用 Ping() 检查数据库是否真的连通
	if err = db.Ping(); err != nil {
		log.Error(err)
		return
	}

	// 第三步：创建 Engine 实例并返回
	e = &Engine{db: db}
	log.Info("Connect database success")
	return
}

// Close方法 关闭数据库连接，释放资源。
// 为什么要单独写 Close() 方法？
//   - 因为我们希望 ORM 有统一的管理接口，而不是直接操作 sql.DB
func (engine *Engine) Close() {
	if err := engine.db.Close(); err != nil {
		log.Error("Failed to close database") // 数据库关闭失败
	}
	log.Info("Close database success") // 日志输出：成功关闭
}

// NewSession方法 创建一个新的 Session 实例，供 ORM 操作使用。
// 每次调用都将返回一个全新的 Session 实例，它与 Engine 共享同一个数据库连接池，但拥有独立的 SQL 构建状态
// Session 内部持有 db，可以构建并执行 SQL 语句。
func (engine *Engine) NewSession() *session.Session {
	return session.New(engine.db)
}

/* 示例用法：
   engine, err := geeorm.NewEngine("sqlite3", "gee.db")
   if err != nil {
       panic("数据库连接失败")
   }
   defer engine.Close()

   session := engine.NewSession()
   session.Raw("CREATE TABLE User(Name text, Age integer)").Exec()
*/
