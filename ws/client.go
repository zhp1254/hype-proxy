package ws

import (
	"bytes"
	"hype-proxy/logger"
	"compress/gzip"
	"context"
	"github.com/gorilla/websocket"
	"io/ioutil"
	"time"
)

type WsClient struct {
	Url        string
	Closed     chan int
	isClosed   bool
	EnableGizp bool
	Send       chan string
	Receive    chan string
	conn       *websocket.Conn
}

func (ws *WsClient) Close() {
	defer func() {
		if err := recover(); err != nil {
			//log.Debug(err)
		}
	}()
	if !ws.isClosed {
		ws.isClosed = true
		close(ws.Closed)

		if ws.conn != nil {
			ws.conn.Close()
		}
	}
}

// SendMsg block to send msg
func (ws *WsClient) SendMsg(msg string) {
	ws.Send <- msg
}

// ReceiveMsg block to receive msg
func (ws *WsClient) ReceiveMsg() string {
	msg := <-ws.Receive
	return msg
}

func NewWsClient(url string) *WsClient {
	var dialer *websocket.Dialer
	ctx, _ := context.WithTimeout(context.Background(), 30 * time.Second)
	conn, _, err := dialer.DialContext(ctx, url, nil)
	if err != nil {
		panic(err)
		//logger.Errorf("err: %+v", err)
		//return nil
	}

	client := &WsClient{
		Url:      url,
		Closed:   make(chan int),
		isClosed: false,
		Send:     make(chan string),
		Receive:  make(chan string),
		conn:     conn,
	}

	go client.processSend()
	go client.processReceive()
	return client
}

// 读取data 并发送到server
func (ws *WsClient) processSend() {
	defer func() {
		if e := recover(); e != nil {
			logger.Errorf("%+v", e)
		}
		ws.Close()
	}()

	for {
		select {
		case <-ws.Closed:
			return
		case data := <-ws.Send:
			{
				err := ws.conn.WriteMessage(websocket.TextMessage, []byte(data))
				if err != nil {
					logger.Errorf("send msg fail:%+v", data)
					return
				}
				logger.Debugf("send msg succ:%+v", data)
			}
		}
	}
}

// 接收server data 并放到chan
func (ws *WsClient) processReceive() {
	defer func() {
		if e := recover(); e != nil {
			logger.Errorf("%+v", e)
		}
		ws.Close()
	}()

	for {
		_, data, err := ws.conn.ReadMessage()
		if err != nil {
			logger.Debugf("ws read err:%+v", err)
			return
		}

		var msg []byte
		if ws.EnableGizp {
			msg, err = ungzip(data)
			if err != nil {
				logger.Errorf("ungzip receive data error: %+v", err)
				continue
			}
		} else {
			msg = data
		}

		logger.Debugf("receive msg:%+v", string(msg))
		select {
		case <-ws.Closed:
			return
		case ws.Receive <- string(msg):
		}
	}
}

func ungzip(in []byte) ([]byte, error) {
	if len(in) > 0 {
		reader := bytes.NewReader(in)
		r, err1 := gzip.NewReader(reader)
		if err1 != nil {
			return nil, err1
		}

		out, err2 := ioutil.ReadAll(r)
		if err2 != nil {
			return nil, err2
		}
		return out, nil
	}
	return []byte{}, nil
}
