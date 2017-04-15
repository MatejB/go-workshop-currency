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
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		exch, err := fetch("http://www.hnb.hr/tecajn/htecajn.htm")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		exch.ServeHTTP(w, req)
	})

	s := &http.Server{
		Addr:           ":5555",
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		IdleTimeout:    10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	log.Fatal(s.ListenAndServe())
}

type Exchange struct {
	Date  time.Time       `json:"date"`
	Rates map[string]Rate `json:"rates"`
}

func (e *Exchange) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	data, err := json.Marshal(e)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_, err = fmt.Fprintf(w, "%s", data)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

type Rate struct {
	Buy    *big.Float `json:"buy"`
	Middle *big.Float `json:"middle"`
	Sell   *big.Float `json:"sell"`
}

func (rt Rate) MarshalJSON() ([]byte, error) {
	fixedPrecision := struct {
		Buy    string `json:"buy"`
		Middle string `json:"middle"`
		Sell   string `json:"sell"`
	}{
		Buy:    fmt.Sprintf("%.6f", rt.Buy),
		Middle: fmt.Sprintf("%.6f", rt.Middle),
		Sell:   fmt.Sprintf("%.6f", rt.Sell),
	}

	return json.Marshal(fixedPrecision)
}

func fetch(source string) (exchange Exchange, err error) {
	exchange.Rates = make(map[string]Rate, 0)

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Get(source)
	if err != nil {
		return exchange, fmt.Errorf("Error in fetching data from %q: %s", source, err)
	}
	defer resp.Body.Close()

	scanner := bufio.NewScanner(resp.Body)

	for scanner.Scan() {
		rate := Rate{}

		line := scanner.Text()
		if len(line) == 0 {
			continue
		}

		if len(line) == 21 {
			exchange.Date, err = time.Parse("02012006", line[11:19])
			if err != nil {
				return exchange, fmt.Errorf("Error in parsing date from %q: %s", line, err)
			}
			continue
		}

		parts := strings.Fields(line)
		if len(parts) != 4 {
			return exchange, fmt.Errorf("Unknown exchange format %q", line)
		}

		currency := parts[0][3:6]
		units, err := strconv.Atoi(parts[0][6:])
		if err != nil {
			return exchange, fmt.Errorf("Error in parsing units from %q: %s", line, err)
		}

		rate.Buy, err = normaliseRate(engFloat(parts[1]), units)
		if err != nil {
			return exchange, fmt.Errorf("Error while normalizing rate %q: %s", line, err)
		}

		rate.Middle, err = normaliseRate(engFloat(parts[2]), units)
		if err != nil {
			return exchange, fmt.Errorf("Error while normalizing rate %q: %s", line, err)
		}

		rate.Sell, err = normaliseRate(engFloat(parts[3]), units)
		if err != nil {
			return exchange, fmt.Errorf("Error while normalizing rate %q: %s", line, err)
		}

		exchange.Rates[currency] = rate
	}

	if err := scanner.Err(); err != nil {
		return exchange, fmt.Errorf("Error in scanning of response: %s", err)
	}

	return
}

func engFloat(in string) (out string) {
	// out = strings.Replace(in, ".", ",", -1)
	// out = strings.Replace(out, ",", ".", -1)

	return strings.Map(
		func(c rune) rune {
			switch c {
			case ',':
				return '.'
			case '.':
				return ','
			default:
				return c
			}
		},
		in,
	)
}

func normaliseRate(value string, units int) (*big.Float, error) {
	number := new(big.Float)
	number, _, err := number.Parse(value, 10)
	if err != nil {
		return number, err
	}

	if units != 1 {
		divisor := new(big.Float)
		divisor.SetInt64(int64(units))

		number = number.Quo(number, divisor)
	}

	return number, nil
}
