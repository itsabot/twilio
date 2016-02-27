// Package twilio implements the github.comitsabot/abot/interface/sms driver
// interface.
package twilio

import (
	"encoding/xml"
	"errors"
	"regexp"
	"strings"

	"github.com/itsabot/abot/shared/interface/sms"
	"github.com/itsabot/abot/shared/interface/sms/driver"
	"github.com/subosito/twilio"
)

type drv struct {
	fromKey string
	toKey   string
	msgKey  string
}

func (d *drv) Open(name string) (driver.Conn, error) {
	auth := strings.Split(name, ":")
	c := conn(*twilio.NewClient(auth[0], auth[1], nil))
	return &c, nil
}

func (d *drv) FromKey() string {
	return d.fromKey
}

func (d *drv) ToKey() string {
	return d.toKey
}

func (d *drv) MsgKey() string {
	return d.msgKey
}

func init() {
	sms.Register("twilio", &drv{
		fromKey: "From",
		toKey:   "To",
		msgKey:  "Body",
	})
}

type conn twilio.Client

// Send an SMS using a Twilio client to a specific phone number in the following
// valid international format ("+13105555555").
func (c *conn) Send(from, to, msg string) error {
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

type phone string

// Valid determines the validity of a phone number according to Twilio's
// expectations for the number's formatting. If valid, error will be nil. If
// invalid, the returned error will contain the reason.
func (p phone) Valid() (valid bool, err error) {
	if len(p) < 10 || len(p) > 20 || !phoneRegex.MatchString(string(p)) {
		return false, errors.New("invalid phone number format: must have E.164 formatting")
	}
	if len(p) == 11 && p[0] != '1' {
		return false, errors.New("unsupported international number")
	}
	if len(p) == 12 && p[0] == '+' && p[1] != '1' {
		return false, errors.New("unsupported international number")
	}
	return true, nil
}

type TwilioResp struct {
	XMLName xml.Name `xml:"Response"`
	Message string
}
