package main

import (
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
	"gopkg.in/olahol/melody"
	"fmt"
)

// GopherInfo contains information about the gopher on screen
type GopherInfo struct {
	ID, X, Y string
}

func main() {
	router := gin.Default()
	mrouter := melody.New()
	gophers := make(map[*melody.Session]*GopherInfo)
	lock := new(sync.Mutex)
	counter := 0

	router.GET("/", func(c *gin.Context) {
		http.ServeFile(c.Writer, c.Request, "index.html")
	})

	router.GET("/ws", func(c *gin.Context) {  //处理网页端发起的ws连接，协议升级（主要包括握手）
		mrouter.HandleRequest(c.Writer, c.Request)
	})

	//上面mrouter.HandleRequest中最终调用的是此处的HandleConnect
	mrouter.HandleConnect(func(s *melody.Session) { //上面建立连接后，此处开始处理建立的连接,把此函数存到mrouter.connectHandler
		lock.Lock()
		for _, info := range gophers {
			s.Write([]byte("set " + info.ID + " " + info.X + " " + info.Y))
		}
		gophers[s] = &GopherInfo{strconv.Itoa(counter), "0", "0"}
		fmt.Println("=====HandleConnect=========","iam " + gophers[s].ID)
		s.Write([]byte("iam " + gophers[s].ID))
		counter++
		lock.Unlock()
	})

	mrouter.HandleDisconnect(func(s *melody.Session) {  //把此函数存到mrouter.disconnectHandler
		lock.Lock()
		mrouter.BroadcastOthers([]byte("dis "+gophers[s].ID), s)
		delete(gophers, s)
		lock.Unlock()
	})

	mrouter.HandleMessage(func(s *melody.Session, msg []byte) { //把此函数存到mrouter.messageHandler
		fmt.Println("=======HandleMessage========",string(msg))
		p := strings.Split(string(msg), " ")
		lock.Lock()
		info := gophers[s]
		if len(p) == 2 {
			info.X = p[0]
			info.Y = p[1]
			mrouter.BroadcastOthers([]byte("set "+info.ID+" "+info.X+" "+info.Y), s) //把坐标广播到所有连接上
		}
		lock.Unlock()
	})

	router.Run(":8080")
}
