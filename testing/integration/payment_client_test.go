package integration

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"testing"
	"time"

	"answer/task3/testing/setup"
)

func TestPaymentClient_ExternalAPI(t *testing.T) {
	wiremock := setup.StartWireMock(t)

	t.Run("success", func(t *testing.T) {
		stubWireMock(t, wiremock.BaseURL, `{
		  "request": {
		    "method": "POST",
		    "url": "/payments"
		  },
		  "response": {
		    "status": 200,
		    "jsonBody": {
		      "id": "pay_123",
		      "status": "authorized"
		    },
		    "headers": {
		      "Content-Type": "application/json"
		    }
		  }
		}`)

		// Example:
		// client := payment.NewClient(wiremock.BaseURL, httpClient)
		// resp, err := client.Authorize(context.Background(), payment.Request{...})
		// require.NoError(t, err)
		// require.Equal(t, "authorized", resp.Status)
	})

	t.Run("failure", func(t *testing.T) {
		stubWireMock(t, wiremock.BaseURL, `{
		  "request": {
		    "method": "POST",
		    "url": "/payments"
		  },
		  "response": {
		    "status": 402,
		    "jsonBody": {
		      "code": "card_declined",
		      "message": "card was declined"
		    },
		    "headers": {
		      "Content-Type": "application/json"
		    }
		  }
		}`)

		// Assert the client maps provider-specific response codes into domain-safe errors.
	})

	t.Run("timeout", func(t *testing.T) {
		stubWireMock(t, wiremock.BaseURL, `{
		  "request": {
		    "method": "POST",
		    "url": "/payments"
		  },
		  "response": {
		    "status": 200,
		    "fixedDelayMilliseconds": 3000,
		    "jsonBody": {
		      "status": "authorized"
		    },
		    "headers": {
		      "Content-Type": "application/json"
		    }
		  }
		}`)

		client := &http.Client{Timeout: time.Second}
		req, _ := http.NewRequestWithContext(context.Background(), http.MethodPost, wiremock.BaseURL+"/payments", bytes.NewBufferString(`{}`))
		req.Header.Set("Content-Type", "application/json")

		_, err := client.Do(req)
		if err == nil {
			t.Fatal("expected timeout error")
		}
	})
}

func stubWireMock(t *testing.T, baseURL, body string) {
	t.Helper()

	req, err := http.NewRequest(http.MethodPost, baseURL+"/__admin/mappings", bytes.NewBufferString(body))
	if err != nil {
		t.Fatalf("new wiremock request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("register wiremock stub: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusMultipleChoices {
		data, _ := io.ReadAll(resp.Body)
		t.Fatalf("wiremock stub failed: status=%d body=%s", resp.StatusCode, string(data))
	}
}
