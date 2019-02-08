package email

import (
	"errors"
	"fmt"
	"net/mail"
	"strings"

	"golang.org/x/net/idna"
)

// Address - contains unexported fields only of parsed email or mailto-parts.
type Address struct {
	user, domain, domainASCII, name string
	err                             error
}

// Get - returns email address in Unicode as if it was properly parsed or empty string otherwise.
func (a *Address) Get() string {
	if !a.IsValid() {
		return ""
	}
	return fmt.Sprintf("%s@%s", a.user, a.domain)
}

// GetASCII - returns email address in ASCII punycode if it was properly parsed or empty string otherwise.
func (a *Address) GetASCII() string {
	if !a.IsValid() {
		return ""
	}
	return fmt.Sprintf("%s@%s", a.user, a.domainASCII)
}

// Domain - returns domain-part of email in Unicode if address is valid or empty string otherwise.
func (a *Address) Domain() string {
	if !a.IsValid() {
		return ""
	}
	return a.domain
}

// DomainASCII - eturns domain-part of email in ASCII punycode if address is valid or empty string otherwise.
func (a *Address) DomainASCII() string {
	if !a.IsValid() {
		return ""
	}
	return a.domainASCII
}

// UserPart - returns user-part of email in Unicode or empty string if address is not valid.
func (a *Address) UserPart() string {
	if a == nil {
		return ""
	}
	return a.user
}

// UserName - returns email owner name, if it was set in source mailto like this `John Smith <john@smith.com>`
func (a *Address) UserName() string {
	if a == nil {
		return ""
	}
	return a.name
}

// IsValid - returns parsing and IDNA-checking result of given email.
// Always returns false for Address zero values.
func (a *Address) IsValid() bool {
	return a != nil &&
		a.err == nil &&
		a.user != "" && a.domain != "" && a.domainASCII != ""
}

// Error - returns error description if it was occurred.
func (a *Address) Error() string {
	if a == nil {
		return "mailto: nil *Address"
	}
	if a.err != nil {
		return a.err.Error()
	}
	return ""
}

// NewAddress - creates address instance after parsing and validation (mostly sematic) of given mailto.
func NewAddress(mailTo string) *Address {
	addr := &Address{}
	ma, err := mail.ParseAddress(mailTo)
	if err != nil {
		addr.err = err
		return addr
	}
	user, domain, err := splitAddress(ma.Address)
	if err != nil {
		addr.err = err
		return addr
	}
	addr.user, addr.name = user, ma.Name
	idn := idna.New(
		idna.RemoveLeadingDots(false),
		idna.StrictDomainName(true),
		idna.Transitional(false),
		idna.ValidateForRegistration(),
		idna.ValidateLabels(true),
		idna.VerifyDNSLength(true),
	)
	addr.domain, addr.err = idn.ToUnicode(domain)
	if addr.err != nil {
		return addr
	}
	addr.domainASCII, addr.err = idn.ToASCII(domain)
	return addr
}

// splitAddress - helper to split email into 2 part by @.
func splitAddress(address string) (user, domain string, err error) {
	d := strings.SplitN(address, "@", 2)
	if len(d) != 2 {
		return "", "", errors.New("mailto: there is no domain part")
	}
	return d[0], d[1], nil
}
