// Copyright © 2022 Sloan Childers
package sink

import (
	"fmt"
	"net"
	"net/smtp"
	"strconv"
	"strings"
	"time"

	"github.com/mcnijman/go-emailaddress"
	"github.com/rs/zerolog/log"
	whois "github.com/undiabler/golang-whois"
)

type INetwork interface {
	SMTPCheck(em *emailaddress.EmailAddress) EmailLiveLookupInfo
	PreferredMX(addr string) string
	ContactMx(host string, domain string, email string, timeout time.Duration) error
	WhoIs(domain string) (*WhoIsInfo, error)
}

type WhoIsInfo struct {
	Domain             string
	DomainAgeInDays    int
	DomainAgeInYears   int
	DomainAgeDate      string
	IsRegisteredDomain bool
}

type Network struct {
	mxs map[string]string
}

type EmailLiveLookupInfo struct {
	Email              string
	IsValidEmail       bool
	IsValidIcannSuffix bool
	IsResolveable      bool
	IsBadAccount       bool
	IsSmtpVerified     bool
	Reason             string
}

func NewNetwork() *Network {
	cache_mx := make(map[string]string)
	return &Network{
		mxs: cache_mx}
}

func (x *Network) SMTPCheck(em *emailaddress.EmailAddress) EmailLiveLookupInfo {
	var info EmailLiveLookupInfo
	info.Email = em.String()
	info.IsValidEmail = false
	info.IsValidIcannSuffix = false
	info.IsResolveable = false
	info.IsSmtpVerified = false
	info.IsBadAccount = false
	info.Reason = "unknown status"

	mx := x.PreferredMX(em.Domain)
	if mx != "" {
		info.IsResolveable = true
		err := x.ContactMx(mx, em.Domain, em.String(), time.Second*2)
		if err == nil {
			info.IsSmtpVerified = true
			info.Reason = ""
		} else {
			if strings.Contains(err.Error(), "i/o timeout") {
				info.Reason = "SMTP I/O timeout"
			} else {
				// Common SMTP 400 error codes
				// Error code	Description
				// 421	Service isn't available, try again later
				// 450	Requested action wasn't taken because the user's mailbox was unavailable
				// 451	Message not sent because of server error
				// 452	Command stopped because there isn’t enough server storage
				// 455	Server can't deal with the command right now

				// Common SMTP 500 error codes
				// Error code	Description
				// 500	Server couldn't recognize the command because of a syntax error
				// 501	Syntax error found in command parameters or arguments
				// 502	Command not implemented
				// 503	Server had bad sequence of commands
				// 541	Message rejected by the recipient address
				// 550	Requested command failed because the user’s mailbox was unavailable, or the receiving server rejected the message because it was likely spam
				// 551	Intended recipient mailbox isn't available on the receiving server
				// 552	Message wasn't sent because the recipient mailbox doesn't have enough storage
				// 553	Command stopped because the mailbox name doesn't exist
				// 554	Transaction failed without additional details
				code, _ := strconv.Atoi(err.Error()[0:3])
				if code >= 550 && code <= 553 {
					info.IsBadAccount = true
					info.IsSmtpVerified = true
					info.Reason = fmt.Sprintf("SMTP error code %d", code)
				}
			}
		}
	}

	return info
}

func (x *Network) WhoIs(domain string) (*WhoIsInfo, error) {
	var info WhoIsInfo
	info.Domain = domain

	result, err := whois.GetWhoisTimeout(domain, time.Second)
	if err != nil {
		return nil, err
	}

	var date string
	var days int
	for _, line := range strings.Split(result, "\n") {
		// HACK ALERT:  this is fragile, there needs to be a better way
		if strings.Contains(line, "Creation Date:") {
			date = strings.Split(line, ": ")[1]
			date = strings.Split(date, "T")[0]
			days, err = GetDaysFromDate(date)
		}
	}

	info.DomainAgeDate = date
	if err != nil || days == -1 {
		log.Error().Str("component", "network").Str("key", domain).Msg("whois")
		return nil, err
	} else {
		info.DomainAgeInDays = days
		info.DomainAgeInYears = days / 365
		info.IsRegisteredDomain = true
	}

	return &info, nil
}

// Finds the MX record for the highest priority mail server in the list
func (x *Network) PreferredMX(addr string) string {
	mx := x.mxs[addr]
	if mx == "" {
		if mxs, err := net.LookupMX(addr); err == nil {
			pref := mxs[0].Pref
			host := mxs[0].Host
			for _, mx := range mxs {
				if mx.Pref < pref {
					pref = mx.Pref
					host = mx.Host
				}
			}
			x.mxs[addr] = host
			return host
		}
	}
	return mx
}

func (x *Network) ContactMx(host string, domain string, email string, timeout time.Duration) error {
	// net.smtp doesn't use a timeout for the connection, it depends
	//   on the OS to kill the connection, could be 2+ minutes at least
	//   otherwise we'd use the funcationality built into emailaddress.ValidateHost
	d := net.Dialer{Timeout: timeout}
	conn, err := d.Dial("tcp", fmt.Sprintf("%s:%d", host, 25))
	if err != nil {
		return err
	}
	client, err := smtp.NewClient(conn, host)
	if err != nil {
		return err
	}
	defer client.Close()

	if err = client.Hello(domain); err == nil {
		if err = client.Mail(fmt.Sprintf("hello@%s", domain)); err == nil {
			if err = client.Rcpt(email); err == nil {
				client.Reset()
				client.Quit()
				return nil
			}
		}
	}
	return err
}
