// Package hnb makes HNB currency exchange rates available for consumption.
package hnb

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

// HNB manages data fetching from HNB.
type HNB struct {
	remote string
}

// New will create HNB manager.
func New() *HNB {
	return &HNB{
		remote: "http://www.hnb.hr/tecajn/htecajn.htm",
	}
}

// LatestExchange will return fresh exchange rates.
func (hnb *HNB) LatestExchange() (Exchange, error) {
	return fetch(hnb.remote)
}

// Exchange holds exchange rates for date of application.
type Exchange struct {
	Date  time.Time       `json:"date"`
	Rates map[string]Rate `json:"rates"`
}

// ServeHTTP makes Exchange available via HTTP.
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

// Rate holds different exchange rates for a currency.
type Rate struct {
	Buy    *big.Float `json:"buy"`
	Middle *big.Float `json:"middle"`
	Sell   *big.Float `json:"sell"`
}

// MarshalJSON satisfies json.Marshaler interface making
// rates have a fixed 6 decimal precision in JSON representation.
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

	return strings.Map(func(c rune) rune {
		switch c {
		case ',':
			return '.'
		case '.':
			return ','
		default:
			return c
		}
	}, in)
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
