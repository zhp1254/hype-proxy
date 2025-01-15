package main

import (
	"flag"
	"fmt"
	"hype-proxy/db"
	"hype-proxy/httpserver/routers"
	"hype-proxy/hype"
	"hype-proxy/hype/client"
	"hype-proxy/logger"
	"os"
	"strings"
	"time"
)

var (
	version    = "1.0.0"
	help       = flag.Bool("h", false, "show help")
	ogmiosIp   = flag.String("ws", "wss://api-ui.hyperliquid.xyz/ws", "hype ws server ip")
	rpcIp      = flag.String("rpc", "https://api-ui.hyperliquid.xyz", "hype rpc server ip")
	httpServer = flag.String("listen", "0.0.0.0:9009", "http rpc")
)

func init() {
	flag.Parse()
	fmt.Fprint(os.Stderr, "Version: ", version)
	fmt.Println("")
	if *help {
		flag.PrintDefaults()
		os.Exit(0)
	}

	ogmiosUrl := *ogmiosIp
	if !strings.HasPrefix(ogmiosUrl, "ws://") && !strings.HasPrefix(ogmiosUrl, "wss://") {
		ogmiosUrl = "ws://" + ogmiosUrl
	}

	if strings.HasPrefix(ogmiosUrl, "/") {
		ogmiosUrl = strings.TrimSuffix(ogmiosUrl, "/")
	}

	rpcUrl := *rpcIp
	if strings.HasPrefix(rpcUrl, "/") {
		rpcUrl = strings.TrimSuffix(rpcUrl, "/")
	}

	hype.Init(ogmiosUrl, rpcUrl)
}

func runSyncBlock() {
	for {
		func() {
			defer func() {
				if err := recover(); err != nil {
					logger.Errorf("+%v", err)
				}
			}()
			maxNum := int64(0)
			hype.SyncBlockTransfer(func(block []*client.Transfer) bool {
				for _, row := range block {
					//fmt.Println(row.BlockNumber, row.TxHash)
					//fmt.Println(row.User, row.Time)
					if maxNum < row.BlockNumber {
						maxNum = row.BlockNumber
					} else {
						fmt.Println("sync error blockNum: ", row.BlockNumber)
					}

					//fmt.Println(row.Action.Type, row.TxHash)
					if row.Action.Type != "SpotSend" && row.Action.Type != "UsdSend" {
						continue
					}

					if row.BlockNumber <= 0 {
						continue
					}
					// 交易入库
					fmt.Println("begin process transfer tx", row.TxHash)
					err := db.AddTx(&db.Tx{
						BlockNo:          row.BlockNumber,
						BlockTime:        row.Time,
						BlockHash:        "",
						TxHash:           row.TxHash,
						Amount:           row.Action.Amount,
						AddressFrom:      row.User,
						Destination:      row.Action.Destination,
						Token:            row.Action.Token,
						SignatureChainId: row.Action.SignatureChainId,
						HyperLiquidChain: row.Action.HyperliquidChain,
						ActionType:       row.Action.Type,
						Error:            fmt.Sprintf("%+v", row.Error),
					})

					if err != nil {
						logger.Error(err.Error())
					}
				}
				db.SetBestHeight(maxNum)
				//fmt.Println("sync last block no:", maxNum)
				return true
			})

		}()
		fmt.Println("runSyncBlock loop end")
		time.Sleep(time.Second * 5)
	}
}

func main() {
	//启动扫描块高线程
	go runSyncBlock()

	fmt.Println("start http server:", *httpServer)
	if err := routers.Init().Run(*httpServer); err != nil {
		fmt.Printf("start app occurs err: %v", err)
		os.Exit(0)
	}
}
