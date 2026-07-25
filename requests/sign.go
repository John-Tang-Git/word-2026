package requests

import (
	"context"
	"fmt"
	"strconv"
	"time"
	"word/config"
	"word/redis"

	"github.com/gin-gonic/gin"
)

// 根据当前年月，返回上一个年月
func LastYearMonth(ym string) string {
	ym_int, _ := strconv.Atoi(ym)
	result_int := 0
	if ym[4:] == "01" {
		// 如果是1月，就必须把年份减一
		result_int = ym_int - 89
	} else {
		// 如果不是1月，只需要-1即可
		result_int = ym_int - 1
	}
	return strconv.Itoa(result_int)
}

// 根据年月，返回天数
func HowManyDays(ym string) string {

	// 先解析出年和月
	year := ym[:4]
	month := ym[4:]

	// 判断平年闰年
	int_year, _ := strconv.Atoi(year)
	isLeap := int_year%4 == 0 && (int_year%100 != 0 || int_year%400 == 0)

	// 根据年份与月份返回对应天数（字符串格式）
	var result string
	switch month {
	case "01":
		result = "31"
	case "02":
		if isLeap {
			result = "29"
		} else {
			result = "28"
		}
	case "03":
		result = "31"
	case "04":
		result = "30"
	case "05":
		result = "31"
	case "06":
		result = "30"
	case "07":
		result = "31"
	case "08":
		result = "31"
	case "09":
		result = "30"
	case "10":
		result = "31"
	case "11":
		result = "30"
	case "12":
		result = "31"
	}

	return result
}

// 封装成函数：统计某个月（不是当前月）连续签到了多少天
func MonthConstantSignedDays(userID int, ym string) (int, bool) {

	// 先统计提供的年月是不是当前月
	if time.Now().Format("200601") == ym {
		fmt.Println("不能统计当前年月！")
		return -1, false
	}

	// 不是当前年月，可以统计
	day, _ := strconv.Atoi(HowManyDays(ym))
	rctx := context.Background()
	rdb := redis.GetRedis()
	redis_key := "sign:" + strconv.Itoa(userID) + ":" + ym

	// 统计该月连续签到天数
	result, _ := rdb.BitField(rctx, redis_key, "GET", "u"+HowManyDays(ym), 0).Result()
	bits := result[0]

	cnt := 0
	for {
		if (bits & 1) == 0 {
			break
		}
		bits >>= 1
		cnt++
	}

	// 判断这个月是否签满
	isFull := false
	if cnt == day {
		isFull = true
	}

	return cnt, isFull
}

func ConstantSignedDays(userID int, day int, ym string) int {

	// 获取rdb
	rdb := redis.GetRedis()
	rctx := context.Background()

	// 获取当月记录
	sign_key := "sign:" + strconv.Itoa(userID) + ":" + ym
	result, err := rdb.BitField(rctx, sign_key, "GET", "u"+strconv.Itoa(day), 0).Result()
	if err != nil {
		fmt.Println("统计连续签到天数时发生错误！错误：", err)
	}
	bits := result[0]

	// 统计当前连续签到天数
	// 先统计当前年月
	isFull := false // isFull用来表示当前月份是否签满
	cnt := 0
	for {
		if (bits & 1) == 0 { // 如果当前这一天是0，直接断了
			break
		}
		cnt++
		bits = bits >> 1 // 右移一位
	}
	if cnt == day {
		isFull = true
	}
	// 再统计过往年月
	tmp_ym := ym
	cnt_add := 0
	for {
		if !isFull {
			break
		}
		tmp_ym = LastYearMonth(tmp_ym)
		cnt_add, isFull = MonthConstantSignedDays(userID, tmp_ym)
		cnt += cnt_add
	}

	return cnt
}

// 加分函数，根据签到情况加分
func AddScore(userID int, day int, ym string) {

	// 获取rdb
	rdb := redis.GetRedis()
	rctx := context.Background()

	// 拼接redis key
	score_key := "score:" + strconv.Itoa(userID)

	// 首先，日常签到加10分
	rdb.IncrBy(rctx, score_key, 10)

	// 再根据连续签到天数加分
	bonus_score := 0
	// 先获取连续签到天数
	constant_signed_days := ConstantSignedDays(userID, day, ym)
	// 将连续签到天数存入redis
	constant_key := "constant:" + strconv.Itoa(userID)
	rdb.Set(rctx, constant_key, constant_signed_days, time.Second*3600*24*365) // 一年过期
	// 再根据连续签到天数加bonus
	switch constant_signed_days {
	case 2:
		bonus_score = 5
	case 5:
		bonus_score = 20
	case 10:
		bonus_score = 50
	case 20:
		bonus_score = 200
	}
	rdb.IncrBy(rctx, score_key, int64(bonus_score))
}

// 单日学习够50词就签到
func Sign(userID int, ym string, day int) {

	// 获取rdb
	rdb := redis.GetRedis()
	rctx := context.Background()

	// 根据给出的userID和年月拼接redis key
	sign_key := "sign:" + strconv.Itoa(userID) + ":" + ym

	// 签到操作
	_, err := rdb.SetBit(rctx, sign_key, int64(day-1), 1).Result()
	if err != nil {
		fmt.Println("当天签到操作失败！错误：", err)
		return
	}

	// 日常签到加分
	AddScore(userID, day, ym)
}

// 按月份获取签到记录
func GetSign() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// 获取URL参数
		ym := ctx.Param("ym")

		// 从上下文获取用户id
		user, _ := ctx.Get("user")
		userID := user.(config.UserInfo).UserID

		// 获取rdb
		rdb := redis.GetRedis()
		rctx := context.Background()

		// 构建本月份对应的redis key
		ym_key := "sign:" + strconv.Itoa(int(userID)) + ":" + ym
		fmt.Println("ym_key:", ym_key)

		// 从redis中，获取当前年月对应的记录
		result, err := rdb.BitField(rctx, ym_key, "GET", "u"+HowManyDays(ym), 0).Result()
		if err != nil {
			fmt.Println("获取当前年月的签到记录失败！错误：", err)
		}
		bits := result[0]
		day, _ := strconv.Atoi(HowManyDays(ym))
		bitsStr := fmt.Sprintf("%0*b", day, bits) // bitsStr是当前月签到记录对应的0101字符串

		// 获取当前连续签到天数
		constant_key := "constant:" + strconv.Itoa(int(userID))
		constant_result, err := rdb.Get(rctx, constant_key).Result()
		if err != nil {
			fmt.Println("获取该用户连续签到天数失败！错误：", err)
		}
		constant_signed_days := constant_result

		// 返回JSON
		ctx.JSON(200, gin.H{
			"code":               0,
			"info":               "获取该月签到记录成功！",
			"bitsStr":            bitsStr,
			"constantSignedDays": constant_signed_days,
		})
	}
}

// 中间件：学习单词计数器
func WordCounter() gin.HandlerFunc {
	return func(ctx *gin.Context) {

		// 这是一个中间件，每次学完一个单词（会or不会），都给当天的redis key + 1
		rdb := redis.GetRedis()
		rctx := context.Background()

		// 从上下文获取用户id
		user, _ := ctx.Get("user")
		userID := user.(config.UserInfo).UserID

		// 拼接redis key
		cur_date := time.Now().Format("20060102")
		counter_key := "counter:" + strconv.Itoa(int(userID)) + ":" + cur_date

		// 给当天该用户的单词学习计数器加一
		newVal, err := rdb.Incr(rctx, counter_key).Result()
		if err != nil {
			fmt.Println("给当前用户学习单词书+1时出错！错误：", err)
			ctx.Abort()
		}

		// 如果newVal已经达到50，就达到签到标准
		if newVal == 5 {
			ym := cur_date[:6]
			day, _ := strconv.Atoi(cur_date[6:])
			Sign(int(userID), ym, day)
		}

		// 中间件继续
		ctx.Next()
	}
}
