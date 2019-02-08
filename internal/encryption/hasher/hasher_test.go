package hasher

import (
	"math/rand"
	"testing"
)

func randomString(l int) string {
	letters := []rune("1234567890-_!@#$%&*qwertyuiopasdfghjklzxcvbnmQWERTYUIOPASDFGHJKLZXCVBNMйцукенгшщзхъфывапролджэячсмитьбюЙЦУКЕНГШЩЗХЪФЫВАПРОЛДЖЭЯЧСМИТЬБЮ")
	result := make([]rune, l)
	for i := range result {
		result[i] = letters[rand.Intn(len(letters)-1)]
	}
	return string(result)
}

func TestMD5DigestHasher(t *testing.T) {
	t.Log("MD5 digest hashing test")
	t.Log("testing with empty secret")
	unsecureHasher := NewMD5DigestHasher("")
	if unsecureHasher.Hash("") != "" {
		t.Error("Got non-emtpy md5 digest with empty secret and data")
	}
	if unsecureHasher.Hash("some-string") == "" {
		t.Error("Got empty md5 digest with empty secret and non-empty data")
	}
	for i := 1; i <= 100; i++ {
		s := randomString(1 + rand.Intn(99))
		h := unsecureHasher.Hash(s)
		t.Logf("%s -> %s", s, h)
		if len(h) != 32 {
			t.Errorf("Got unexpected (%d) md5 digest length (32) with empty secret for: %s", len(h), s)
		}
	}

	secret := "hasher_secret"
	t.Logf("testing with %q secret", secret)
	hasher := NewMD5DigestHasher(secret)
	if hasher.Hash("") == "" {
		t.Error("Got empty md5 digest with non empty secret and empty data")
	}
	if hasher.Hash("some-string") == "" {
		t.Error("Got empty md5 digest with non-empty secret and data")
	}
	if hasher.Hash("some-string") == unsecureHasher.Hash("some-string") {
		t.Error("MD5 digest generated with secret is equal to hash, generated without secret for the same string")
	}
	for i := 1; i <= 100; i++ {
		s := randomString(1 + rand.Intn(99))
		h := hasher.Hash(s)
		t.Logf("%s -> %s", s, h)
		if len(h) != 32 {
			t.Errorf(
				"Got unexpected (%d) md5 digest length (32) with non-empty secret (%s) for: %s",
				len(h),
				secret,
				s,
			)
		}
	}
}
