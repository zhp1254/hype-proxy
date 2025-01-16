package hype

import (
	"encoding/json"
	"fmt"
	rpc "hype-proxy/hype/client"
	"hype-proxy/logger"
	"hype-proxy/ws"
	"reflect"
	"strings"
	"sync"
	"time"
)

var (
	wsUrl       string
	rpcClient   rpc.CustomClient
	hypeClient  []*rpc.HypeClient
	proxyIpList []string
	proxyLock   sync.RWMutex
	proxyIndex  int
)

func Init(_wsUrl, rpcUrl string) {
	fmt.Println("ws url :", _wsUrl)
	fmt.Println("rpc url :", rpcUrl)
	wsUrl = _wsUrl
	rpcClient = rpc.NewClient(rpcUrl, 10)
	go func() {
		for {
			func() {
				defer func() {
					if err := recover(); err != nil {
						fmt.Println(err)
					}
				}()
				proxyIpList = rpc.GetKuaidailiIp()
				proxyClient := make([]*rpc.HypeClient, 0)

				Par(15, proxyIpList, func(ip string) {
					cli := rpc.NewProxyClient(rpcUrl, ip, 10)
					if cli.Check() {
						fmt.Println(ip, " available: ")
						proxyClient = append(proxyClient, cli)
						return
					}
					//fmt.Println(ip, " not available: ")
				})

				if len(proxyClient) == 0 {
					return
				}
				proxyLock.Lock()
				defer proxyLock.Unlock()
				hypeClient = proxyClient
			}()
			time.Sleep(time.Minute * 3)
		}
	}()

	//hypeClient = make([]*rpc.HypeClient, 0)
	//hypeClient = append(hypeClient, rpc.NewProxyClient(rpcUrl, "", 10))
}

func GetHttpClient() *rpc.CustomClient {
	return &rpcClient
}

func GetProxyClient() (int, *rpc.HypeClient) {
	if len(hypeClient) == 0 {
		return 0, nil
	}

	proxyLock.RLock()
	defer proxyLock.RUnlock()
	proxyIndex += 1
	if proxyIndex >= len(hypeClient) {
		proxyIndex = 0
	}
	return proxyIndex, hypeClient[proxyIndex]
}

func RemoveProxyClient(index int) {
	proxyLock.Lock()
	defer proxyLock.Unlock()

	l := len(hypeClient)-1
	if index > l {
		return
	}

	hypeClient[index] = hypeClient[l]
	hypeClient = hypeClient[:l]
}

func SyncBlockHeader(handle RollForwardHandle) {
	client := ws.NewWsClient(wsUrl)
	client.SendMsg(`{"method":"subscribe","subscription":{"type":"explorerBlock"}}`)
	for resp := range client.Receive {
		if strings.Index(resp, "subscriptionResponse") != -1 {
			continue
		}

		var ret []*rpc.BlockHeader
		err := json.Unmarshal([]byte(resp), &ret)
		if err != nil {
			logger.Errorf("SyncBlock err:%+v", err, resp)
			//goto GOTO
			continue
		}
		handle(ret)
	}
}

func SyncBlockTransfer(handle RollTransferHandle) {
	client := ws.NewWsClient(wsUrl)
	client.SendMsg(`{"method":"subscribe","subscription":{"type":"explorerTxs"}}`)
	for resp := range client.Receive {
		if strings.Index(resp, "subscriptionResponse") != -1 {
			continue
		}

		//fmt.Println(string(resp))
		var ret []*rpc.Transfer
		err := json.Unmarshal([]byte(resp), &ret)
		if err != nil {
			logger.Errorf("SyncBlock err:%+v", err, resp)
			//goto GOTO
			continue
		}
		handle(ret)
	}
}

func Par(concurrency int, arr interface{}, f interface{}) {
	throttle := make(chan struct{}, concurrency)
	var wg sync.WaitGroup

	varr := reflect.ValueOf(arr)
	l := varr.Len()

	rf := reflect.ValueOf(f)

	wg.Add(l)
	for i := 0; i < l; i++ {
		throttle <- struct{}{}

		go func(i int) {
			defer wg.Done()
			defer func() {
				<-throttle
			}()
			rf.Call([]reflect.Value{varr.Index(i)})
		}(i)
	}

	wg.Wait()
}
