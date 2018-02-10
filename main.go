package main

import (
	"net/http"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
)

func main() {
	// 顯示說明
	Usage()
	// 載入設定檔資料
	Conf = LoadConfig()

	// 設置 router
	router := SetupRouter()

	// 連接 websocket
	router.GET(Conf.App.Site, StartServerPush())

	// 測試用的頁面，正式使用時會註解掉
	if !strings.Contains(GetAppEnv(), "prod") {
		router.LoadHTMLGlob(GetAppRoot() + "/public/*")
		router.GET("/", func(c *gin.Context) {
			c.HTML(http.StatusOK, "index.html", nil)
		})
	}

	// // 建立 server
	server := CreateServer(router, Conf.App.Port, Conf.App.Host)

	waitFinish := new(sync.WaitGroup)
	waitFinish.Add(2)

	// 設置 Pusher 監聽
	go SetupPusherListen(router, waitFinish)

	// 系統信號監聽
	go SignalListenAndServe(server, waitFinish)

	waitFinish.Done()
}
