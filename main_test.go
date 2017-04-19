package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/big"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"
)

func TestFetch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`
059240320172503201713
036AUD001       5,207988       5,223659       5,239330
124CAD001       5,119576       5,134981       5,150386
203CZK001       0,273458       0,274281       0,275104
208DKK001       0,993382       0,996371       0,999360
348HUF100       2,388187       2,395373       2,402559
392JPY100       6,161252       6,179791       6,198330
578NOK001       0,806093       0,808519       0,810945
752SEK001       0,775638       0,777972       0,780306
756CHF001       6,900694       6,921458       6,942222
826GBP001       8,539728       8,565424       8,591120
840USD001       6,839371       6,859951       6,880531
978EUR001       7,388573       7,410805       7,433037
985PLN001       1,730993       1,736202       1,741411
`))
	}))
	defer server.Close()

	expDate, err := time.Parse("02.01.2006.", "25.03.2017.")
	if err != nil {
		t.Errorf("Unexpected error %q.", err)
	}

	expected := Exchange{
		Date: expDate,
		Rates: map[string]Rate{
			"EUR": Rate{Buy: big.NewFloat(7.388573), Middle: big.NewFloat(7.410805), Sell: big.NewFloat(7.433037)},
			"DKK": Rate{Buy: big.NewFloat(0.993382), Middle: big.NewFloat(0.996371), Sell: big.NewFloat(0.99936)},
			"NOK": Rate{Buy: big.NewFloat(0.806093), Middle: big.NewFloat(0.808519), Sell: big.NewFloat(0.810945)},
			"SEK": Rate{Buy: big.NewFloat(0.775638), Middle: big.NewFloat(0.777972), Sell: big.NewFloat(0.780306)},
			"CHF": Rate{Buy: big.NewFloat(6.900694), Middle: big.NewFloat(6.921458), Sell: big.NewFloat(6.942222)},
			"GBP": Rate{Buy: big.NewFloat(8.539728), Middle: big.NewFloat(8.565424), Sell: big.NewFloat(8.59112)},
			"USD": Rate{Buy: big.NewFloat(6.839371), Middle: big.NewFloat(6.859951), Sell: big.NewFloat(6.880531)},
			"PLN": Rate{Buy: big.NewFloat(1.730993), Middle: big.NewFloat(1.736202), Sell: big.NewFloat(1.741411)},
			"AUD": Rate{Buy: big.NewFloat(5.207988), Middle: big.NewFloat(5.223659), Sell: big.NewFloat(5.23933)},
			"CAD": Rate{Buy: big.NewFloat(5.119576), Middle: big.NewFloat(5.134981), Sell: big.NewFloat(5.150386)},
			"CZK": Rate{Buy: big.NewFloat(0.273458), Middle: big.NewFloat(0.274281), Sell: big.NewFloat(0.275104)},
			"HUF": Rate{Buy: big.NewFloat(0.02388187), Middle: big.NewFloat(0.02395373), Sell: big.NewFloat(0.02402559)},
			"JPY": Rate{Buy: big.NewFloat(0.06161252), Middle: big.NewFloat(0.06179791), Sell: big.NewFloat(0.0619833)},
		},
	}

	recived, err := fetch(server.URL)
	if err != nil {
		t.Errorf("Unexpected error %q.", err)
	}

	for currency, expRate := range expected.Rates {
		recRate, ok := recived.Rates[currency]
		if !ok {
			t.Errorf("Expected currency %q.", currency)
			continue
		}

		if fmt.Sprintf("%.9f", recRate.Buy) != fmt.Sprintf("%.9f", expRate.Buy) {
			t.Errorf("Expected %v got %v on currency %s.", expRate.Buy, recRate.Buy, currency)
		}
		if fmt.Sprintf("%.9f", recRate.Middle) != fmt.Sprintf("%.9f", expRate.Middle) {
			t.Errorf("Expected %v got %v on currency %s.", expRate.Middle, recRate.Middle, currency)
		}
		if fmt.Sprintf("%.9f", recRate.Sell) != fmt.Sprintf("%.9f", expRate.Sell) {
			t.Errorf("Expected %v got %v on currency %s.", expRate.Sell, recRate.Sell, currency)
		}

	}
}

func TestServe(t *testing.T) {
	req := httptest.NewRequest("GET", "http://address-of-our-service/", nil)
	w := httptest.NewRecorder()

	expDate, err := time.Parse("02.01.2006.", "25.03.2017.")
	if err != nil {
		t.Errorf("Unexpected error %q.", err)
	}

	expected := Exchange{
		Date: expDate,
		Rates: map[string]Rate{
			"EUR": Rate{Buy: big.NewFloat(7.388573), Middle: big.NewFloat(7.410805), Sell: big.NewFloat(7.433037)},
			"DKK": Rate{Buy: big.NewFloat(0.993382), Middle: big.NewFloat(0.996371), Sell: big.NewFloat(0.99936)},
			"NOK": Rate{Buy: big.NewFloat(0.806093), Middle: big.NewFloat(0.808519), Sell: big.NewFloat(0.810945)},
			"SEK": Rate{Buy: big.NewFloat(0.775638), Middle: big.NewFloat(0.777972), Sell: big.NewFloat(0.780306)},
			"CHF": Rate{Buy: big.NewFloat(6.900694), Middle: big.NewFloat(6.921458), Sell: big.NewFloat(6.942222)},
			"GBP": Rate{Buy: big.NewFloat(8.539728), Middle: big.NewFloat(8.565424), Sell: big.NewFloat(8.59112)},
			"USD": Rate{Buy: big.NewFloat(6.839371), Middle: big.NewFloat(6.859951), Sell: big.NewFloat(6.880531)},
			"PLN": Rate{Buy: big.NewFloat(1.730993), Middle: big.NewFloat(1.736202), Sell: big.NewFloat(1.741411)},
			"AUD": Rate{Buy: big.NewFloat(5.207988), Middle: big.NewFloat(5.223659), Sell: big.NewFloat(5.23933)},
			"CAD": Rate{Buy: big.NewFloat(5.119576), Middle: big.NewFloat(5.134981), Sell: big.NewFloat(5.150386)},
			"CZK": Rate{Buy: big.NewFloat(0.273458), Middle: big.NewFloat(0.274281), Sell: big.NewFloat(0.275104)},
			"HUF": Rate{Buy: big.NewFloat(0.02388187), Middle: big.NewFloat(0.02395373), Sell: big.NewFloat(0.02402559)},
			"JPY": Rate{Buy: big.NewFloat(0.06161252), Middle: big.NewFloat(0.06179791), Sell: big.NewFloat(0.0619833)},
		},
	}

	expected.ServeHTTP(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Errorf("Expected %d got %d.", 200, resp.StatusCode)
	}

	if resp.Header.Get("Content-Type") != "application/json; charset=utf-8" {
		t.Errorf("Expected %q got %q.", "application/json; charset=utf-8", resp.Header.Get("Content-Type"))
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Errorf("Unexpected error %q.", err)
	}

	if len(body) == 0 {
		t.Fatal("Unexpected response body to be empty")
	}

	var recived Exchange

	err = json.Unmarshal(body, &recived)
	if err != nil {
		t.Errorf("Unexpected error %q.", err)
	}

	if !reflect.DeepEqual(recived, expected) {
		t.Errorf("Expected %#v got %#v.", expected, recived)
	}
}
