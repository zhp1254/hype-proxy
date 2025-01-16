package client

import (
	"crypto/tls"
	"fmt"
	"hype-proxy/logger"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type CustomClient struct {
	Domain string
	client *http.Client
}

const (
	POST = "POST"
	GET  = "GET"
)

var (
	client CustomClient
)

func init() {
	client = NewClient("", 360)
}

func NewClient(domain string, timeout int) CustomClient {
	client := http.DefaultClient
	client.Timeout = time.Second * time.Duration(timeout)
	return CustomClient{
		Domain: domain,
		client: client,
	}
}

func NewProxyClient(domain, proxyUrl string, timeout int) *HypeClient {
	client := http.DefaultClient
	if len(proxyUrl) > 0 {
		proxy, _ := url.Parse(proxyUrl)
		tr := &http.Transport{
			Proxy:           http.ProxyURL(proxy),
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}

		client = &http.Client{
			Transport: tr,
		}
	}

	client.Timeout = time.Second * time.Duration(timeout)
	cli := HypeClient(CustomClient{
		Domain: domain,
		client: client,
	})
	return &cli
}

func JsonHeader() map[string]string {
	ret := make(map[string]string, 2)
	ret["Content-Type"] = "application/json"
	ret["accept"] = "application/json"
	return ret
}

func (c CustomClient) Request(
	reqType string,
	path string,
	postData string,
	requstHeaders map[string]string) (int, []byte, error) {
	reqUrl := fmt.Sprintf("%s%s", c.Domain, path)
	req, err := http.NewRequest(reqType, reqUrl, strings.NewReader(postData))
	if err != nil {
		logger.Errorf("http.NewRequest err:%+v", err)
		return 0, nil, err
	}

	req.Close = true
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 5.1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/31.0.1650.63 Safari/537.36")
	if requstHeaders != nil {
		for k, v := range requstHeaders {
			req.Header.Add(k, v)
		}
	} else {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}

	resp, err := c.client.Do(req)
	if err != nil {
		//logger.Errorf("client.Do err:%+v", err)
		return 0, nil, err
	}

	defer resp.Body.Close()
	bodyData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.Errorf("ioutil.ReadAll(resp.Body) err:%+v", err)
		return resp.StatusCode, nil, err
	}

	if resp.StatusCode != 200 {
		if tr, ok := c.client.Transport.(*http.Transport); ok && tr.Proxy != nil {
			pUrl, _ := tr.Proxy(nil)
			logger.Infof("resp.StatusCode is:%+v, res.Body is:%+v, proxy:%+v",
				resp.StatusCode, string(bodyData), pUrl)
		} else {
			logger.Infof("resp.StatusCode is:%+v, res.Body is:%+v",
				resp.StatusCode, string(bodyData))
		}
	}
	return resp.StatusCode, bodyData, nil
}
