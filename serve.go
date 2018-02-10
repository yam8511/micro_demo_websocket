package main

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
)

// SetupRouter 配置路由
func SetupRouter() *gin.Engine {
	var router *gin.Engine

	if os.Getenv("GIN_MODE") == "debug" {
		router = gin.Default()
		// log.SetFlags(log.LstdFlags | log.Lshortfile)
	} else {
		gin.SetMode(gin.ReleaseMode)
		router = gin.New()
		router.Use(gin.Recovery())
	}

	return router
}

// CreateServer 建立伺服器
func CreateServer(router *gin.Engine, port, host string, args ...string) *http.Server {
	// 建立 Server
	server := &http.Server{
		Addr:         port,
		Handler:      router,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		// MaxHeaderBytes: 1 << 20,
	}

	return server
}

// SignalListenAndServe 開啟Server & 系統信號監聽
func SignalListenAndServe(server *http.Server, waitFinish *sync.WaitGroup) {
	defer func() {
		if err := recover(); err != nil {
			errMessage := fmt.Sprintf("❌  Server 發生意外 Error: %v ❌", err)
			WriteLog("ERROR", errMessage)
			NotifyEngineer(errMessage)
		}
	}()

	wg := new(sync.WaitGroup)
	wg.Add(1)
	sigs := make(chan os.Signal)
	signal.Notify(sigs, syscall.SIGKILL, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		err := server.ListenAndServe()
		WriteLog("WARNING", fmt.Sprintf("🎃  Server 回傳 error (%v) 🎃", err))
		wg.Done()
	}()

	go func() {
		receivedSignal := <-sigs
		fmt.Println()
		WriteLog("INFO", fmt.Sprintf("🎃  接受訊號 <- %v 🎃", receivedSignal))
		wg.Done()
	}()

	WriteLog("INFO", "🐠  Server Push 開始服務!🐠")
	defer WriteLog("INFO", "🔥  Server Push 結束服務!🔥")

	wg.Wait()
	waitFinish.Done()
}
