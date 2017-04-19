// Your mission, should you choose to accept it,
// is to create a currency exchange rates JSON service.
//
// Mission is divided in following phases:
//
// Phase 1:
//     - Fetch the exchange rates.
//     - Convert the exchange rates to the internal
//       code representation, a structure.
//     - Make sure all rates are based on 1 unit
//       of currency.
//     - Print all currencies and their rates
//       on standard output.
//
// Phase 2:
//     - Make a http server.
//     - Create API endpoint that returns
//       exchange rates in a JSON format.
//
// Phase 3:
//     - Make a service resilient by storing
//       exchange rates to internal memory that
//       is updated every hour. If new exchange
//       can not be fetched, API remains to serve
//       existing data.
//
// Phase 4:
//     - Create API endpoint that given a value
//       in Kuna, Label of currency and a rate
//       type returns converted value to
//       requested currency.
//
// Phase 5:
//     - Measure your service response speed,
//       cpu and memory consumption.
//     - Find places in code that can be
//       optimized.
//
// Currency rates can be found on
// http://www.hnb.hr/tecajn/htecajn.htm
// in the following format:
//
// |----------------------------------+----------+--------+------------|
// | Field                            | Format   | Length | Field type |
// |----------------------------------+----------+--------+------------|
// | Header, first line               |          |        |            |
// |----------------------------------+----------+--------+------------|
// | Exchange number                  |          |      3 | number     |
// | Date of creation                 | ddmmyyyy |      8 | number     |
// | Date of application              | ddmmyyyy |      8 | number     |
// | Number of currencies that follow |          |      2 | number     |
// |----------------------------------+----------+--------+------------|
// | Currency records                 |          |        |            |
// |----------------------------------+----------+--------+------------|
// | Code                             |          |      3 | number     |
// | Label                            |          |      3 | alpha      |
// | Number of units                  |          |      3 | number     |
// | Buy rate                         |          |    8,6 | number     |
// | Middle rate                      |          |    8,6 | number     |
// | Sell rate                        |          |    8,6 | number     |
// |----------------------------------+----------+--------+------------|
//
// To help you in your mission execute the following command line
// godoc -http :9001 &
// then load http://127.0.0.1:9001/ in you browser.
//
// For Phase 1 you will need to get familiar with following methods:
// |--------------------+-----------------------------------|
// | Method             | Usage                             |
// |--------------------+-----------------------------------|
// | http.Client.Get    | making HTTP requests              |
// | bufio.Scanner.Scan | traversing text                   |
// | time.Parse         | parsing date formats              |
// | strings.Fields     | splitting string                  |
// | strconv.Atoi       | converting string to int          |
// | strings.Map        | converting characters in a string |
// | big.Float.Parse    | parsing string to big.Float       |
// | big.Flaot.SetInt64 | storing int64 to big.Float        |
// | big.Flaot.Quo      | dividing big.Float                |
// | fmt.Printf         | output                            |
// |--------------------+-----------------------------------|
//
// Good luck Gopher!

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"net/http"
	_ "net/http/pprof"
	"time"

	"github.com/MatejB/go-workshop-currency/hnb"
)

func main() {
	hnbRates := hnb.New()

	http.HandleFunc("/", ratesHandler(hnbRates))
	http.HandleFunc("/convert", conversionHandler(hnbRates))

	s := &http.Server{
		Addr:           ":5555",
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		IdleTimeout:    10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	err := s.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}

type exchanger interface {
	LatestExchange() (hnb.Exchange, error)
}

func ratesHandler(exchanger exchanger) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		exch, err := exchanger.LatestExchange()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		exch.ServeHTTP(w, req)
	}
}

type conversionRequest struct {
	Value    float64 `json:"value"`
	Currency string  `json:"currency"`
	Rate     string  `json:"rate"`
}

type conversionResponse struct {
	Result float64 `json:"result"`
}

func conversionHandler(exchanger exchanger) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		defer req.Body.Close()

		dec := json.NewDecoder(req.Body)

		var reqData conversionRequest
		err := dec.Decode(&reqData)
		if err != nil {
			http.Error(w, fmt.Sprintf("Request decoding: %s", err), http.StatusBadRequest)
			return
		}

		exch, err := exchanger.LatestExchange()
		if err != nil {
			http.Error(w, fmt.Sprintf("Fetching exchange: %s", err), http.StatusInternalServerError)
			return
		}

		rates, ok := exch.Rates[reqData.Currency]
		if !ok {
			http.Error(w, "Unknown currency.", http.StatusBadRequest)
			return
		}
		var multipler *big.Float
		switch reqData.Rate {
		case "buy":
			multipler = rates.Buy
		case "middle":
			multipler = rates.Middle
		case "sell":
			multipler = rates.Sell
		default:
			http.Error(w, "Unknown rate value.", http.StatusBadRequest)
			return
		}

		result := big.NewFloat(reqData.Value)
		result = result.Mul(result, multipler)

		resFloat, _ := result.Float64()
		resData := conversionResponse{
			Result: resFloat,
		}

		out, err := json.Marshal(resData)
		if err != nil {
			http.Error(w, fmt.Sprintf("Response encoding: %s", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		fmt.Fprintf(w, "%s", out)
	}
}
