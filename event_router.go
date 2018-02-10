package main

import (
	"fmt"
)

// PushEventManager 推送資訊的事件管理中心
func PushEventManager(pushData PushData) {
	mappingEventHandler := map[string]func(data interface{}) error{
		"ping": PingHandler,
	}
	WriteLog("INFO", fmt.Sprintf("✨  Server #%d 準備推送 event (%s) , data (%v) ✨", Conf.Pusher.ID, pushData.Event, pushData.Data))

	eventHandler, ok := mappingEventHandler[pushData.Event]
	if ok {
		WriteLog("INFO", fmt.Sprintf("✨  Server #%d 對應匹配 event (%s) 開始推送 ✨", Conf.Pusher.ID, pushData.Event))
		err := eventHandler(pushData.Data)
		if err != nil {
			WriteLog("ERROR", fmt.Sprintf("❌  Server #%d 推送 event (%s) 發生預期外 Error (%v) ❌", Conf.Pusher.ID, pushData.Event, err))
		}
		WriteLog("INFO", fmt.Sprintf("✨  Server #%d 推送OK event (%s) , data (%v) ✨", Conf.Pusher.ID, pushData.Event, pushData.Data))
	} else {
		WriteLog("WARNING", fmt.Sprintf("🎃  Server #%d 尚未登記 event (%s) 🎃", Conf.Pusher.ID, pushData.Event))
	}
}

// PingHandler 測試ws連線
func PingHandler(data interface{}) (err error) {
	defer func() {
		if catchErr := recover(); catchErr != nil {
			err = fmt.Errorf("Ping 發生預期外 Error (%v)", catchErr)
		}
	}()
	UserSocketMutex.RLock()
	for _, userSocket := range UserSocketList {
		userSocket.Channel <- []byte(data.(string))
	}
	UserSocketMutex.RUnlock()
	return
}
