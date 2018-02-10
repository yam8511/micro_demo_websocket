package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/naoina/toml"
)

// Usage 顯示說明
func Usage() {
	if GetAppEnv() == "" || GetAppSite() == "" {
		fmt.Printf(`
			📖 IN體育的 Websocket 說明 📖
			需傳入以下環境變數：

			⚙  PROJECT_ENV : 專案環境
				✏ dev 開發
				✏ prod 正式

			⚙  PROJECT_SITE : 專案端口
				✏ game 遊戲
				✏ user 玩家

			📌  舉例： PROJECT_ENV=dev PROJECT_SITE=game ./websocket_demo
`)
		os.Exit(0)
	} else {
		WriteLog("INFO", fmt.Sprintf("⚙  PROJECT_ENV: %s", GetAppEnv()))
		WriteLog("INFO", fmt.Sprintf("⚙  PROJECT_SITE: %s", GetAppSite()))
	}
}

// GetAppEnv 取環境變數
func GetAppEnv() string {
	return os.Getenv("PROJECT_ENV")
}

// GetAppSite 取賽程控客端變數
func GetAppSite() string {
	return os.Getenv("PROJECT_SITE")
}

// GetAppRoot 取專案的根目錄
func GetAppRoot() string {
	root, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		WriteLog("WARNING", "🎃 GetAppRoot：取根目錄失敗，自動抓取 PROJECT_ROOT 的環境變數 🎃")
		return os.Getenv("PROJECT_ROOT")
	}
	return root
}

// LoadConfig 載入 config
func LoadConfig() *Config {
	var configBody *Config
	configFile := GetAppRoot() + "/config/" + GetAppEnv() + "/" + GetAppSite() + ".toml"
	tomlData, readFileErr := ioutil.ReadFile(configFile)
	if readFileErr != nil {
		log.Fatalf("❌ 讀取Config錯誤： %v ❌", readFileErr)
	}

	err := toml.Unmarshal(tomlData, &configBody)
	if err != nil {
		log.Fatalf("❌ 載入Config錯誤： %v ❌", err)
	}
	return configBody
}

// WriteLog 寫Log記錄檔案
func WriteLog(tag string, msg string) {
	//設定時間
	now := time.Now()

	// 組合字串
	logStr := now.Format("[2006-01-02 15:04:05]") + "【" + tag + "】" + msg + "\n"
	log.Print(logStr)

	// 設定檔案位置
	fileName := "melon-server.log"
	folderPath := GetAppRoot() + now.Format("/logs/2006-01-02/15/")

	//檢查今日log檔案是否存在
	if _, err := os.Stat(folderPath + fileName); os.IsNotExist(err) {
		//建立資料夾
		os.MkdirAll(folderPath, 0777)
		//建立檔案
		_, err := os.Create(folderPath + fileName)
		if err != nil {
			log.Printf("❌ WriteLog: 建立檔案錯誤 [%v] ❌ \n----> %s\n", err, msg)
			return
		}
	}

	//開啟檔案準備寫入
	logFile, err := os.OpenFile(folderPath+fileName, os.O_RDWR|os.O_APPEND, 0777)
	defer logFile.Close()
	if err != nil {
		log.Printf("❌ WriteLog: 開啟檔案錯誤 [%v] ❌ \n----> %s\n", err, msg)
		return
	}

	_, err = logFile.WriteString(logStr)

	if err != nil {
		log.Printf("❌ WriteLog: 寫入檔案錯誤 [%v] ❌ \n----> %s\n", err, msg)
	}
}

// NotifyEngineer 通知工程師
func NotifyEngineer(msg string) {
	WriteLog("INFO", fmt.Sprintf("✨ 通知訊息： %v ✨\n", msg))
	// if Bot == nil {
	// 	WriteLog("WARNING", fmt.Sprintf("通知訊息失敗： 機器人尚未設定 [%s]\n", msg))
	// 	return
	// }

	// message := tgbotapi.NewMessage(Conf.Bot.ChatID, msg)
	// _, err := Bot.Send(message)

	// if err != nil {
	// 	WriteLog("WARNING", fmt.Sprintf("通知訊息發送失敗： %v\n", err))
	// }
}
