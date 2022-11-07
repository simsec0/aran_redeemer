package redeemer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"

	http "github.com/bogdanfinn/fhttp"
)

func (in *Instance) Redeem() (string, error) {
	payload, err := json.Marshal(map[string]interface{}{
		"channel_id":        nil,
		"payment_source_id": in.PaymentSourceId,
	})

	if err != nil {
		return "", err
	}

	link := fmt.Sprintf("https://discord.com/api/v9/entitlements/gift-codes/%s/redeem", strings.Split(in.PromoCode, "https://discord.com/billing/promotions/")[1])
	req, err := http.NewRequest("POST", link, bytes.NewBuffer(payload))

	if err != nil {
		return "", err
	}

	req = DiscordHeaders(true, map[string]string{
		"X-Debug-Options":    "bugReporterEnabled",
		"Content-Type":       "application/json",
		"X-Discord-Locale":   "en-US",
		"X-Super-Properties": in.XSuperProp,
		"Host":               "discord.com",
		"Referer":            in.PromoCode,
		"Authorization":      in.Token,
	}, req)

	resp, err := in.Client.Do(req)

	if err != nil {
		return "", err
	}

	defer resp.Body.Close()
	bodyText, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return "", err
	}

	var jsonBody struct {
		Message   string `json:"message,omitempty"`
		PaymentId string `json:"payment_id,omitempty"`
	}

	err = json.Unmarshal(bodyText, &jsonBody)

	if err != nil {
		return "", fmt.Errorf("an error occured while unmarshalling body [Payment ID (Response Auth)], error: %v", err)
	}

	if jsonBody.Message == "Authentication required" {
		in.PaymentId = jsonBody.PaymentId
		return "auth", nil

	} else if !ok(resp.StatusCode) {
		return jsonBody.Message, nil

	} else {
		return "success", nil
	}
}
