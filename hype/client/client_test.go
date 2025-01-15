package client

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestGetTx(t *testing.T) {
	txid := "0xac00cc86004d2d7325b50419e6f2ca018200f85e736f625243e1bea81123620c"
	cli := HypeClient(NewClient("https://api-ui.hyperliquid.xyz", 60))
	fmt.Println(cli.GetTransactionByTxid(txid))
}

func TestGetBlock(t *testing.T) {
	cli := HypeClient(NewClient("https://api-ui.hyperliquid.xyz", 60))
	info, err := cli.GetBlock(434565834)
	fmt.Println(err, info.Type, info.BlockDetails.Height, info.BlockDetails.Hash)
	fmt.Println(info.BlockDetails.NumTxs, len(info.BlockDetails.Txs))
	for _, row := range info.BlockDetails.Txs {
		if row.Action.Type == "spotSend" {
			fmt.Println(row.TxHash, row.User, row.Action.Token, row.Action.Amount, row.Action.Destination)
		}
	}
}

func TestGetAsset(t *testing.T) {
	cli := HypeClient(NewClient("https://api-ui.hyperliquid.xyz", 60))
	fmt.Println(cli.GetAddressSpotAsset("0x633a84ee0ab29d911e5466e5e1cb9cdbf5917e72"))
}

func TestGetToken(t *testing.T) {
	cli := HypeClient(NewClient("https://api-ui.hyperliquid.xyz", 60))
	fmt.Println(cli.GetTokenInfo())
}

func TestSendRawTx(t *testing.T) {
	tx := `{"signature":{"r":"7823058545ae78a89529be1aeeddaca0a624acd736d958f5aff4e1532f98520f","s":"9f3caf8d1d6f4745e5351aba739b34363582353f72e56c5cc9600550573bff6c","v":1},"nonce":1735643685280,"action":{"type":"usdSend","hyperliquidChain":"Testnet","signatureChainId":"0x66eee","destination":"0x26f5e66C919c67102624C8781720663933dB0e7F","amount":"1","time":1735643685280}}`
	cli := HypeClient(NewClient("https://api-ui.hyperliquid.xyz", 60))
	fmt.Println(cli.SendRawTransaction(tx))
}

func TestBlockHeader(t *testing.T)  {
	data := `[{"height":449530049,"blockTime":1736162835569,"hash":"0xadaf0dc18f3ff1f2d6b7a96225ff053790acc5b4d3006be136dc091146ed2970","proposer":"0x5795ab6e71ecbefa255fc4728cc34893ba992d44","numTxs":334}]`
	var header []*BlockHeader
	fmt.Println(json.Unmarshal([]byte(data), & header))
}