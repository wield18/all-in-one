package entity

import (
	"fmt"
	"time"

	"gorm.io/gorm"
)

// 暴露一些可能用了验证的key名
// 尽量使用这里的
const (
	ID    = "id"
	Name  = "name"
	Email = "email"
	Token = "token"
)

type UserV1 struct {
	ID       uint32 `gorm:"primaryKey;autoIncrement"`
	Name     string `gorm:"type:varchar(32);not null;index"`                                   // 字符串，加索引
	Password string `gorm:"type:char(60);not null"`                                            // 字符串
	Email    string `gorm:"type:varchar(32);not null;unique"`                                  // 邮箱 唯一
	Status   string `gorm:"type:enum('happy','sad','anxious','angry','calm');default:'happy'"` // 枚举
	// todo: 这个东西只要登录了就是true,登出就为false,现在不管一致为true
	IsActive  bool      `gorm:"default:true"`   // 布尔值
	CreatedAt time.Time `gorm:"autoCreateTime"` // 自动创建时间
	UpdatedAt time.Time `gorm:"autoUpdateTime"` // 自动更新时间
	// 二进制数据
	Thumbnail []byte `gorm:"type:blob"` // 二进制数据
	// 软删除
	DeletedAt gorm.DeletedAt `gorm:"index"`

	// todo: 对应的所有fingerprint可能
	// 这必须外键一对多

}

// 这里只拿到了 Name Email 跟用户传的token 来生成map
// 整体小写规定
func (u *UserV1) ToMapToken(token string) map[string]string {
	return map[string]string{
		ID:    fmt.Sprintf("%d", u.ID),
		Name:  u.Name,
		Email: u.Email,
		Token: token,
	}
}

// // 浏览器生成唯一标识
// const fingerprint = {
//     userAgent: navigator.userAgent,        // 浏览器版本
//     platform: navigator.platform,           // 操作系统
//     timezone: Intl.DateTimeFormat().resolvedOptions().timeZone,
//     // canvasFingerprint: getCanvasFingerprint(), //  canvas 指纹
//     // webglFingerprint: getWebGLFingerprint(),   // WebGL 指纹
//     // fonts: getInstalledFonts(),                // 已安装字体
//     // plugins: getPlugins()                      // 浏览器插件
// }

// 注册的话这几个必填
type UserV1_allRequired struct {
	Name     string `form:"name" validate:"required"`
	Password string `form:"password" validate:"required"`
	Email    string `form:"email" validate:"required,email"`
	Captcha  string `form:"captcha" validate:"required"`
}

type UserV1_ExceptName struct {
	Name     string `form:"name"`
	Password string `form:"password" validate:"required"`
	Email    string `form:"email" validate:"required,email"`
	Captcha  string `form:"captcha" validate:"required"`
}

// 登录的话 我想啊
// 浏览器版本 操作系统 timeZone, 这三任意一个不在user常用的能找到 就需要email验证码
// 我这里不考虑ip
// email验证码登录时Name可以没有
// 基本登录时 Email Captcha 可以没有
type UserV1_passwordRequired struct {
	Name     string `form:"name"`
	Password string `form:"password" validate:"required"`
	Email    string `form:"email" validate:"email"`
	Captcha  string `form:"captcha"`
}

// 需要 id name status Thumbnail email newEmail password NewPassword captcha
type UserV1_update struct {
	ID       uint32 `form:"id" validate:"required"`
	Password string `form:"password"`

	Captcha      string `form:"captcha"`
	NewEmail     string `form:"new-email" validate:"email"`
	NewPassword  string `form:"new-password"`
	NewName      string `form:"new-name"`
	NewStatus    string `form:"new-status" validate:"oneof=happy sad anxious angry calm"`
	NewThumbnail string `form:"new-thumbnail"`
}

type UserV1_del struct {
	ID       uint32 `form:"id" validate:"required"`
	Password string `form:"password" validate:"required"`
	Email    string `form:"email" validate:"required,email"`
	Captcha  string `form:"captcha" validate:"required"`
}

type UserV1_loginout struct {
	ID uint32 `form:"id" validate:"required"`
}
