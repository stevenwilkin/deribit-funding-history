package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"
)

type fundingRate struct {
	Interest  float64 `json:"interest_1h"`
	Timestamp int64   `json:"timestamp"`
}

type fundingRatesResponse struct {
	Result []fundingRate `json:"result"`
}

func getPage(startTime, endTime int64) ([]fundingRate, error) {
	v := url.Values{
		"instrument_name": {"BTC-PERPETUAL"},
		"start_timestamp": {fmt.Sprintf("%d", startTime)},
		"end_timestamp":   {fmt.Sprintf("%d", endTime)}}

	u := url.URL{
		Scheme:   "https",
		Host:     "www.deribit.com",
		Path:     "/api/v2/public/get_funding_rate_history",
		RawQuery: v.Encode()}

	resp, err := http.Get(u.String())
	if err != nil {
		return []fundingRate{}, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return []fundingRate{}, err
	}

	var result fundingRatesResponse
	json.Unmarshal(body, &result)

	return result.Result, nil
}

func getFundingRates(initialTime int64) ([]float64, error) {
	var (
		results      []float64
		fundingRates []fundingRate
		err          error
	)

	startTime := initialTime
	endTime := time.Now().UnixNano() / int64(time.Millisecond)

	for {
		fundingRates, err = getPage(startTime, endTime)
		if err != nil {
			return []float64{}, err
		}

		for _, item := range fundingRates {
			results = append(results, item.Interest)
		}

		if len(fundingRates) == 744 {
			// results are in chronological order but each page is bounded by
			// the end time
			endTime = fundingRates[0].Timestamp - 1
		} else {
			break
		}
	}

	return results, err
}

func startTime() int64 {
	var days int

	if len(os.Args) == 2 {
		days, _ = strconv.Atoi(os.Args[1])
	}

	if days == 0 {
		days = 30
	}

	ts := time.Now().AddDate(0, 0, days*-1)

	return ts.UnixNano() / int64(time.Millisecond)
}

func main() {
	fundingRates, err := getFundingRates(startTime())
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	var total float64
	days := float64(len(fundingRates)) / 24

	for _, funding := range fundingRates {
		total += funding
	}

	annualised := total / (days / 364)

	fmt.Printf("Days:  %.1f\n", days)
	fmt.Printf("Total: %.2f%%\n", total*100)
	fmt.Printf("APR:   %.2f%%\n", annualised*100)
}
