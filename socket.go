package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// UserSocketList 已建立 ws 連線的會員列表
var UserSocketList = map[uint64]*UserSocket{}

// UserSocketMutex 會員列表的讀寫鎖
var UserSocketMutex = new(sync.RWMutex)

// StartServerPush 開始 Server Push
func StartServerPush() gin.HandlerFunc {
	WriteLog("INFO", "✨  開啟 Websocket 服務 ✨")

	// 初始化 websocket
	upgrader := websocket.Upgrader{
		// 先允許所有的Origin都可以進來
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	// 連線次數
	var connectedCount uint64
	// 連線次數的讀寫鎖
	connectedMutex := new(sync.RWMutex)

	return func(c *gin.Context) {
		var currentCount uint64
		connectedMutex.Lock()
		connectedCount++
		currentCount = connectedCount
		connectedMutex.Unlock()

		WriteLog("INFO", fmt.Sprintf("✨  #%d 開始建立 Websocket 連線 ✨", currentCount))
		// 建立 websocket 連線
		ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		defer ws.Close()
		if err != nil {
			WriteLog("WARNING", fmt.Sprintf("🎃  #%d 建立 Websocket 連線失敗 (%v) 🎃", currentCount, err))
			return
		}

		// 讀取第一次的 WS 訊息
		WriteLog("INFO", fmt.Sprintf("✨  #%d Websocket 等待第一次訊息接收... ✨", currentCount))
		_, message, err := ws.ReadMessage()
		if err != nil {
			WriteLog("WARNING", fmt.Sprintf("🎃  #%d 建立 Websocket 第一次訊息接收 Error (%v) 🎃", currentCount, err))
			return
		}

		// 將第一次接收的 WS 訊息 轉換成 UserSocket
		// 期待訊息是 {"user_id": 123, "line_type": 3, "handicap_type": "0"}
		WriteLog("INFO", fmt.Sprintf("✨  #%d Websocket 接收第一次訊息 (%s)，開始解碼 ✨", currentCount, string(message)))
		userSocket := new(UserSocket)
		err = json.Unmarshal(message, &userSocket)
		if err != nil {
			WriteLog("WARNING", fmt.Sprintf("🎃  #%d Websocket 第一次訊息 (%s) 解碼失敗 (%v) 🎃", currentCount, string(message), err))
			return
		}
		WriteLog("INFO", fmt.Sprintf("✨  #%d Websocket 第一次訊息解碼OK --> 會員ID (%d)、水盤類別 (%d)、盤口類別 (%d) ✨", currentCount, userSocket.UserID, userSocket.LineType, userSocket.HandicapType))

		// 如果有設定【最大連線數量】，要先檢查數量
		if Conf.Pusher.MaxLink > 0 {
			var link int64
			UserSocketMutex.RLock()
			for _, existsUserSocket := range UserSocketList {
				if existsUserSocket.UserID == userSocket.UserID {
					link++
				}
			}
			UserSocketMutex.RUnlock()

			// 如果連線數量已超過【最大連線數量】
			if link >= Conf.Pusher.MaxLink {
				WriteLog("WARNING", fmt.Sprintf("🎃  #%d Websocket 關閉，連線數量已超過 %d 條 (User #%d) 🎃", currentCount, Conf.Pusher.MaxLink, userSocket.UserID))
				return
			}
		}

		// 建立通道
		userSocket.Channel = make(chan []byte)

		// 加入會員列表
		UserSocketMutex.Lock()
		UserSocketList[currentCount] = userSocket
		UserSocketMutex.Unlock()

		WriteLog("INFO", fmt.Sprintf("🦐  #%d Websocket 開始監聽資料通道 (User #%d) 🦐", currentCount, userSocket.UserID))

		waitSocket := new(sync.WaitGroup)
		waitSocket.Add(2)

		// 持續檢查 socket 是否還連線中
		go CheckSocketClosed(ws, waitSocket, userSocket, currentCount)

		// 持續監聽並推送訊息到前端
		go WaitMessageToPush(ws, waitSocket, userSocket, currentCount)

		waitSocket.Wait()

		UserSocketMutex.Lock()
		// 移除會員
		delete(UserSocketList, currentCount)
		UserSocketMutex.Unlock()

		WriteLog("INFO", fmt.Sprintf("💧  #%d Websocket 關閉資料通道 (User #%d) 💧", currentCount, userSocket.UserID))

		return
	}
}

// CheckSocketClosed 確認socket連線有無關閉
func CheckSocketClosed(ws *websocket.Conn, wg *sync.WaitGroup, userSocket *UserSocket, currentCount uint64) {
	defer func() {
		if err := recover(); err != nil {
			WriteLog("ERROR", fmt.Sprintf("❌  #%d Websocket -【CheckSocketClosed】發生預期外 Error (%v) (User #%d) ❌", currentCount, err, userSocket.UserID))
		}
		wg.Done()
	}()

	for {
		_, _, err := ws.NextReader()
		if err != nil {
			WriteLog("WARNING", fmt.Sprintf("🎃  #%d Websocket 連線已斷線 (User #%d) (%v) 🎃", currentCount, userSocket.UserID, err))
			// 關閉通道
			close(userSocket.Channel)
			return
		}
	}
}

// WaitMessageToPush 等待訊息去推送
func WaitMessageToPush(ws *websocket.Conn, wg *sync.WaitGroup, userSocket *UserSocket, currentCount uint64) {
	defer func() {
		if err := recover(); err != nil {
			WriteLog("ERROR", fmt.Sprintf("❌  #%d Websocket -【WaitMessageToPush】發生預期外 Error (%v) (User #%d) ❌", currentCount, err, userSocket.UserID))
		}
		wg.Done()
	}()
	for {
		byteMessage := <-userSocket.Channel
		err := ws.WriteMessage(websocket.TextMessage, byteMessage)
		if err != nil {
			WriteLog("WARNING", fmt.Sprintf("🎃  #%d Websocket 發送訊息失敗 (%s) (User #%d) (Error %v) 🎃", currentCount, string(byteMessage), userSocket.UserID, err))
			return
		}
	}
}
