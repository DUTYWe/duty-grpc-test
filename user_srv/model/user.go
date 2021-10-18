package model

import (
	"time"

	"gorm.io/gorm"
)

type Basemodel struct {
	ID        int32     `gorm:"primarykey"`
	CreateAt  time.Time `gorm:"column:add_time;default:CURRENT_TIMESTAMP"`
	UpdateAt  time.Time `gorm:"column:update_time;default:CURRENT_TIMESTAMP"` //默认为当前时间
	DeleteAt  gorm.DeletedAt
	IsDeleted bool
}

/*
1、密文
2、密文不可反解
非对称加密：md5算法
*/

type User struct {
	Basemodel
	Mobile   string     `gorm:"index:idx_mobile;unique;type:varchar(11);not null"`
	Password string     `gorm:"type:varchar(100);not null"`
	NickName string     `gorm:"type:varchar(20)"`
	Birthday *time.Time `gorm:"type:datetime"`
	Gender   string     `gorm:"common:gender;default:male;type:varchar(6) comment 'female表示女，male表示男'"` //性别female/male
	Role     int        `gorm:"column:role;default:1;type:int comment '1表示普通用户，2表示管理员'"`
}
