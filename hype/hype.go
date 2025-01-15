package hype

import (
	"encoding/json"
	"fmt"
	rpc "hype-proxy/hype/client"
	"hype-proxy/logger"
	"hype-proxy/ws"
	"strings"
)

var (
	wsUrl     string
	rpcClient rpc.CustomClient
)

func Init(_wsUrl, rpcUrl string) {
	fmt.Println("ws url :", _wsUrl)
	fmt.Println("rpc url :", rpcUrl)
	wsUrl = _wsUrl
	rpcClient = rpc.NewClient(rpcUrl, 30)
}

func GetHttpClient() *rpc.CustomClient {
	return &rpcClient
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
