/*
raw.go 文件提供了一个安全、可复用的会话对象，用于构建和执行底层的原始 SQL 语句。
  - 支持构建 SQL 的链式调用方式。
  - 自动管理参数绑定，防止 SQL 注入。
  - 每次执行完毕自动 Clear，避免状态污染。
  - 使用自定义日志输出 SQL 语句与执行错误，便于调试。package session
*/

package session

import (
	"database/sql" // Go 标准库，数据库接口
	"geeorm/log"   // 自定义日志模块
	"strings"      // 用于构建 SQL 语句
)

// Session 是 ORM 的基础结构，封装对数据库的操作。
// 它持有一个 *sql.DB 指针，并记录即将执行的 SQL 语句及其参数。
type Session struct {
	db      *sql.DB         // 持有一个数据库连接池的指针，所有的数据库操作都将通过这个 db 实例进行
	sql     strings.Builder // 一个用于拼接 SQL 字符串的构建器（相比于'+'拼接，性能更好）
	sqlVars []interface{}   // 一个切片，用于存储 SQL 查询中的变量参数。这是防止 SQL 注入攻击的关键，它会将参数与 SQL 语句分开传递
}

// New 构造函数，接受一个 *sql.DB 实例，并返回一个新的 Session 实例指针
func New(db *sql.DB) *Session {
	return &Session{db: db}
}

// Clear方法 用于每次数据库操作完，重置会话状态，防止脏数据污染下次使用
func (s *Session) Clear() {
	s.sql.Reset()   // 清除已构建的 SQL 语句
	s.sqlVars = nil // 清空参数切片
}

// DB方法 一个简单的获取方法，返回底层的 *sql.DB 实例
func (s *Session) DB() *sql.DB {
	return s.db
}

// Exec 执行构造好的 SQL（用于 INSERT、UPDATE、DELETE 等）
// 封装原生 Exec 方法，记录日志；最后自动清理状态，这样Session可以复用，开启一次会话，可以执行多次 SQL
func (s *Session) Exec() (result sql.Result, err error) {
	defer s.Clear() // 保证无论成功或失败，状态都会被清理

	log.Info(s.sql.String(), s.sqlVars) // 用你自定义的日志包记录 SQL 语句和参数，便于调试。

	// 调用底层 *sql.DB 的 Exec() 方法来执行 SQL
	if result, err = s.DB().Exec(s.sql.String(), s.sqlVars...); err != nil { // s.sql.String() 获取完整的 SQL 字符串; sqlVars 参数列表
		log.Error(err) // 打印错误日志
	}
	return
}

// QueryRow 执行可能返回单行结果的查询（如 SELECT * FROM user WHERE id=1 LIMIT 1）
// 封装原生 Exec 方法，记录日志；最后自动清理状态，这样Session可以复用，开启一次会话，可以执行多次 SQL
func (s *Session) QueryRow() *sql.Row {
	defer s.Clear()

	log.Info(s.sql.String(), s.sqlVars)

	return s.DB().QueryRow(s.sql.String(), s.sqlVars...)
}

// QueryRows 查询多行结果（如 SELECT * FROM user）
func (s *Session) QueryRows() (rows *sql.Rows, err error) {
	defer s.Clear()

	log.Info(s.sql.String(), s.sqlVars)

	if rows, err = s.DB().Query(s.sql.String(), s.sqlVars...); err != nil {
		log.Error(err)
	}
	return
}

// Raw方法 用于构建 SQL 语句，将原始 SQL 字符串和参数写入 session 中
// 支持链式调用，例如：
// session.Raw("SELECT * FROM users WHERE name = ?", "Tom").QueryRow()
func (s *Session) Raw(sql string, values ...interface{}) *Session {
	s.sql.WriteString(sql)                   // 将传入的 SQL 字符串片段追加到 strings.Builder 中。
	s.sql.WriteString(" ")                   // 追加一个空格，以确保 SQL 语句的各个部分之间有正确的间隔。
	s.sqlVars = append(s.sqlVars, values...) // 添加参数
	return s                                 // 返回会话自身的指针，允许开发者使用'方法链'的方式来构建复杂的 SQL 语句
}
