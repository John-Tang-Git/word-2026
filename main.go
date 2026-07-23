package main

import (
	"word/auth"
	"word/database"
	"word/redis"
	"word/requests"
	usercontrol "word/userControl"

	"github.com/gin-gonic/gin"
)

func main() {
	// 初始化数据库
	database.InitDB()
	redis.InitRedis()

	// 默认路由
	route := gin.Default()

	// 登录
	route.POST("/login", usercontrol.Login())

	// 选择单词表页面
	route.GET("/index", auth.Authorization())

	// 个人信息页
	route.GET("/info", auth.Authorization(), requests.GetInfo())

	// 获取下一个学习单词
	route.GET("/study/:alphabet", auth.Authorization(), requests.GetStudy())

	// 完成某个单词的学习
	route.POST("/study/:alphabet", auth.Authorization(), requests.WordCounter(), requests.PostStudy())

	// 获取某个月的签到记录
	route.GET("/sign/:ym", auth.Authorization(), requests.GetSign())

	// 原神，启动！
	route.Run()
}
