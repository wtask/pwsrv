package email

import "testing"

func TestValidAddresses(t *testing.T) {
	cases := []struct {
		email                     string
		expAddr, expAddrASCII     string
		expDomain, expDomainASCII string
		expLocalPart, expUserName string
	}{
		{"user@gmail.com",
			"user@gmail.com", "user@gmail.com", "gmail.com", "gmail.com", "user", ""},
		{"user+suffix@gmail.com",
			"user+suffix@gmail.com", "user+suffix@gmail.com", "gmail.com", "gmail.com", "user+suffix", ""},
		{"user+suffix@mail.gmail.com",
			"user+suffix@mail.gmail.com", "user+suffix@mail.gmail.com",
			"mail.gmail.com", "mail.gmail.com", "user+suffix", ""},
		{"user.name@g-mail.com",
			"user.name@g-mail.com", "user.name@g-mail.com", "g-mail.com", "g-mail.com", "user.name", ""},
		{"user.name_!_!_!@gmail.com",
			"user.name_!_!_!@gmail.com", "user.name_!_!_!@gmail.com", "gmail.com", "gmail.com", "user.name_!_!_!", ""},
		{"User Name <user.name@g-mail.com>",
			"user.name@g-mail.com", "user.name@g-mail.com", "g-mail.com", "g-mail.com", "user.name", "User Name"},
		{"   User Name <user.name@g-mail.com>   ",
			"user.name@g-mail.com", "user.name@g-mail.com", "g-mail.com", "g-mail.com", "user.name", "User Name"},
		{"Иван Иванов <ваня@g-mail.com>",
			"ваня@g-mail.com", "ваня@g-mail.com", "g-mail.com", "g-mail.com", "ваня", "Иван Иванов"},
		{"Иван Иванов <ваня@почта.рф>",
			"ваня@почта.рф", "ваня@xn--80a1acny.xn--p1ai", "почта.рф", "xn--80a1acny.xn--p1ai", "ваня", "Иван Иванов"},
		{"Иван Иванов <ваня@xn--80a1acny.xn--p1ai>",
			"ваня@почта.рф", "ваня@xn--80a1acny.xn--p1ai", "почта.рф", "xn--80a1acny.xn--p1ai", "ваня", "Иван Иванов"},
		{"/user/name/@www.gmail.com",
			"/user/name/@www.gmail.com", "/user/name/@www.gmail.com", "www.gmail.com", "www.gmail.com", "/user/name/", ""},
	}
	for _, c := range cases {
		a := NewAddress(c.email)
		if !a.IsValid() {
			t.Errorf("Unexpected invalid address (%s): %q", c.email, a.Error())
		}
		if a.Get() != c.expAddr {
			t.Errorf("Address.Get(), expected: %q, actual: %q", c.expAddr, a.Get())
		}
		if a.GetASCII() != c.expAddrASCII {
			t.Errorf("Address.GetASCII(), expected: %q, actual: %q", c.expAddrASCII, a.GetASCII())
		}
		if a.Domain() != c.expDomain {
			t.Errorf("Address.Domain(), expected: %q, actual: %q", c.expDomain, a.Domain())
		}
		if a.DomainASCII() != c.expDomainASCII {
			t.Errorf("Address.DomainASCII(), expected: %q, actual: %q", c.expDomainASCII, a.DomainASCII())
		}
		if a.LocalPart() != c.expLocalPart {
			t.Errorf("Address.LocalPart(), expected: %q, actual: %q", c.expLocalPart, a.LocalPart())
		}
		if a.UserName() != c.expUserName {
			t.Errorf("Address.UserName(), expected: %q, actual: %q", c.expUserName, a.UserName())
		}
	}
}

func TestInvalidAddresses(t *testing.T) {
	cases := []struct {
		email string
	}{
		{""},
		{" "},
		{" user "},
		{"user.gmail.com"},
		{"user.@gmail.com"},
		{"user@gmail.com."},
		{"user@gmail!com"},
		{`"user@gmail.com"`},
		{`user@xn--gmail.com`},
		{"<User Name <user.name@g-mail.com>"},
		{"- User Name <user.name@g-mail.com> -"},
		{"<user.name@www.gmail.com/search?>"},
		{"user:name@gmail.com"},
		{"<user>@gmail.com"},
		{"@gmail.com"},
	}
	for _, c := range cases {
		a := NewAddress(c.email)
		if a.IsValid() {
			t.Errorf("Invalid source email %q been valid!", c.email)
		}
	}
}

func TestEmptyAddress(t *testing.T) {
	if (&Address{}).IsValid() {
		t.Error("Zero value of *Address{} became valid but must not!")
	}
}

func TestNilReceiver(t *testing.T) {
	var a *Address
	if a.IsValid() {
		t.Error("*Address<nil>.IsValid() must not return true!")
	}
	if a.Error() == "" {
		t.Error("*Address<nil>.Error() must return non-empty error message!")
	}
	if a.Get() != "" {
		t.Errorf("*Address<nil>.Get() must return empty string, got: %q", a.Get())
	}
	if a.GetASCII() != "" {
		t.Errorf("*Address<nil>.GetASCII() must return empty string, got: %q", a.GetASCII())
	}
	if a.Domain() != "" {
		t.Errorf("*Address<nil>.Domain() must return empty string, got: %q", a.Domain())
	}
	if a.DomainASCII() != "" {
		t.Errorf("*Address<nil>.DomainASCII() must return empty string, got: %q", a.DomainASCII())
	}
	if a.LocalPart() != "" {
		t.Errorf("*Address<nil>.LocalPart() must return empty string, got: %q", a.LocalPart())
	}
	if a.UserName() != "" {
		t.Errorf("*Address<nil>.UserName() must return empty string, got: %q", a.UserName())
	}
}
