package main

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

var (
	ErrSearchFailed = errors.New("search failed")
)

func main() {
	fmt.Println("Bilibili Anime App Search")
	fmt.Println("---------------------")
	for {
		fmt.Print("Keyword -> ")
		text, err := ReadInput()
		if err != nil {
			fmt.Printf("Search failed: %s", err.Error())
		}

		if text == "" {
			os.Exit(0)
		}

		results, err := Search(text)
		if err != nil {
			fmt.Printf("Search failed: %s", err.Error())
		}

		fmt.Println("Search Result:")
		for _, item := range results {
			fmt.Printf("(%d) %s\n", item.SeasonID, item.Title)
		}
		fmt.Println("----------------")
	}

}

func ReadInput() (string, error) {
	reader := bufio.NewReader(os.Stdin)
	text, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	text = strings.Replace(text, "\n", "", -1)
	text = strings.Replace(text, "\r", "", -1)
	return text, nil
}

type restGetSearchResultResponse struct {
	Code    int64          `json:"code"`
	Message string         `json:"message"`
	TTL     int64          `json:"ttl"`
	Data    restSearchData `json:"data"`
}

type restSearchData struct {
	Pages int64            `json:"pages"`
	Total int64            `json:"total"`
	Items []restSearchItem `json:"items"`
}

type restSearchItem struct {
	SeasonID int64  `json:"season_id"`
	Title    string `json:"title"`
	Cover    string `json:"cover"`
}

func Search(keyword string) ([]restSearchItem, error) {
	ctx := context.TODO()

	c := &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:        10,
			IdleConnTimeout:     30 * time.Second,
			DisableCompression:  true,
			MaxIdleConnsPerHost: 10,
		},
	}

	endpoint := fmt.Sprintf(
		"https://app.biliintl.com/intl/gateway/v2/app/search/type?platform=app&s_locale=en_US&keyword=%s&highlight=0&type=7",
		url.QueryEscape(keyword),
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, ErrSearchFailed
	}
	res, err := c.Do(req)
	if err != nil {
		return nil, ErrSearchFailed
	}

	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, ErrSearchFailed
	}

	defer res.Body.Close()

	if res.StatusCode != 200 {
		return nil, ErrSearchFailed
	}

	var o restGetSearchResultResponse
	err = json.Unmarshal(data, &o)
	if err != nil {
		return nil, ErrSearchFailed
	}

	return o.Data.Items, nil
}
