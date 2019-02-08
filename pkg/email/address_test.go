package email

import "testing"

func TestValidAddresses(t *testing.T) {
	cases := []struct {
		mailTo                    string
		expAddr, expAddrASCII     string
		expDomain, expDomainASCII string
		expUserPart, expUserName  string
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
		a := NewAddress(c.mailTo)
		if !a.IsValid() {
			t.Errorf("Unexpected invalid address (%s): %q", c.mailTo, a.Error())
		}
		if a.Get() != c.expAddr {
			t.Errorf(".Get(), expected: %q, actual: %q", c.expAddr, a.Get())
		}
		if a.GetASCII() != c.expAddrASCII {
			t.Errorf(".GetASCII(), expected: %q, actual: %q", c.expAddrASCII, a.GetASCII())
		}
		if a.Domain() != c.expDomain {
			t.Errorf(".Domain(), expected: %q, actual: %q", c.expDomain, a.Domain())
		}
		if a.DomainASCII() != c.expDomainASCII {
			t.Errorf(".DomainASCII(), expected: %q, actual: %q", c.expDomainASCII, a.DomainASCII())
		}
		if a.UserPart() != c.expUserPart {
			t.Errorf(".UserPart(), expected: %q, actual: %q", c.expUserPart, a.UserPart())
		}
		if a.UserName() != c.expUserName {
			t.Errorf(".UserName(), expected: %q, actual: %q", c.expUserName, a.UserName())
		}
	}
}

func TestInvalidAddresses(t *testing.T) {
	cases := []struct {
		mailTo string
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
		a := NewAddress(c.mailTo)
		if a.IsValid() {
			t.Errorf("Invalid mailto %q expression been valid", c.mailTo)
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
		t.Error("Nil *Address .IsValid() must not return true!")
	}
	if a.Error() == "" {
		t.Error("Nil *Address .Error() must return non-empty error message!")
	}
	if a.Get() != "" {
		t.Errorf("Nil *Address .Get() must return empty string, got: %q", a.Get())
	}
	if a.GetASCII() != "" {
		t.Errorf("Nil *Address .GetASCII() must return empty string, got: %q", a.GetASCII())
	}
	if a.Domain() != "" {
		t.Errorf("Nil *Address .Domain() must return empty string, got: %q", a.Domain())
	}
	if a.DomainASCII() != "" {
		t.Errorf("Nil *Address .DomainASCII() must return empty string, got: %q", a.DomainASCII())
	}
	if a.UserPart() != "" {
		t.Errorf("Nil *Address .UserPart() must return empty string, got: %q", a.UserPart())
	}
	if a.UserName() != "" {
		t.Errorf("Nil *Address .UserName() must return empty string, got: %q", a.UserName())
	}
}
