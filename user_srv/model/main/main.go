package main

import (
	"crypto/sha512"
	"fmt"
	"log"
	"mytest/user_srv/model"
	"os"
	"time"

	"github.com/anaskhan96/go-password-encoder"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

//md5加密
// func genMd5(code string) string {
// 	md5 := md5.New()
// 	_, _ = io.WriteString(md5, code)
// 	return hex.EncodeToString(md5.Sum(nil))
// }

func main() {
	dsn := "root:12345678@/myshop_user_srv?charset=utf8mb4&parseTime=True&loc=Local"

	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
		logger.Config{
			SlowThreshold: time.Second, // 慢 SQL 阈值
			LogLevel:      logger.Info, // Log level
			Colorful:      true,        // 禁用彩色打印
		},
	)

	// 全局模式
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
		Logger: newLogger,
	})
	if err != nil {
		panic(err)
	}

	options := &password.Options{16, 100, 32, sha512.New}
	salt, encodedPwd := password.Encode("admin123", options)
	newPassword := fmt.Sprintf("$pbkdf2-sha512$%s$%s", salt, encodedPwd)
	fmt.Println(newPassword)

	for i := 0; i < 10; i++ {
		user := model.User{
			// Basemodel: model.Basemodel{
			// 	UpdateAt: time.Now(),
			// 	CreateAt: time.Now(),
			// },
			NickName: fmt.Sprintf("bobby%d", i),
			Mobile:   fmt.Sprintf("1878222222%d", i),
			Password: newPassword,
		}
		db.Save(&user)
	}

	// _ = db.AutoMigrate(&model.User{})

	//fmt.Println(genMd5("123456"))

	// Using the default options
	// salt, encodedPwd := password.Encode("generic password", nil)
	// fmt.Println(salt, ",", encodedPwd)
	// check := password.Verify("generic password", salt, encodedPwd, nil)
	// fmt.Println(check) // true

	// Using custom options
	// options := &password.Options{10, 100, 32, sha512.New}
	// salt, encodedPwd := password.Encode("generic password", options)
	// pwd := fmt.Sprintf("$pkbdf2-sha512$%s$%s", salt, encodedPwd)
	// // fmt.Println(salt)
	// // fmt.Println(encodedPwd)
	// fmt.Println(len(pwd))

	//check password
	// passwordinfo := strings.Split(pwd, "$")
	// check := password.Verify("generic password", passwordinfo[2], passwordinfo[3], options)
	// fmt.Println(check) // true

}
