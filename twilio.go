// Package twilio implements the github.comitsabot/abot/interfaces/sms
// driver interface.
package twilio

import (
	"bytes"
	"encoding/xml"
	"errors"
	"net/http"
	"os"
	"regexp"

	"github.com/itsabot/abot/core"
	"github.com/labstack/echo"
	"github.com/subosito/twilio"
)

// PhoneRegex determines whether a string is a phone number
var PhoneRegex = regexp.MustCompile(`^\+?[0-9\-\s()]+$`)

// NewClient returns an authorized Twilio client using TWILIO_ACCOUNT_SID and
// TWILIO_AUTH_TOKEN environment variables.
func NewClient() *twilio.Client {
	accountSID := os.Getenv("TWILIO_ACCOUNT_SID")
	authToken := os.Getenv("TWILIO_AUTH_TOKEN")
	return twilio.NewClient(accountSID, authToken, nil)
}

// SentMessage sends an SMS using a Twilio client to a specific phone number in
// the following valid international format ("+13105555555") from an owned
// Twilio phone number retrieved from the TWILIO_NUMBER environment variable.
func SendMessage(tc *twilio.Client, to, msg string) error {
	params := twilio.MessageParams{Body: msg}
	_, _, err := tc.Messages.Send(os.Getenv("TWILIO_NUMBER"), to, params)
	return err
}

// TwilioResp is an XML struct constructed as a response to Twilio's API to
// respond to user messages via SMS.
type TwilioResp struct {
	XMLName xml.Name `xml:"Response"`
	Message string
}

// StringToTwiml converts a string, such as a message to send to a user, into an
// XML struct that Twilio can understand.
func StringToTwiml(s string) (string, error) {
	var buf bytes.Buffer
	r := &TwilioResp{Message: s}
	enc := xml.NewEncoder(&buf)
	if err := enc.Encode(r); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// handlerTwilio responds to SMS messages with XML. Unlike other handlers, we
// process internal errors without returning here, since any errors should not
// be presented directly to the user -- they should be "humanized"
func handlerTwilio(c *echo.Context) error {
	c.Set("cmd", c.Form("Body"))
	c.Set("flexid", c.Form("From"))
	c.Set("flexidtype", 2)
	errMsg := "Something went wrong with my wiring... I'll get that fixed up soon."
	errSent := false
	ret, uid, err := core.ProcessText(db, mc, ner, offensive, c)
	if err != nil {
		ret = errMsg
		errSent = true
		handlerError(err, c)
	}
	if err = ws.NotifySockets(c, uid, c.Form("Body"), ret); err != nil {
		if !errSent {
			ret = errMsg
			errSent = true
			handlerError(err, c)
		}
	}
	// TODO take SMS connection and Send a message
	var resp sms.TwilioResp
	if len(ret) == 0 {
		resp = sms.TwilioResp{}
	} else {
		resp = sms.TwilioResp{Message: ret}
	}
	if err = c.XML(http.StatusOK, resp); err != nil {
		if !errSent {
			handlerError(err, c)
		}
	}
	return nil
}

func validatePhone(s string) error {
	if len(s) < 10 || len(s) > 20 || !PhoneRegex.MatchString(s) {
		return errors.New(
			"Your phone must be a valid U.S. number with the area code.")
	}
	if len(s) == 11 && s[0] != '1' {
		return errors.New(
			"Sorry, Ava only serves U.S. customers for now.")
	}
	if len(s) == 12 && s[0] == '+' && s[1] != '1' {
		return errors.New(
			"Sorry, Ava only serves U.S. customers for now.")
	}
	return nil
}
