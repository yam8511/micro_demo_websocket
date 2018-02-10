package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
)

// SetupPusherListen 設置 Pusher 功能
func SetupPusherListen(router *gin.Engine, waitFinish *sync.WaitGroup) {
	wg := new(sync.WaitGroup)
	wg.Add(1)
	sigs := make(chan os.Signal)
	signal.Notify(sigs, syscall.SIGKILL, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		receivedSignal := <-sigs
		fmt.Println()
		WriteLog("INFO", fmt.Sprintf("🎃  接受訊號 <- %v 🎃", receivedSignal))
		wg.Done()
	}()

	go func() {
		for Conf.Pusher.ID == 0 {
			// 註冊 Pusher
			err := RegisterPusher()
			if err != nil {
				WriteLog("WARNING", fmt.Sprintf("🎃  註冊 Pusher 失敗 (%v) 🎃", err))
				time.Sleep(time.Second * 2)
			} else {
				WriteLog("INFO", fmt.Sprintf("✨  註冊 Pusher 成功：ID %d ✨", Conf.Pusher.ID))

				// 註冊 pusher callback
				router.PUT("/pusher", PusherCallback)
			}
		}
	}()

	wg.Wait()

	// 當 Server 關閉時，需註銷 Pusher
	if Conf.Pusher.ID != 0 {
		err := DeregisterPusher()
		if err != nil {
			WriteLog("WARNING", fmt.Sprintf("🎃  註銷 Pusher 失敗：ID %d (%v) 🎃", Conf.Pusher.ID, err))
		} else {
			WriteLog("INFO", fmt.Sprintf("✨  註銷 Pusher 成功：ID %d ✨", Conf.Pusher.ID))
		}
	}

	waitFinish.Done()
}

// RegisterPusher 註冊伺服器推送服務
func RegisterPusher() error {
	type RegisterInfo struct {
		IP       string `json:"ip"`
		Host     string `json:"host"`
		Port     string `json:"port"`
		APIKey   string `json:"api_key"`
		ChatRoom string `json:"chat_room"`
	}

	type RegisterResponse struct {
		ErrorCode int64  `json:"error_code"`
		ErrorText string `json:"error_text"`
		ID        uint64 `json:"id"`
	}

	type RPCRequest struct {
		Service string       `json:"service"`
		Method  string       `json:"method"`
		Request RegisterInfo `json:"request"`
	}

	rpcRequest := RPCRequest{
		Service: Conf.Pusher.RPC.Register.Service,
		Method:  Conf.Pusher.RPC.Register.Method,
		Request: RegisterInfo{
			IP:       Conf.App.IP,
			Host:     Conf.App.Host,
			Port:     Conf.App.Port,
			ChatRoom: Conf.Pusher.ChatRoom,
		},
	}

	prePayload, err := json.Marshal(rpcRequest)
	if err != nil {
		return err
	}

	url := "http://" + Conf.Pusher.IP + Conf.Pusher.Port + "/rpc"

	body, err := CurlRPC(string(prePayload), url, Conf.Pusher.Host, Conf.Pusher.APIKey)
	if err != nil {
		return err
	}

	var registerRes RegisterResponse
	err = json.Unmarshal(body, &registerRes)
	if err != nil {
		return err
	}

	if registerRes.ErrorCode != 0 {
		return errors.New(registerRes.ErrorText)
	}

	Conf.Pusher.ID = registerRes.ID
	return nil
}

// DeregisterPusher 註銷伺服器推送服務
func DeregisterPusher() error {
	type DeregisterInfo struct {
		IP     string `json:"ip"`
		Host   string `json:"host"`
		Port   string `json:"port"`
		APIKey string `json:"api_key"`
		ID     uint64 `json:"id"`
	}

	type DeregisterResponse struct {
		ErrorCode int64  `json:"error_code"`
		ErrorText string `json:"error_text"`
	}

	type RPCRequest struct {
		Service string         `json:"service"`
		Method  string         `json:"method"`
		Request DeregisterInfo `json:"request"`
	}

	rpcRequest := RPCRequest{
		Service: Conf.Pusher.RPC.Deregister.Service,
		Method:  Conf.Pusher.RPC.Deregister.Method,
		Request: DeregisterInfo{
			IP:   Conf.App.IP,
			Host: Conf.App.Host,
			Port: Conf.App.Port,
			ID:   Conf.Pusher.ID,
		},
	}

	prePayload, err := json.Marshal(rpcRequest)
	if err != nil {
		return err
	}

	url := "http://" + Conf.Pusher.IP + Conf.Pusher.Port + "/rpc"

	body, err := CurlRPC(string(prePayload), url, Conf.Pusher.Host, Conf.Pusher.APIKey)
	if err != nil {
		return err
	}

	var deregisterRes DeregisterResponse
	err = json.Unmarshal(body, &deregisterRes)
	if err != nil {
		return err
	}

	if deregisterRes.ErrorCode != 0 {
		return errors.New(deregisterRes.ErrorText)
	}

	return nil
}

// CurlRPC 呼叫微服務
func CurlRPC(prePayload, url, host, apiKey string) (body []byte, err error) {
	payload := strings.NewReader(prePayload)

	req, _ := http.NewRequest("POST", url, payload)

	req.Header.Add("host", Conf.Pusher.Host)
	req.Header.Add("api-key", Conf.Pusher.APIKey)
	req.Header.Add("content-type", "application/json")

	res, curlErr := http.DefaultClient.Do(req)
	if curlErr != nil {
		err = curlErr
		return
	}

	defer res.Body.Close()
	body, err = ioutil.ReadAll(res.Body)
	return
}

// PusherCallback Pusher的回呼func
func PusherCallback(c *gin.Context) {

	// 接收請求
	WriteLog("INFO", fmt.Sprintf("✨  Server #%d 收到 Pusher 推送請求 (%v) ✨", Conf.Pusher.ID, c.Request.Body))

	// 解析請求
	body, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		WriteLog("WARNING", fmt.Sprintf("🎃  Server #%d 解析請求 Error (%v) 🎃", Conf.Pusher.ID, err))
		c.String(http.StatusOK, fmt.Sprintf("Error (%v)", err))
		return
	}
	WriteLog("INFO", fmt.Sprintf("✨  Server #%d 解析請求 OK (%s) ✨", Conf.Pusher.ID, string(body)))

	// 解析推送資料
	var pushData PushData
	err = json.Unmarshal(body, &pushData)
	if err != nil {
		WriteLog("WARNING", fmt.Sprintf("🎃  Server #%d (%s) JSON 解碼 Error (%v) 🎃", Conf.Pusher.ID, string(body), err))
		c.String(http.StatusOK, fmt.Sprintf("Error (%v)", err))
		return
	}

	// 協程推送資料
	go func(pushData PushData) {
		// 如果事件名稱不為空才能推送
		if pushData.Event != "" {
			PushEventManager(pushData)
		}
	}(pushData)

	c.String(http.StatusOK, "ok")
	return
}
