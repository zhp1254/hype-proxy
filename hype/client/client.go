package client

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type HypeClient CustomClient

const SUCCESS = "ok"

func (c *HypeClient) Check() bool {
	httpCode, _, err := CustomClient(*c).Request(GET, "/explorer", "", nil)
	if err != nil {
		return false
	}
	return httpCode == 405
}

func (c *HypeClient) GetTransactionByTxid(txid string) (*TxDetails, error) {
	if len(txid) == 0 {
		return nil, errors.New("txid is empty")
	}

	data := fmt.Sprintf(`{"hash":"%s","type":"txDetails"}`, txid)
	path := "/explorer"
	httpCode, body, err := CustomClient(*c).Request(POST, path, data, JsonHeader())
	if err != nil {
		return nil, err
	}

	if httpCode != http.StatusOK && httpCode != http.StatusAccepted {
		return nil, fmt.Errorf("http status error:%+v, %+v", httpCode, string(body))
	}

	if len(body) == 0 {
		return nil, errors.New("got body nil")
	}

	fmt.Println(string(body))
	var ret TxDetails
	err = json.Unmarshal(body, &ret)
	if err != nil {
		return nil, err
	}

	if ret.Tx.Error != nil {
		return nil, errors.New(string(body))
	}

	return &ret, nil
}

// GetAddressAsset https://hyperliquid.gitbook.io/hyperliquid-docs/for-developers/api/info-endpoint/spot#retrieve-a-users-token-balances
func (c *HypeClient) GetAddressSpotAsset(address string) (*Asset, error) {
	data := fmt.Sprintf(`{"user":"%s","type":"spotClearinghouseState"}`, address)
	path := "/info"
	httpCode, body, err := CustomClient(*c).Request(POST, path, data, JsonHeader())
	if err != nil {
		return nil, err
	}

	if httpCode != http.StatusOK && httpCode != http.StatusAccepted {
		return nil, fmt.Errorf("http status error:%+v, %+v", httpCode, string(body))
	}

	if len(body) == 0 {
		return nil, errors.New("got body nil")
	}

	var ret Asset
	err = json.Unmarshal(body, &ret)
	return &ret, err
}

func (c *HypeClient) GetAddressPerpetualsAsset(address string) (*Perpetuals, error) {
	data := fmt.Sprintf(`{"user":"%s","type":"clearinghouseState"}`, address)
	path := "/info"
	httpCode, body, err := CustomClient(*c).Request(POST, path, data, JsonHeader())
	if err != nil {
		return nil, err
	}

	if httpCode != http.StatusOK && httpCode != http.StatusAccepted {
		return nil, fmt.Errorf("http status error:%+v, %+v", httpCode, string(body))
	}

	if len(body) == 0 {
		return nil, errors.New("got body nil")
	}

	// fmt.Println(string(body))
	var ret Perpetuals
	err = json.Unmarshal(body, &ret)
	return &ret, err
}

func (c *HypeClient) GetTokenInfo() (*TokenInfo, error) {
	data := `{"type":"spotMeta"}`
	path := "/info"
	httpCode, body, err := CustomClient(*c).Request(POST, path, data, JsonHeader())
	if err != nil {
		return nil, err
	}

	if httpCode != http.StatusOK && httpCode != http.StatusAccepted {
		return nil, fmt.Errorf("http status error:%+v, %+v", httpCode, string(body))
	}

	if len(body) == 0 {
		return nil, errors.New("got body nil")
	}

	var ret TokenInfo
	err = json.Unmarshal(body, &ret)
	return &ret, err
}

func (c *HypeClient) GetLastBlock() (uint64, error) {
	return 0, nil
}

func (c *HypeClient) GetBlock(number uint64) (*Block, error) {
	if number <= 0 {
		return nil, errors.New("block number must > 0")
	}

	data := fmt.Sprintf(`{"height":%d,"type":"blockDetails"}`, number)
	path := "/explorer"
	httpCode, body, err := CustomClient(*c).Request(POST, path, data, JsonHeader())
	if err != nil {
		return nil, err
	}

	if httpCode != http.StatusOK && httpCode != http.StatusAccepted {
		if httpCode == 429 {
			time.Sleep(3 * time.Second)
		}
		return nil, fmt.Errorf("http status error:%+v, %+v", httpCode, string(body))
	}

	if len(body) == 0 {
		return nil, errors.New("got body nil")
	}

	var ret Block
	err = json.Unmarshal(body, &ret)
	return &ret, err
}

func (c *HypeClient) SendRawTransaction(raw string) error {
	path := "/exchange"
	httpCode, body, err := CustomClient(*c).Request(POST, path, raw, JsonHeader())
	if err != nil {
		return err
	}

	if httpCode != http.StatusOK && httpCode != http.StatusAccepted {
		return fmt.Errorf("http status error:%+v", httpCode)
	}

	if len(body) == 0 {
		return errors.New("got http body nil")
	}

	var ret map[string]interface{}
	err = json.Unmarshal(body, &ret)
	if err != nil {
		return err
	}

	status, ok := ret["status"].(string)
	if !ok || !strings.EqualFold(SUCCESS, status) {
		response, _ := ret["response"].(string)
		return errors.New(response)
	}
	return nil
}
