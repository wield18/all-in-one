// Package mysql 当前只用来登录验证
package mysql

import (
	"fmt"

	"github.com/wield18/all-in-one/config"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DB 全局数据库连接实例
type MDBServer struct {
	DB *gorm.DB
}

var mdb = MDBServer{}

func GetMDBServer() *MDBServer {
	return &mdb
}

func GetMDB() *gorm.DB {
	return mdb.DB
}

func NewMDBServer(dbConfig *config.Mysql) *MDBServer {
	// var dbConfig = config.Config.Db

	// 构建数据库连接DSN(数据源名称)
	// 格式: "用户名:密码@tcp(主机:端口)/数据库名?charset=字符集&parseTime=True&loc=Local"
	url := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local",
		dbConfig.Username,
		dbConfig.Password,
		dbConfig.Host,
		dbConfig.Port,
		dbConfig.Db,
		dbConfig.Charset)

	// 使用GORM打开MySQL数据库连接
	db, err := gorm.Open(mysql.Open(url), &gorm.Config{
		// 设置日志模式为Info级别，显示执行的SQL语句
		Logger: logger.Default.LogMode(logger.Info),
		// 迁移时禁用外键约束
		DisableForeignKeyConstraintWhenMigrating: true,
	})

	// 检查连接错误
	if err != nil {
		panic(err)
	}

	// 检查GORM错误
	if db.Error != nil {
		panic(db.Error)
	}

	// 获取底层的sql.DB对象以设置连接池参数
	sqlDB, _ := db.DB()

	// 设置连接池参数
	// SetMaxIdleConns: 设置最大空闲连接数
	sqlDB.SetMaxIdleConns(dbConfig.MaxIdle)
	// SetMaxOpenConns: 设置最大打开连接数
	sqlDB.SetMaxOpenConns(dbConfig.MaxOpen)
	// log.Println("数据库连接正常")

	// 写入
	mdb.DB = db
	return &mdb
}

func (db *MDBServer) Shutdown() error {

	return nil
}
