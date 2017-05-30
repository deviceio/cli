package cli

import (
	"crypto/sha512"
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/pquerna/otp/totp"
	"golang.org/x/crypto/ed25519"
)

type HMAPIAuth struct {
	UserID         string
	UserTOTPSecret string
	UserPrivateKey string
}

func (t *HMAPIAuth) Sign(r *http.Request) {
	passcode, err := totp.GenerateCode(t.UserTOTPSecret, time.Now())

	if err != nil {
		logrus.WithField("error", err.Error()).Fatal("error generating totp code")
	}

	message := strings.Join(
		[]string{
			t.UserID,
			passcode,
			r.Method,
			r.Host,
			r.URL.Path,
			r.URL.RawQuery,
			r.Header.Get("Content-Type"),
		},
		"\r\n",
	)

	hash := sha512.New()
	hash.Write([]byte(message))

	keyb, err := base64.StdEncoding.DecodeString(t.UserPrivateKey)

	if err != nil {
		logrus.WithField("error", err.Error()).Fatal("error decoding private key")
	}

	signature := ed25519.Sign(ed25519.PrivateKey(keyb), hash.Sum(nil))

	r.Header.Set("Authorization", fmt.Sprintf(
		"DEVICEIO-HUB-AUTH %v:%v",
		t.UserID,
		base64.StdEncoding.EncodeToString(signature),
	))
}
