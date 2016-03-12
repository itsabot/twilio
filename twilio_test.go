package twilio

import (
	"os"
	"testing"

	"github.com/itsabot/abot/shared/interface/sms"
)

func TestOpenAndSend(t *testing.T) {
	if len(os.Getenv("TWILIO_TEST_ACCOUNT_SID")) == 0 ||
		len(os.Getenv("TWILIO_TEST_AUTH_TOKEN")) == 0 {
		t.Fatal("must set TWILIO_TEST_ACCOUNT_SID and TWILIO_TEST_AUTH_TOKEN env vars")
	}
	auth := os.Getenv("TWILIO_TEST_ACCOUNT_SID") + ":" +
		os.Getenv("TWILIO_TEST_AUTH_TOKEN")
	conn, err := sms.Open("twilio", auth)
	if err != nil {
		t.Fatal(err)
	}
	err = conn.Send("15005550006", "15005550005", "test message")
	if err != nil {
		t.Fatal(err)
	}
}
