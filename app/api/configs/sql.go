package configs

import (
	User "LanshanSummerProject/app/api/internal/model/user"
	"fmt"
	"os"

	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var (
	Db  *gorm.DB
	Cli *redis.Client
)

type DBConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
}

func LoadDBConfig() (*DBConfig, error) {
	// 设置配置文件路径
	viper.SetConfigFile("app/api/configs/config.yaml")

	// 读取配置文件
	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}

	// 解析 database 部分到结构体
	var cfg DBConfig
	if err := viper.UnmarshalKey("database", &cfg); err != nil {
		return nil, fmt.Errorf("解析配置失败: %w", err)
	}

	// 可选：验证必要字段不为空
	if cfg.Host == "" || cfg.User == "" || cfg.Password == "" {
		return nil, fmt.Errorf("配置不完整，请检查 database 部分")
	}

	return &cfg, nil
}

func (c *DBConfig) DSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		c.User, c.Password, c.Host, c.Port, c.DBName)
}

func InitDB() {
	MyDbConfig, err := LoadDBConfig()
	if err != nil {
		Logger.Fatal("InitDb", zap.Error(err))
	}
	Db, err = gorm.Open(mysql.Open(MyDbConfig.DSN()), &gorm.Config{})
	if err != nil {
		Logger.Fatal("InitDb", zap.Error(err))
	}
	err = Db.AutoMigrate(&User.User{})
	if err != nil {
		Logger.Fatal("InitDb", zap.Error(err))
	}
	redisHost := os.Getenv("REDIS_HOST")
	if redisHost == "" {
		redisHost = "127.0.0.1"
	}
	redisPort := os.Getenv("REDIS_PORT")
	if redisPort == "" {
		redisPort = "6379"
	}
	Cli = redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%s", redisHost, redisPort),
	})
}
