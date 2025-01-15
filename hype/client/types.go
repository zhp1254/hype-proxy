package client

import (
	"github.com/shopspring/decimal"
)

type Action struct {
	Type             string          `json:"type"`
	SignatureChainId string          `json:"signatureChainId"`
	HyperliquidChain string          `json:"hyperliquidChain"`
	Destination      string          `json:"destination"`
	Token            string          `json:"token"`
	Amount           decimal.Decimal `json:"amount"`
	Time             int64           `json:"time"`
}

type Transfer struct {
	Time        int64       `json:"time"`
	User        string      `json:"user"`
	Action      Action      `json:"action"`
	BlockNumber int64       `json:"block"`
	TxHash      string      `json:"hash"`
	Error       interface{} `json:"error"`
}

type TxDetails struct {
	Type string   `json:"type"`
	Tx   Transfer `json:"tx"`
}

type BlockDetails struct {
	Height    int64      `json:"height"`
	BlockTime int64      `json:"blockTime"`
	Hash      string     `json:"hash"`
	Proposer  string     `json:"proposer"`
	NumTxs    int        `json:"numTxs"`
	Txs       []Transfer `json:"txs"`
}

type Block struct {
	Type         string       `json:"type"`
	BlockDetails BlockDetails `json:"blockDetails"`
}

type Asset struct {
	Balances []struct {
		Coin     string          `json:"coin"`
		Token    int64           `json:"token"`
		Hold     decimal.Decimal `json:"hold"`
		Total    decimal.Decimal `json:"total"`
		EntryNtl decimal.Decimal `json:"entryNtl"`
	} `json:"balances"`
}

type TokenInfo struct {
	Tokens []struct {
		Name        string `json:"name"`
		SzDecimals  uint64 `json:"szDecimals"`
		WeiDecimals uint64 `json:"weiDecimals"`
		Index       int64  `json:"index"`
		TokenId     string `json:"tokenId"`
		IsCanonical bool   `json:"isCanonical"`
		EvmContract string `json:"evmContract"`
		FullName    string `json:"fullName"`
	}
}

type Perpetuals struct {
	MarginSummary              interface{}     `json:"marginSummary"`
	CrossMarginSummary         interface{}     `json:"crossMarginSummary"`
	CrossMaintenanceMarginUsed string          `json:"crossMaintenanceMarginUsed"`
	Withdrawable               decimal.Decimal `json:"withdrawable"`
	AssetPositions             []interface{}   `json:"assetPositions"`
	Time                       int64           `json:"time"`
}

type BlockHeader struct {
	Height    int64  `json:"height"`
	BlockTime int64  `json:"blockTime"`
	Hash      string `json:"hash"`
	Proposer  string `json:"proposer"`
	NumTxs    int64  `json:"numTxs"`
}
