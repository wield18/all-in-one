package entity

type CaptchaEmail struct {
	Email string `form:"email" valudate:"required, email"`
}
