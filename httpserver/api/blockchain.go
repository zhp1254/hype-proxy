package api

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"hype-proxy/db"
	"hype-proxy/hype"
	"hype-proxy/hype/client"
	"io"
	"io/ioutil"
	"strconv"
	"strings"
	"time"
)

func GetBestBlock(c *gin.Context) {
	c.JSON(200, db.GetBestHeight())
}

func GetBlock(c *gin.Context) {
	blockNoStr, _ := c.GetQuery("blockNo")
	blockNo, _ := strconv.ParseInt(strings.TrimSpace(blockNoStr), 10, 64)
	if blockNo <= 0 {
		c.String(500, "params blockNo invalid")
		return
	}

	rows, err := db.FindByBlockNum(blockNo)
	if err != nil {
		c.String(500, err.Error())
		return
	}

	ret := &client.Block{
		BlockDetails: client.BlockDetails{
			Height:    blockNo,
			BlockTime: time.Now().UnixMilli(),
			Hash:      blockNoStr,
			Proposer:  blockNoStr,
			NumTxs:    len(rows),
			Txs:       make([]client.Transfer, 0),
		},
	}

	if rows == nil {
		c.JSON(200, ret)
		return
	}

	blockTime := rows[0].BlockTime
	ret.BlockDetails.BlockTime = blockTime

	for _, row := range rows {
		ret.BlockDetails.Txs = append(ret.BlockDetails.Txs, client.Transfer{
			Time:        row.BlockTime,
			User:        row.AddressFrom,
			BlockNumber: row.BlockNo,
			TxHash:      row.TxHash,
			Error:       row.Error,
			Action: client.Action{
				Type:             row.ActionType,
				SignatureChainId: row.SignatureChainId,
				HyperliquidChain: row.HyperLiquidChain,
				Destination:      row.Destination,
				Token:            row.Token,
				Amount:           row.Amount.String(),
				Time:             row.BlockTime,
			},
		})
	}

	c.JSON(200, ret)
	return
}

func ProxyHypeExchange(c *gin.Context) {
	body, err := ioutil.ReadAll(c.Request.Body)
	if err != nil && err != io.EOF {
		c.String(500, err.Error())
		return
	}

	if len(body) == 0 {
		c.String(500, "request body empty")
		return
	}

	fmt.Println(string(body))
	httpCode, body, err := hype.GetHttpClient().Request(client.POST, "/exchange", string(body), client.JsonHeader())
	if err != nil {
		c.String(httpCode, err.Error())
		return
	}
	c.String(httpCode, string(body))
}

func ProxyHypeExplorer(c *gin.Context) {
	body, err := ioutil.ReadAll(c.Request.Body)
	if err != nil && err != io.EOF {
		c.String(500, err.Error())
		return
	}

	if len(body) == 0 {
		c.String(500, "request body empty")
		return
	}

	fmt.Println(string(body))
	httpCode, body, err := hype.GetHttpClient().Request(client.POST, "/explorer", string(body), client.JsonHeader())
	if err != nil {
		c.String(httpCode, err.Error())
		return
	}
	c.String(httpCode, string(body))
}

func ProxyHypeInfo(c *gin.Context) {
	body, err := ioutil.ReadAll(c.Request.Body)
	if err != nil && err != io.EOF {
		c.String(500, err.Error())
		return
	}

	if len(body) == 0 {
		c.String(500, "request body empty")
		return
	}

	fmt.Println(string(body))
	httpCode, body, err := hype.GetHttpClient().Request(client.POST, "/info", string(body), client.JsonHeader())
	if err != nil {
		c.String(httpCode, err.Error())
		return
	}
	c.String(httpCode, string(body))
}
