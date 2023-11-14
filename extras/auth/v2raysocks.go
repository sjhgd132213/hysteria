package auth

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/apernet/hysteria/core/server"
)

var _ server.Authenticator = &V2RaySocksApiProvider{}

type V2RaySocksApiProvider struct {
	Client *http.Client
	URL    string
}

// 用户列表
var (
	usersMap map[string]User
	lock     sync.Mutex
)

type User struct {
	ID         int     `json:"id"`
	UUID       string  `json:"uuid"`
	SpeedLimit *uint32 `json:"speed_limit"`
}
type ResponseData struct {
	Users []User `json:"users"`
}

func getUserList(url string) ([]User, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var responseData ResponseData
	err = json.NewDecoder(resp.Body).Decode(&responseData)
	if err != nil {
		return nil, err
	}

	return responseData.Users, nil
}

func UpdateUsers(url string, interval time.Duration) {

	fmt.Println("用户列表自动更新服务已激活")

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			userList, err := getUserList(url)
			if err != nil {
				fmt.Println("Error:", err)
				continue
			}
			lock.Lock()
			usersMap = make(map[string]User)
			for _, user := range userList {
				usersMap[user.UUID] = user
			}
			lock.Unlock()
		}
	}
}

// 验证代码
func (v *V2RaySocksApiProvider) Authenticate(addr net.Addr, auth string, tx uint64) (ok bool, id string) {

	// 获取判断连接用户是否在用户列表内
	lock.Lock()
	defer lock.Unlock()

	if user, exists := usersMap[auth]; exists {
		return true, strconv.Itoa(user.ID)
	}
	return false, ""
}
