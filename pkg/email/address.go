package email

import (
	"errors"
	"fmt"
	"net/mail"
	"strings"

	"golang.org/x/net/idna"
)

// Address - contains unexported fields only parsed email.
type Address struct {
	localPart, domain, domainASCII, userName string
	err                                      error
}

// Get - returns email address in Unicode as if it was properly parsed or empty string otherwise.
func (a *Address) Get() string {
	if !a.IsValid() {
		return ""
	}
	return fmt.Sprintf("%s@%s", a.localPart, a.domain)
}

// GetASCII - returns email address in ASCII punycode if it was properly parsed or empty string otherwise.
func (a *Address) GetASCII() string {
	if !a.IsValid() {
		return ""
	}
	return fmt.Sprintf("%s@%s", a.localPart, a.domainASCII)
}

// Domain - returns domain-part of email in Unicode if address is valid or empty string otherwise.
func (a *Address) Domain() string {
	if !a.IsValid() {
		return ""
	}
	return a.domain
}

// DomainASCII - returns domain-part of email in ASCII punycode if address is valid or empty string otherwise.
func (a *Address) DomainASCII() string {
	if !a.IsValid() {
		return ""
	}
	return a.domainASCII
}

// LocalPart - returns local-part of email in Unicode or empty string if address is not valid.
func (a *Address) LocalPart() string {
	if a == nil {
		return ""
	}
	return a.localPart
}

// UserName - returns user name for email address if it was set in source.
// For example, if source was string `John Smith <john@smith.com>` this method should return `John Smith`.
func (a *Address) UserName() string {
	if a == nil {
		return ""
	}
	return a.userName
}

// IsValid - returns parsing and IDNA-checking result of given email.
// Always returns false for Address zero values.
func (a *Address) IsValid() bool {
	return a != nil &&
		a.err == nil &&
		a.localPart != "" && a.domain != "" && a.domainASCII != ""
}

// Error - returns error description if it was occurred.
func (a *Address) Error() string {
	if a == nil {
		return "email: uninitialized *Address"
	}
	if a.err != nil {
		return a.err.Error()
	}
	return ""
}

// NewAddress - creates Address instance after parsing and validating (mostly sematic) of given source string.
func NewAddress(address string) *Address {
	addr := &Address{}
	ma, err := mail.ParseAddress(address)
	if err != nil {
		addr.err = err
		return addr
	}
	localPart, domain, err := splitAddress(ma.Address)
	if err != nil {
		addr.err = err
		return addr
	}
	addr.localPart, addr.userName = localPart, ma.Name
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
func splitAddress(address string) (localPart, domain string, err error) {
	d := strings.SplitN(address, "@", 2)
	if len(d) != 2 {
		return "", "", errors.New("email: can not split local and domain parts")
	}
	return d[0], d[1], nil
}
