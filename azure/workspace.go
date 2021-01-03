package azure

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"
)

func BuildSignature(message, secret string) (string, error) {

	keyBytes, err := base64.StdEncoding.DecodeString(secret)
	if err != nil {
		return "", err
	}

	mac := hmac.New(sha256.New, keyBytes)
	mac.Write([]byte(message))
	return base64.StdEncoding.EncodeToString(mac.Sum(nil)), nil
}

func PostData(workspaceId string, logName string, secretKey string, data string) (err error, code int) {
	dateString := time.Now().UTC().Format(time.RFC1123)
	dateString = strings.Replace(dateString, "UTC", "GMT", -1)

	stringToHash := fmt.Sprintf(`POST
%d
application/json
x-ms-date:%s
/api/logs`, len(data), dateString)
	hashedString, err := BuildSignature(stringToHash, secretKey)
	statusCode := http.StatusInternalServerError
	if err != nil {
		log.Println(err.Error())
		return err, statusCode
	}

	signature := "SharedKey " + workspaceId + ":" + hashedString
	url := "https://" + workspaceId + ".ods.opinsights.azure.com/api/logs?api-version=2016-04-01"

	client := &http.Client{}
	req, err := http.NewRequest("POST", url, bytes.NewReader([]byte(data)))
	if err != nil {
		return err, statusCode
	}

	req.Header.Add("Log-Type", logName)
	req.Header.Add("Authorization", signature)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("x-ms-date", dateString)
	req.Header.Add("time-generated-field", "")

	resp, err := client.Do(req)
	if resp != nil {
		statusCode = resp.StatusCode
		defer resp.Body.Close()
	}
	if err == nil {
		bodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err, statusCode
		}
		bodyString := string(bodyBytes)
		log.Printf(" < Response code:%d body:%s\n", statusCode, bodyString)
		if statusCode >= 400 {
			return errors.New(bodyString), statusCode
		}
		return nil, statusCode
	}
	return err, statusCode
}
