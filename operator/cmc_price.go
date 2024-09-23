package operator

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
)

func getCMCPrice(symbol string, convert string) interface{} {
	config, err := loadConfig("./config.json")
	if err != nil {
		log.Print(err)
		os.Exit(1)
	}
	client := &http.Client{}
	req, err := http.NewRequest("GET", "https://pro-api.coinmarketcap.com/v2/cryptocurrency/quotes/latest", nil)
	if err != nil {
		log.Print(err)
		os.Exit(1)
	}

	q := url.Values{}
	q.Add("symbol", symbol)
	q.Add("convert_id", convert)

	req.Header.Set("Accepts", "application/json")
	req.Header.Add("X-CMC_PRO_API_KEY", config.CmcApi)
	req.URL.RawQuery = q.Encode()

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending request to server")
		os.Exit(1)
	}
	fmt.Println(resp.Status)
	respBody, _ := ioutil.ReadAll(resp.Body)

	var res map[string]interface{}
	err = json.Unmarshal(respBody, &res)
	if err != nil {
		log.Print(err)
		os.Exit(1)
	}

	data := res["data"].(map[string]interface{})
	btcData := data[symbol].([]interface{})
	fmt.Println("Price of Bitcoin:", btcData[0].(map[string]interface{})["quote"].(map[string]interface{})["825"].(map[string]interface{})["price"])
	return btcData[0].(map[string]interface{})["quote"].(map[string]interface{})["825"].(map[string]interface{})["price"]
}
