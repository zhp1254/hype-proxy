package client

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"time"
)

func removeDuplicateElement(languages []string) []string {
	result := make([]string, 0, len(languages))
	temp := map[string]struct{}{}
	for _, item := range languages {
		if _, ok := temp[item]; !ok {
			temp[item] = struct{}{}
			result = append(result, item)
		}
	}
	return result
}

func getZdayIp() []string {
	url := "/free"
	cli := NewClient("https://www.zdaye.com", 10)
	code, body, err := cli.Request(http.MethodGet, url, "", nil)
	if err != nil {
		fmt.Println("getZdayIp: ", err)
		return nil
	}

	if code != http.StatusOK {
		fmt.Println("getZdayIp: httpcode", code)
		return nil
	}

	tbRegex, _ := regexp.Compile(`(?s)<table[^>]+id="ipc"[^>]+>(.+?)</table>`)
	tbs := tbRegex.FindStringSubmatch(string(body))
	if len(tbs) !=2 {
		return nil
	}

	trRegex, _ := regexp.Compile(`(?s)<tr[^>]*>.+?</tr>`)
	trs := trRegex.FindAllString(tbs[1], -1)

	portRegex, _ := regexp.Compile(`<td[^>]*>\s*(\d+)\s*</td>`)
	ipRegex, _ := regexp.Compile(`<td[^>]*>\s*(\d+\.\d+\.\d+\.\d+)\s*</td>`)

	ips := make([]string, 0)
	for _, tr := range trs {
		ip := ipRegex.FindStringSubmatch(tr)
		port := portRegex.FindStringSubmatch(tr)
		fmt.Println(ip)
		fmt.Println(port)
		if len(ip) == 2 && len(port) == 2 {
			ips = append(ips, fmt.Sprintf("http://%s:%s", ip[1], port[1]))
		}
	}
	return ips
}

func GetKuaidailiIp() []string {
	ips := make([]string, 0)

	for i := 1; i < 15; i++ {
		time.Sleep(time.Second * 2)
		url := fmt.Sprintf("/free/fps/%d/", i)
		cli := NewClient("https://www.kuaidaili.com", 10)
		code, body, err := cli.Request(http.MethodGet, url, "", nil)
		if err != nil {
			fmt.Println("getKuaidailiIp: ", err)
			continue
		}

		if code != http.StatusOK {
			fmt.Println("getKuaidailiIp: httpcode", code)
			continue
		}

		tbRegex, _ := regexp.Compile(`(?s)const\s*fpsList\s*=\s*\[\s*\{(.+?)\}\s*\]\s*;?`)
		tbs := tbRegex.FindStringSubmatch(string(body))
		//fmt.Println(tbs)
		if len(tbs) != 2 {
			continue
		}

		ipData := fmt.Sprintf("[{%s}]", tbs[1])
		//fmt.Println(ipData)
		var ipMap []map[string]interface{}
		err = json.Unmarshal([]byte(ipData), &ipMap)
		if err != nil {
			fmt.Println(err)
			continue
		}

		for _, row := range ipMap {
			if ip, ok := row["ip"].(string); ok {
				if port, ok := row["port"].(string); ok {
					//if valid, ok := row["is_valid"].(bool); ok && valid {
						ips = append(ips, fmt.Sprintf("http://%s:%s", ip, port))
					//}
				}
			}
		}
	}
 	return removeDuplicateElement(ips)
}
