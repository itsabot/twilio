package twilio

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/itsabot/abot/core"
	"github.com/itsabot/abot/shared/interface/sms"
	"github.com/itsabot/abot/shared/log"
	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo"
)

var e *echo.Echo
var db *sqlx.DB
var ner core.Classifier
var offensive map[string]struct{}

func TestMain(m *testing.M) {
	if len(os.Getenv("TWILIO_TEST_ACCOUNT_SID")) == 0 ||
		len(os.Getenv("TWILIO_TEST_AUTH_TOKEN")) == 0 {
		log.Info("must set TWILIO_TEST_ACCOUNT_SID and TWILIO_TEST_AUTH_TOKEN env vars")
		os.Exit(1)
	}
	var err error
	e, err = core.NewServer()
	if err != nil {
		log.Info("couldnt boot server", err)
		os.Exit(1)
	}
	os.Exit(m.Run())
}

func TestOpenAndSend(t *testing.T) {
	auth := os.Getenv("TWILIO_TEST_ACCOUNT_SID") + ":" +
		os.Getenv("TWILIO_TEST_AUTH_TOKEN")
	conn, err := sms.Open("twilio", auth, e)
	if err != nil {
		log.Info("opening conn")
		t.Fatal(err)
	}
	err = conn.Send("15005550006", "15005550005", "test message")
	if err != nil {
		log.Info("sending msg")
		t.Fatal(err)
	}
}

func TestHandler(t *testing.T) {
	u := fmt.Sprintf("http://localhost:%s/twilio", os.Getenv("PORT"))
	data := []byte(`{ "Body": "Test message", "From": "+15005551234" }`)
	c, _ := request("POST", u, data, e)
	if c != 200 {
		t.Fatal("expected 200. got", c)
	}
}

func request(method, path string, data []byte, e *echo.Echo) (int, string) {
	r, err := http.NewRequest(method, path, bytes.NewBuffer(data))
	if err != nil {
		return 0, "err completing request: " + err.Error()
	}
	w := httptest.NewRecorder()
	e.ServeHTTP(w, r)
	return w.Code, w.Body.String()
}
