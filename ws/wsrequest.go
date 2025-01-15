package ws

import (
	"hype-proxy/logger"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// Scrapy ws client pool
type Scrapy struct {
	Url              string
	Timeout          time.Duration
	MaxPool          int64
	RequestQueue     map[string]chan interface{}
	RequestQueueSync *sync.RWMutex
	WsPool           []*WsClient
	LastUsedPool     int
	sync.Mutex
	ResponseId       func(map[string]interface{}) (string, bool)
}

type Response struct {
	RequestId string
	Receive   chan interface{}
	Timeout   time.Duration
	scrapy    *Scrapy
}

// GetScrapy init pool
func GetScrapy(url string, timeout int64, maxPool int64) *Scrapy {
	scrapy := &Scrapy{
		Url:          url,
		Timeout:      time.Duration(timeout) * time.Second,
		MaxPool:      maxPool,
		RequestQueue: make(map[string]chan interface{}),
		RequestQueueSync: new(sync.RWMutex),
	}

	initPool := int64(2)
	if initPool > maxPool {
		initPool = maxPool
	}
	for i := int64(0); i < initPool; i++ {
		scrapy.AddWs(NewWsClient(url))
		logger.Infof("init wspool:%v", len(scrapy.WsPool))
		time.Sleep(5 * time.Second)
	}

	go func(uri string) {
		//维护 wspool
		for {
			time.Sleep(10 * time.Second)
			if len(scrapy.WsPool) < int(scrapy.MaxPool) {
				go func() {
					defer func() {
						if err := recover(); err != nil {
							logger.Errorf("%+v", err)
						}
					}()
					scrapy.AddWs(NewWsClient(uri))
				}()
			}
			logger.Infof("wspool len:%+v", len(scrapy.WsPool))
		}
	}(url)

	return scrapy
}

// RemoveWs 移除ws
func (s *Scrapy) RemoveWs(wc *WsClient) {
	s.Lock()
	defer s.Unlock()
	wc.Close()
	for i, w := range s.WsPool {
		if w == wc {
			s.WsPool = append(s.WsPool[0:i], s.WsPool[i+1:]...)
			logger.Debugf("client %v closed ", i)
			break
		}
	}
}

func (s *Scrapy) RemoveWsWithoutClose(wc *WsClient) {
	s.Lock()
	defer s.Unlock()
	for i, w := range s.WsPool {
		if w == wc {
			s.WsPool = append(s.WsPool[0:i], s.WsPool[i+1:]...)
			logger.Debugf("client %v closed ", i)
			break
		}
	}
}

// AddWs 创建ws
func (s *Scrapy) AddWs(wc *WsClient) int {
	if wc == nil {
		return -1
	}
	s.Lock()
	defer s.Unlock()
	index := len(s.WsPool)
	s.WsPool = append(s.WsPool, wc)
	go func(client *WsClient) {
		<-client.Closed
		s.RemoveWs(client)
	}(wc)

	//接收数据
	go func(i int, client *WsClient) {
	LOOP12:
		for {
			select {
			case <-client.Closed:
				logger.Infof("client %v: closed", i)
				break LOOP12
			case msg := <-client.Receive:
				{
					//log.Debugf("client %v: 收到ws的消息：%s", i, msg)
					var result map[string]interface{}
					var err error
					if err = json.Unmarshal([]byte(msg), &result); err != nil {
						client.Close()
						logger.Infof("client %v: 收到无法解析的数据：%s", i, msg)
						break LOOP12
					}

					if ping, ok := result["ping"]; ok {
						//失败后尝试两次pong
					LOOP6:
						for j := 1; ; j++ {
							select {
							case client.Send <- fmt.Sprintf(`{"pong":%v}`, int64(ping.(float64))):
								break LOOP6
							default:
								if j < 3 {
									time.Sleep(time.Second)
									logger.Infof("client %v: 重试pong %v次：%s", i, j, msg)
									continue LOOP6
								}
								client.Close()
								logger.Infof("client %v: 回复pong 失败 %v次：%s", i, j, msg)
								break LOOP12
							}
						}
						continue LOOP12
					}

					//log.Debugln("序列化后数据:", result)
					if reqId, ok1 := s.ResponseId(result); ok1 {
						//log.Infof("client %v: receive requestId:%v", i, reqId)
						if receive, ok2 := s.GetRequest(reqId); ok2 {
							select {
							case receive <- msg:
							default:
							}
							s.RemoveRequest(reqId)
						}
						continue LOOP12
					}
					//收到未知消息
					logger.Infof("client %v: 未知的消息:%s", i, msg)
				}
			}
		}
	}(index, wc)

	return index
}

// Close 关闭所有ws
func (s *Scrapy) Close() {
	for _, w := range s.WsPool {
		w.Close()
	}
}

// GetWs 获取一个ws
func (s *Scrapy) GetWs() (int, *WsClient) {
	if len(s.WsPool) == 0 {
		//s.AddWs(ws.NewWsClient(s.Url))
		return 0, nil
	}

	s.Lock()
	defer s.Unlock()
	s.LastUsedPool = s.LastUsedPool + 1
	if s.LastUsedPool >= len(s.WsPool) {
		s.LastUsedPool = 0
	}
	return s.LastUsedPool, s.WsPool[s.LastUsedPool]
}

func (s *Scrapy) GetRequest(reqId string) (chan interface{}, bool) {
	s.RequestQueueSync.RLock()
	defer s.RequestQueueSync.RUnlock()
	c, o := s.RequestQueue[reqId]
	return c, o
}

func (s *Scrapy) AddRequest(reqId string, val chan interface{}) {
	s.RequestQueueSync.Lock()
	defer s.RequestQueueSync.Unlock()
	s.RequestQueue[reqId] = val
}

func (s *Scrapy) RemoveRequest(reqId string) {
	s.RequestQueueSync.Lock()
	defer s.RequestQueueSync.Unlock()
	delete(s.RequestQueue, reqId)
}

// Request 发送命令
func (s *Scrapy) Request(reqId string, data string) (*Response, error) {
	if reqId == "" {
		return nil, fmt.Errorf("requestId:%s 不能为空", reqId)
	}

	if _, ok := s.GetRequest(reqId); ok {
		return nil, fmt.Errorf("重复的request:%s", data)
	}

	index, client := s.GetWs()
	if client == nil {
		return nil, fmt.Errorf("没有可用的链接request:%s", data)
	}

	receive := make(chan interface{})
	s.AddRequest(reqId, receive)
	sendTicker := time.NewTimer(10 * time.Second)
	defer sendTicker.Stop()

	//发送命令
	select {
	case client.Send <- data:
		//log.Infof("client %v send succ requestId:%s", index, reqId)
	case <-sendTicker.C:
		s.RemoveRequest(reqId)
		close(receive)
		return nil, fmt.Errorf("client %v send time out, request:%s", index, data)
	}

	return &Response{
		RequestId: reqId,
		scrapy:    s,
		Receive:   receive,
		Timeout:   s.Timeout,
	}, nil
}

// GetResponse 同步等待返回数据
func (r *Response) GetResponse(ret interface{}) (interface{}, error) {
	ticker := time.NewTimer(r.Timeout)
	defer func() {
		ticker.Stop()
		r.scrapy.RemoveRequest(r.RequestId)
		close(r.Receive)
	}()

	select {
	case resp := <-r.Receive:
		/*if resp.(map[string]interface{})["status"] != "ok" {
			return resp, fmt.Errorf("ws server return error:%v", resp.(map[string]interface{})["err-msg"])
		}*/

		if ret == nil {
			return resp, nil
		}
		return resp, json.Unmarshal([]byte(resp.(string)), ret)
	case <-ticker.C:
		return nil, fmt.Errorf("response time out")
	}
}
