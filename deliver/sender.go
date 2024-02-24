package deliver

import (
	"bytes"
	"crypto/rsa"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/Songmu/go-httpdate"
	"github.com/sirupsen/logrus"
	"github.com/yukimochi/Activity-Relay/models"
)

func sendActivity(inboxURL string, KeyID string, body []byte, privateKey *rsa.PrivateKey) error {
	req, _ := http.NewRequest("POST", inboxURL, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/activity+json")
	req.Header.Set("User-Agent", fmt.Sprintf("%s (golang net/http; Activity-Relay %s; %s)", GlobalConfig.ServerServiceName(), version, GlobalConfig.ServerHostname().Host))
	req.Header.Set("Date", httpdate.Time2Str(time.Now()))
	models.AppendSignature(req, &body, KeyID, privateKey)
	resp, err := HttpClient.Do(req)
	if err != nil {
		urlErr := err.(*url.Error)
		errMsg := ""

		if urlErr.Timeout() {
			errMsg = "Client.Timeout exceeded while awaiting headers"
		} else {
			errMsg = urlErr.Unwrap().Error()
		}
		return errors.New(inboxURL + ": " + errMsg)
	}
	defer resp.Body.Close()

	logrus.Debug(inboxURL, " ", resp.StatusCode)
	if resp.StatusCode/100 != 2 {
		return errors.New(inboxURL + ": " + resp.Status)
	}

	return nil
}
