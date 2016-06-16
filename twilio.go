// Package twilio implements the github.com/itsabot/abot/interface/sms/driver
// interface.
package twilio

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"errors"
	"net/http"
	"os"
	"regexp"

	"github.com/itsabot/abot/core"
	"github.com/itsabot/abot/core/log"
	"github.com/itsabot/abot/shared/datatypes"
	"github.com/itsabot/abot/shared/interface/sms"
	"github.com/itsabot/abot/shared/interface/sms/driver"
	"github.com/julienschmidt/httprouter"
	"github.com/subosito/twilio"
)

type drv struct{}

func (d *drv) Open(r *httprouter.Router) (driver.Conn, error) {
	c := conn(*twilio.NewClient(os.Getenv("TWILIO_ACCOUNT_SID"),
		os.Getenv("TWILIO_AUTH_TOKEN"), nil))
	hm := dt.NewHandlerMap([]dt.RouteHandler{
		{
			// Path is prefixed by "twilio" automatically. Thus the
			// path below becomes "/twilio"
			Path:    "/",
			Method:  "POST",
			Handler: handlerTwilio,
		},
	})
	hm.AddRoutes("twilio", r)
	return &c, nil
}

func init() {
	sms.Register("twilio", &drv{})
}

type conn twilio.Client

// Send an SMS using a Twilio client to a specific phone number in the following
// valid international format ("+13105555555"). From is handled by the driver.
func (c *conn) Send(to, msg string) error {
	var from string
	if os.Getenv("ABOT_ENV") == "test" {
		from = "15005550006"
	} else {
		from = os.Getenv("TWILIO_PHONE")
	}
	params := twilio.MessageParams{Body: msg}
	_, _, err := c.Messages.Send(from, to, params)
	return err
}

// Close the connection, but since Twilio connections are open as needed, there
// is nothing for us to close here. Return nil.
func (c *conn) Close() error {
	return nil
}

// phoneRegex determines whether a string is a phone number
var phoneRegex = regexp.MustCompile(`^\+?[1-9]\d{1,14}$`)

type Phone string

// Valid determines the validity of a phone number according to Twilio's
// expectations for the number's formatting. If valid, error will be nil. If
// invalid, the returned error will contain the reason.
func (p Phone) Valid() (valid bool, err error) {
	if len(p) < 10 || len(p) > 20 || !phoneRegex.MatchString(string(p)) {
		return false, errors.New("invalid phone number format: must have E.164 formatting")
	}
	if len(p) == 11 && p[0] != '1' {
		return false, errors.New("unsupported international number")
	}
	if p[0] != '+' {
		return false, errors.New("first character in phone number must be +")
	}
	if len(p) == 12 && p[1] != '1' {
		return false, errors.New("unsupported international number")
	}
	return true, nil
}

// TwilioResp is a valid XML Twilio struct that contains an SMS response to be
// sent back to the user.
type TwilioResp struct {
	XMLName xml.Name `xml:"Response"`
	Message string
}

// handlerTwilio responds to SMS messages sent through Twilio. Unlike other
// handlers, we process internal errors without returning here, since any errors
// should not be presented directly to the user -- they should be "humanized"
func handlerTwilio(w http.ResponseWriter, r *http.Request) {
	var ret string
	if err := r.ParseForm(); err != nil {
		log.Info("failed parsing twilio post form.", err)
		ret = "Something went wrong with my wiring... I'll get that fixed up soon."
	}
	tmp := struct {
		CMD        string
		FlexID     string
		FlexIDType dt.FlexIDType
	}{
		CMD:        r.FormValue("Body"),
		FlexID:     r.FormValue("From"),
		FlexIDType: 2,
	}
	byt, err := json.Marshal(tmp)
	if err != nil {
		log.Info("failed marshaling req struct.", err)
		ret = "Something went wrong with my wiring... I'll get that fixed up soon."
	}
	r, err = http.NewRequest(r.Method, r.URL.String(), bytes.NewBuffer(byt))
	if err != nil {
		log.Info("failed building http request.", err)
		ret = "Something went wrong with my wiring... I'll get that fixed up soon."
	}
	ret, err = core.ProcessText(r)
	if err != nil {
		log.Info("failed processing text.", err)
		ret = "Something went wrong with my wiring... I'll get that fixed up soon."
	}
	if _, err = w.Write([]byte(ret)); err != nil {
		log.Info("failed to write response.", err)
	}
}
