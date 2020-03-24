package adapters

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/hublabs/common/api"

	try "github.com/matryer/try"
	"github.com/pangpanglabs/goutils/httpreq"
	"github.com/pangpanglabs/goutils/test"
	"github.com/sirupsen/logrus"
)

func TestResrApiRetry(t *testing.T) {
	responseCounter := 0
	responses := []func(w http.ResponseWriter, r *http.Request){
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(500)
			http.Error(w, fmt.Sprintf("ioutil.ReadAll: 1st error"), 500)
			response, _ := json.Marshal(map[string]interface{}{
				"success": false,
				"result":  1,
				"error":   errors.New("1st error"),
			})
			w.Write([]byte(response))
		},
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(500)
			http.Error(w, fmt.Sprintf("ioutil.ReadAll: 1st error"), 500)
			response, _ := json.Marshal(map[string]interface{}{
				"success": false,
				"result":  2,
				"error":   errors.New("2st error"),
			})
			w.Write([]byte(response))
		},
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)
			response, _ := json.Marshal(map[string]interface{}{
				"success": true,
				"result":  3,
				"error":   nil,
			})
			w.Write([]byte(response))
		},
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		responses[responseCounter](w, r)
		responseCounter++
	}))
	defer ts.Close()

	var resp struct {
		Success bool        `json:"success"`
		Result  interface{} `json:"result"`
		Error   api.Error   `json:"error"`
	}

	t.Run("RetryCheck", func(t *testing.T) {
		err := try.Do(func(attempt int) (bool, error) {
			_, err := httpreq.New(http.MethodGet, ts.URL, nil).
				CallWithClient(&resp, client)
			logrus.WithField("attempt", attempt).Info("attempt")
			if err != nil {
				time.Sleep(2 * time.Second) // wait a Second
			}
			return attempt < 5, err
		})
		test.Ok(t, err)
		test.Equals(t, resp.Success, true)
		test.Equals(t, resp.Result, float64(3))
	})
}
