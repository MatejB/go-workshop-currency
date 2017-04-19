package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/MatejB/go-workshop-currency/hnb"
)

func TestConversionHandler(t *testing.T) {
	expDate, err := time.Parse("02.01.2006.", "25.03.2017.")
	if err != nil {
		t.Errorf("Unexpected error %q.", err)
	}

	sampleExchange := hnb.Exchange{
		Date: expDate,
		Rates: map[string]hnb.Rate{
			"EUR": {big.NewFloat(7.388573), big.NewFloat(7.410805), big.NewFloat(7.433037)},
			"USD": {big.NewFloat(6.839371), big.NewFloat(6.859951), big.NewFloat(6.880531)},
		},
	}

	handler := conversionHandler(&mockExchanger{sampleExchange})

	cases := []struct {
		request    conversionRequest
		respStatus int
		result     float64
	}{
		{
			conversionRequest{10, "USD", "middle"},
			http.StatusOK,
			68.59951,
		},
		{
			conversionRequest{10, "EUR", "buy"},
			http.StatusOK,
			73.88573,
		},
		{
			conversionRequest{10, "EUR", "sell"},
			http.StatusOK,
			74.33037,
		},
		{
			conversionRequest{10, "EUR", "not-valid"},
			http.StatusBadRequest,
			0,
		},
		{
			conversionRequest{10, "WAT", "middle"},
			http.StatusBadRequest,
			0,
		},
	}

	for _, c := range cases {
		func() {
			jsonReqData, err := json.Marshal(c.request)
			if err != nil {
				t.Errorf("Unexpected error %q.", err)
			}

			req, err := http.NewRequest("POST", "http://url-of-a-service", bytes.NewBuffer(jsonReqData))
			if err != nil {
				t.Errorf("Unexpected error %q.", err)
			}

			w := httptest.NewRecorder()

			handler(w, req)

			resp := w.Result()
			defer resp.Body.Close()

			if resp.StatusCode != c.respStatus {
				t.Errorf("Expected %d got %d.", c.respStatus, resp.StatusCode)
			}

			if resp.StatusCode != 200 {
				return
			}

			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				t.Errorf("Unexpected error %q.", err)
			}

			if len(body) == 0 {
				t.Fatalf("Case %q: unexpected response body to be empty", c)
			}

			var recived conversionResponse

			err = json.Unmarshal(body, &recived)
			if err != nil {
				t.Errorf("Unexpected error %q.", err)
			}

			if recived.Result != c.result {
				t.Errorf("Expected %v got %v.", c.result, recived.Result)
			}
		}()
	}
}

type mockExchanger struct {
	hnb.Exchange
}

func (m *mockExchanger) LatestExchange() (hnb.Exchange, error) {
	return m.Exchange, nil
}
