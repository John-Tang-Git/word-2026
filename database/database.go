package database

import (
	"fmt"
	"os"
	"word/config"

	"github.com/xuri/excelize/v2"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var (
	db  *gorm.DB
	err error
)

// 读取单词表文件
func ReadExcel(path string) ([]config.AlphabetDTO, error) {
	// 打开文件
	file, err := excelize.OpenFile(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	sheetName := file.GetSheetName(0)
	rows, err := file.GetRows(sheetName)
	if err != nil {
		return nil, err
	}

	var alphabets []config.AlphabetDTO

	for index, row := range rows {
		if index == 0 {
			continue
		}

		alphabet := config.AlphabetDTO{
			English: row[0],
			Chinese: row[1],
		}

		alphabets = append(alphabets, alphabet)
	}

	return alphabets, nil
}

func InsertAlphabetToDB(db *gorm.DB, tableName string, alphabets []config.AlphabetDTO) error {
	batchSize := 100
	totalSize := len(alphabets)

	for i := 0; i < totalSize; i += batchSize {
		end := i + batchSize
		if end > totalSize {
			end = totalSize
		}

		batch := alphabets[i:end]
		if err := db.Table(tableName).Create(&batch).Error; err != nil {
			return err
		}

		fmt.Printf("已经插入%d到%d条单词数据", end, totalSize)
	}
	return nil
}

// 两张单词表，在初次部署时，需要单独构建
func InitAlphabets() {
	// 先判断高考词汇是否录入
	var cee_count int64
	result := db.Model(&config.CeeInfo{}).Count(&cee_count)

	_, tmp_err := ReadExcel("./data/cee.xlsx")
	dir, _ := os.Getwd()
	if tmp_err != nil {
		fmt.Println("当前工作目录：", dir)
		fmt.Println("无法识别路径，错误:", tmp_err)
	}

	if result.Error != nil {
		fmt.Println("读取高考词汇mysql表错误，错误是：", result.Error)
	}
	if cee_count == 0 {
		// 确认表是空的
		fmt.Println("高考数据表是空的！")
		ceeDTO, err := ReadExcel("./data/cee.xlsx")
		if err != nil {
			fmt.Println("读取高考词汇excel表失败！错误：", err)
		}
		InsertAlphabetToDB(db, "cee_infos", ceeDTO)
	} else {
		fmt.Println("高考数据表存过了！有数据", cee_count)
	}

	// 再判断四级词汇是否录入
	var cet4_count int64
	result = db.Model(&config.CetFourInfo{}).Count(&cet4_count)
	if result.Error != nil {
		fmt.Println("读取四级词汇mysql表错误，错误是：", result.Error)
	}
	if cet4_count == 0 {
		// 确认表是空的
		fmt.Println("四级数据表是空的！")
		ceeDTO, err := ReadExcel("./data/cet4.xlsx")
		if err != nil {
			fmt.Println("读取四级词汇excel表失败！错误：", err)
		}
		InsertAlphabetToDB(db, "cet_four_infos", ceeDTO)
	} else {
		fmt.Println("四级数据表存过了! 有数据", cet4_count)
	}
}

// 初始化总数据库
func InitDB() *gorm.DB {
	sign_dsn := "root:Johntang2005@tcp(127.0.0.1:3306)/word?charset=utf8mb4&parseTime=True&loc=Local"
	db, err = gorm.Open(mysql.Open(sign_dsn), &gorm.Config{})

	if err != nil {
		fmt.Println("数据库连接失败！错误是：", err)
	} else {
		fmt.Println("数据库连接成功！")
	}

	// 自动迁移，一共七张表
	db.AutoMigrate(&config.UserInfo{}, &config.UserSign{}, &config.UserScore{}, &config.CeeInfo{}, &config.CetFourInfo{}, &config.UserProgress{}, &config.UserUnknown{})

	// 判断两张单词表是否已经录入mysql
	InitAlphabets()

	return db
}

func GetDB() *gorm.DB {
	return db
}
