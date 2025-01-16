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
	version      = "1.0.0"
	help         = flag.Bool("h", false, "show help")
	thread       = flag.Int("t", 2, "thread number")
	ogmiosIp     = flag.String("ws", "wss://api-ui.hyperliquid.xyz/ws", "hype ws server ip")
	rpcIp        = flag.String("rpc", "https://api-ui.hyperliquid.xyz", "hype rpc server ip")
	httpServer   = flag.String("listen", "0.0.0.0:9009", "http rpc")
	processChain chan *client.BlockHeader
	cli          client.HypeClient
)

func init() {
	flag.Parse()
	fmt.Fprint(os.Stderr, "Version: ", version)
	fmt.Println("")
	if *help {
		flag.PrintDefaults()
		os.Exit(0)
	}

	ogmiosUrl := strings.TrimSpace(*ogmiosIp)
	if !strings.HasPrefix(ogmiosUrl, "ws://") && !strings.HasPrefix(ogmiosUrl, "wss://") {
		ogmiosUrl = "ws://" + ogmiosUrl
	}

	if strings.HasPrefix(ogmiosUrl, "/") {
		ogmiosUrl = strings.TrimSuffix(ogmiosUrl, "/")
	}

	rpcUrl := strings.TrimSpace(*rpcIp)
	if strings.HasPrefix(rpcUrl, "/") {
		rpcUrl = strings.TrimSuffix(rpcUrl, "/")
	}

	hype.Init(ogmiosUrl, rpcUrl)
	processChain = make(chan *client.BlockHeader, 50)
}

func processBlock(row *client.BlockHeader) {
	if err := recover(); err != nil {
		logger.Errorf("processBlock recover %+v", err)
	}

	if row.NumTxs <= 0 {
		return
	}

	block, err := cli.GetBlock(uint64(row.Height))
	if err != nil {
		fmt.Println(row.Height, " err: ", err)
		return
	}

	for _, tx := range block.BlockDetails.Txs {
		//fmt.Println(tx.Action.Type, "======>", tx.TxHash, tx.BlockNumber)
		if tx.Action.Type != "SpotSend" && tx.Action.Type != "UsdSend" {
			continue
		}

		if tx.BlockNumber <= 0 {
			fmt.Println(tx.TxHash, " block number err")
			continue
		}
		// 交易入库
		fmt.Println("begin process transfer tx", tx.TxHash)
		err := db.AddTx(&db.Tx{
			BlockNo:          tx.BlockNumber,
			BlockTime:        tx.Time,
			BlockHash:        row.Hash,
			TxHash:           tx.TxHash,
			Amount:           tx.Action.Amount,
			AddressFrom:      tx.User,
			Destination:      tx.Action.Destination,
			Token:            tx.Action.Token,
			SignatureChainId: tx.Action.SignatureChainId,
			HyperLiquidChain: tx.Action.HyperliquidChain,
			ActionType:       tx.Action.Type,
			Error:            fmt.Sprintf("%+v", tx.Error),
		})

		if err != nil {
			fmt.Println(tx.TxHash, "insert into db error: ", err.Error())
		}
	}
}

func runSyncBlock() {
	cli = client.HypeClient(*(hype.GetHttpClient()))

	for i := 0; i < *thread; i++ {
		go func(id int) {
			for row := range processChain {
				fmt.Println("thread", i, "process block", row.Height, time.Now())
				processBlock(row)
			}
		}(i)
	}

	for {
		func() {
			defer func() {
				if err := recover(); err != nil {
					logger.Errorf("runSyncBlock recover %+v", err)
				}
			}()
			maxNum := int64(0)
			hype.SyncBlockHeader(func(block []*client.BlockHeader) bool {
				for i, row := range block {
					//fmt.Println(row.BlockNumber, row.TxHash)
					//fmt.Println(row.User, row.Time)
					if maxNum < row.Height {
						maxNum = row.Height
					} else {
						fmt.Println("sync error blockNum: ", row.Height)
					}
					processChain <- block[i]
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

func runSyncBlockTxs() {
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

					fmt.Println(row.Action.Type, "======>", row.TxHash, row.BlockNumber)
					if row.Action.Type != "SpotSend" && row.Action.Type != "UsdSend" {
						continue
					}

					if row.BlockNumber <= 0 {
						fmt.Println(row.TxHash, " block number err")
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
