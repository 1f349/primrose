package smtp

import (
	"bytes"
	"errors"
	"github.com/emersion/go-message"
	"github.com/emersion/go-message/mail"
	"strings"
	"time"
)

var (
	ErrInvalidBodyType       = errors.New("invalid body type")
	ErrMultipleFromAddresses = errors.New("multiple from addresses")
)

type Json struct {
	From     string `json:"from"`
	ReplyTo  string `json:"reply_to"`
	To       string `json:"to"`
	Cc       string `json:"cc"`
	Bcc      string `json:"bcc"`
	Subject  string `json:"subject"`
	BodyType string `json:"body_type"`
	Body     string `json:"body"`
}

func (s Json) parseAddresses() (addrFrom, addrReplyTo, addrTo, addrCc, addrBcc []*mail.Address, err error) {
	// parse addresses
	addrFrom, err = mail.ParseAddressList(s.From)
	if err != nil {
		return
	}
	addrReplyTo, err = mail.ParseAddressList(s.ReplyTo)
	if err != nil {
		return
	}
	addrTo, err = mail.ParseAddressList(s.To)
	if err != nil {
		return
	}
	addrCc, err = mail.ParseAddressList(s.Cc)
	if err != nil {
		return
	}
	addrBcc, err = mail.ParseAddressList(s.Bcc)
	return
}

func (s Json) PrepareMail() (*Mail, error) {
	// parse addresses from json data
	addrFrom, addrReplyTo, addrTo, addrCc, addrBcc, err := s.parseAddresses()
	if err != nil {
		return nil, err
	}

	// only one from address allowed here
	if len(addrFrom) != 1 {
		return nil, ErrMultipleFromAddresses
	}

	// save for use in the caller
	from := addrFrom[0].Address

	// set base headers
	var h mail.Header
	h.SetDate(time.Now())
	h.SetSubject(s.Subject)
	h.SetAddressList("From", addrFrom)
	h.SetAddressList("Reply-To", addrReplyTo)
	h.SetAddressList("To", addrTo)
	h.SetAddressList("Cc", addrCc)

	// set content type header
	switch s.BodyType {
	case "plain":
		h.Set("Content-Type", "text/plain; charset=utf-8")
	case "html":
		h.Set("Content-Type", "text/html; charset=utf-8")
	default:
		return nil, ErrInvalidBodyType
	}

	entity, err := message.New(h.Header, strings.NewReader(s.Body))
	if err != nil {
		return nil, err
	}

	m := &Mail{
		from:    from,
		deliver: CreateSenderSlice(addrTo, addrCc, addrBcc),
	}

	out := new(bytes.Buffer)
	if err := entity.WriteTo(out); err != nil {
		return nil, err
	}

	m.body = out.Bytes()
	return m, nil
}