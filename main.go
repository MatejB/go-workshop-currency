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
	"math/big"
	"strings"
	"time"
)

func main() {

}

type Exchange struct {
	Date  time.Time
	Rates map[string]Rate
}

type Rate struct {
	Buy    *big.Float
	Middle *big.Float
	Sell   *big.Float
}

func fetch(source string) (exchange Exchange, err error) {
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
