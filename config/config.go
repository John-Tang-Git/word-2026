package config

// 1. 用户登录管理
type UserInfo struct {
	UserID   uint `gorm:"primarykey"`
	UserName string
	Password string
}

// 2. 用户签到
type UserSign struct {
	UserID     uint   `gorm:"primarykey"`
	YearMonth  string `gorm:"primarykey"`
	SignedBits string
}

// 3. 用户积分
type UserScore struct {
	UserID uint `gorm:"primarykey"`
	Score  uint
}

// 4. 两种单词表
type CeeInfo struct {
	WordID  uint `gorm:"primarykey"`
	English string
	Chinese string
}

type CetFourInfo struct {
	WordID  uint `gorm:"primarykey"`
	English string
	Chinese string
}

// 5. 用户进度
type UserProgress struct {
	UserID  uint `gorm:"primarykey"`
	Cee     uint
	CetFour uint
	CetSix  uint
}

// 6. 不会的单词
type UserUnknown struct {
	UnknownID uint `gorm:"primarykey"`
	UserID    uint
	Alphabet  string
	WordID    uint
}

type AlphabetDTO struct {
	English string
	Chinese string
}
