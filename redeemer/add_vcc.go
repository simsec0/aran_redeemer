package redeemer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"strings"

	http "github.com/bogdanfinn/fhttp"
)

func (in *Instance) Session() error {
	req, err := http.NewRequest("GET", in.PromoCode, nil)

	if err != nil {
		return err
	}

	req = DiscordHeaders(true, nil, req)

	resp, err := in.Client.Do(req)

	if err != nil {
		return err
	}

	if !ok(resp.StatusCode) {
		return fmt.Errorf("an error occured while getting cookies, Status code: %d", resp.StatusCode)
	}

	defer resp.Body.Close()
	bodyText, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return err
	}

	key := strings.Split(string(bodyText), `STRIPE_KEY: '`)[1]
	in.StripeKey = strings.Split(key, `',`)[0]

	in.XSuperProp = BuildSuperProps(in.UserAgent, in.ChromeVersion)

	XFingerprint, err := func() (string, error) {
		req, err := http.NewRequest("GET", "https://discord.com/api/v9/experiments", nil)

		if err != nil {
			return "", err
		}

		req = DiscordHeaders(true, map[string]string{
			"X-Context-Properties": "eyJsb2NhdGlvbiI6IlJlZ2lzdGVyIn0=",
			"X-Debug-Options":      "bugReporterEnabled",
			"X-Discord-Locale":     "en-US",
			"X-Super-Properties":   in.XSuperProp,
			"Host":                 "discord.com",
			"Referer":              in.PromoCode,
		}, req)

		resp, err := in.Client.Do(req)

		if err != nil {
			return "", err
		} else if !ok(resp.StatusCode) {
			return "", fmt.Errorf("an error occured while getting x-fingerprint, Status code: %d", resp.StatusCode)
		}

		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)

		if err != nil {
			return "", err
		}

		var jsonBody struct {
			Fingerprint string `json:"fingerprint"`
		}

		err = json.Unmarshal(body, &jsonBody)

		if err != nil {
			return "", fmt.Errorf("an error occured while unmarshalling fingerprint body, error: %v", err)
		}

		return jsonBody.Fingerprint, nil

	}()

	if err != nil {
		return err
	}

	in.XFingerprint = XFingerprint

	return nil
}

func (in *Instance) Stripe() error {
	var data = strings.NewReader(`JTdCJTIydjIlMjIlM0ExJTJDJTIyaWQlMjIlM0ElMjJmYjRmNGIwOWEwOWY1YzJlODJiNDU5ZjQwMmMwMDFjMCUyMiUyQyUyMnQlMjIlM0E0MjEuMyUyQyUyMnRhZyUyMiUzQSUyMjQuNS40MiUyMiUyQyUyMnNyYyUyMiUzQSUyMmpzJTIyJTJDJTIyYSUyMiUzQSU3QiUyMmElMjIlM0ElN0IlMjJ2JTIyJTNBJTIyZmFsc2UlMjIlMkMlMjJ0JTIyJTNBMC4zJTdEJTJDJTIyYiUyMiUzQSU3QiUyMnYlMjIlM0ElMjJ0cnVlJTIyJTJDJTIydCUyMiUzQTAlN0QlMkMlMjJjJTIyJTNBJTdCJTIydiUyMiUzQSUyMmVuLUNBJTJDZW4tR0IlMkNlbi1VUyUyQ2VuJTIyJTJDJTIydCUyMiUzQTAlN0QlMkMlMjJkJTIyJTNBJTdCJTIydiUyMiUzQSUyMldpbjMyJTIyJTJDJTIydCUyMiUzQTAuMSU3RCUyQyUyMmUlMjIlM0ElN0IlMjJ2JTIyJTNBJTIyUERGJTIwVmlld2VyJTJDaW50ZXJuYWwtcGRmLXZpZXdlciUyQ2FwcGxpY2F0aW9uJTJGcGRmJTJDcGRmJTJCJTJCdGV4dCUyRnBkZiUyQ3BkZiUyQyUyMENocm9tZSUyMFBERiUyMFZpZXdlciUyQ2ludGVybmFsLXBkZi12aWV3ZXIlMkNhcHBsaWNhdGlvbiUyRnBkZiUyQ3BkZiUyQiUyQnRleHQlMkZwZGYlMkNwZGYlMkMlMjBDaHJvbWl1bSUyMFBERiUyMFZpZXdlciUyQ2ludGVybmFsLXBkZi12aWV3ZXIlMkNhcHBsaWNhdGlvbiUyRnBkZiUyQ3BkZiUyQiUyQnRleHQlMkZwZGYlMkNwZGYlMkMlMjBNaWNyb3NvZnQlMjBFZGdlJTIwUERGJTIwVmlld2VyJTJDaW50ZXJuYWwtcGRmLXZpZXdlciUyQ2FwcGxpY2F0aW9uJTJGcGRmJTJDcGRmJTJCJTJCdGV4dCUyRnBkZiUyQ3BkZiUyQyUyMFdlYktpdCUyMGJ1aWx0LWluJTIwUERGJTJDaW50ZXJuYWwtcGRmLXZpZXdlciUyQ2FwcGxpY2F0aW9uJTJGcGRmJTJDcGRmJTJCJTJCdGV4dCUyRnBkZiUyQ3BkZiUyMiUyQyUyMnQlMjIlM0ExLjklN0QlMkMlMjJmJTIyJTNBJTdCJTIydiUyMiUzQSUyMjE5MjB3XzEwNDBoXzI0ZF8xciUyMiUyQyUyMnQlMjIlM0EwJTdEJTJDJTIyZyUyMiUzQSU3QiUyMnYlMjIlM0ElMjItNCUyMiUyQyUyMnQlMjIlM0EwJTdEJTJDJTIyaCUyMiUzQSU3QiUyMnYlMjIlM0ElMjJmYWxzZSUyMiUyQyUyMnQlMjIlM0EwJTdEJTJDJTIyaSUyMiUzQSU3QiUyMnYlMjIlM0ElMjJzZXNzaW9uU3RvcmFnZS1kaXNhYmxlZCUyQyUyMGxvY2FsU3RvcmFnZS1kaXNhYmxlZCUyMiUyQyUyMnQlMjIlM0EwLjQlN0QlMkMlMjJqJTIyJTNBJTdCJTIydiUyMiUzQSUyMjAxMDAxMDAxMDExMTExMTExMDAxMTExMDExMTExMTExMDExMTAwMTAxMTAxMTExMTAxMTExMTElMjIlMkMlMjJ0JTIyJTNBNDE4LjMlMkMlMjJhdCUyMiUzQTIyNC42JTdEJTJDJTIyayUyMiUzQSU3QiUyMnYlMjIlM0ElMjIlMjIlMkMlMjJ0JTIyJTNBMC4xJTdEJTJDJTIybCUyMiUzQSU3QiUyMnYlMjIlM0ElMjJNb3ppbGxhJTJGNS4wJTIwKFdpbmRvd3MlMjBOVCUyMDEwLjAlM0IlMjBXT1c2NCklMjBBcHBsZVdlYktpdCUyRjUzNy4zNiUyMChLSFRNTCUyQyUyMGxpa2UlMjBHZWNrbyklMjBDaHJvbWUlMkYxMDQuMC4wLjAlMjBTYWZhcmklMkY1MzcuMzYlMjIlMkMlMjJ0JTIyJTNBMCU3RCUyQyUyMm0lMjIlM0ElN0IlMjJ2JTIyJTNBJTIyJTIyJTJDJTIydCUyMiUzQTAuMSU3RCUyQyUyMm4lMjIlM0ElN0IlMjJ2JTIyJTNBJTIyZmFsc2UlMjIlMkMlMjJ0JTIyJTNBMTQzLjIlMkMlMjJhdCUyMiUzQTEuNCU3RCUyQyUyMm8lMjIlM0ElN0IlMjJ2JTIyJTNBJTIyMGJlYTg1MGZmYjliM2FhZGMwZTM4MTFmYjk5NjI4ZjYlMjIlMkMlMjJ0JTIyJTNBNzIuMSU3RCU3RCUyQyUyMmIlMjIlM0ElN0IlMjJhJTIyJTNBJTIyJTIyJTJDJTIyYiUyMiUzQSUyMmh0dHBzJTNBJTJGJTJGR1M3aHFua1pCUnBQXzdXbjktQ0dGaHpxNGtyMlgzSkM0QTNrNkJEQnZwRS5nMnU5LWhxWnZHSXFZSmNQbFBmd0pBZi12M1JneUtfeDFOcHB6QWxBMTJNJTJGQlFkTnl6cExVNG5NNllLenpuYVAxWENEMVcwREoxejNUcG50ejJacElxcyUyRjBzTHkxUFBJSTJobTNPREhpTFp1S2M2UmR5Y0xFMVZybXJ5bnRzWFh0N28lMkZOUGNxbjRDZnRMa2kyX3lBOUVKa0Vwb21sUmxJR1NvT2xuRXRXV3hOODEwJTIyJTJDJTIyYyUyMiUzQSUyMl9JbDFfZzZUOXNyNVdxLXR5SGRlTDFlZUV0ejdPN0lETzFnckMtTlpjVWslMjIlMkMlMjJkJTIyJTNBJTIyTkElMjIlMkMlMjJlJTIyJTNBJTIyTkElMjIlMkMlMjJmJTIyJTNBZmFsc2UlMkMlMjJnJTIyJTNBdHJ1ZSUyQyUyMmglMjIlM0F0cnVlJTJDJTIyaSUyMiUzQSU1QiUyMmxvY2F0aW9uJTIyJTVEJTJDJTIyaiUyMiUzQSU1QiU1RCUyQyUyMm4lMjIlM0E0MzAuMjAwMDAwMDQ3NjgzNyUyQyUyMnUlMjIlM0ElMjJkaXNjb3JkLmNvbSUyMiU3RCUyQyUyMmglMjIlM0ElMjJhNjkzYTE1NmM5Y2I2ZTM4OWMxNyUyMiU3RA==`)

	req, err := http.NewRequest("POST", "https://m.stripe.com/6", data)

	if err != nil {
		return err
	}

	req = StripeHeaders(true, nil, req)

	resp, err := in.Client.Do(req)

	if err != nil {
		return err
	}

	if !ok(resp.StatusCode) {
		return fmt.Errorf("an error occured while getting uuid, Status code: %d", resp.StatusCode)
	}

	defer resp.Body.Close()
	bodyText, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return err
	}

	var jsonBody struct {
		Muid string `json:"muid,omitempty"`
		Guid string `json:"guid,omitempty"`
		Sid  string `json:"sid,omitempty"`
	}

	err = json.Unmarshal(bodyText, &jsonBody)

	if err != nil {
		return fmt.Errorf("an error occured while unmarshalling body, error: %v", err)
	}

	in.Muid = jsonBody.Muid
	in.Guid = jsonBody.Guid
	in.Sid = jsonBody.Sid

	return nil
}

func (in *Instance) StripeToken() error {
	vcc := strings.Split(in.Vcc, ":")
	parsedData := fmt.Sprintf(`card[number]=%v&card[cvc]=%v&card[exp_month]=%v&card[exp_year]=%v&guid=%v&muid=%v&sid=%v&payment_user_agent=%v&time_on_page=%v&key=%v&pasted_fields=%s`, vcc[0], vcc[2], vcc[1][0:2], vcc[1][2:], in.Guid, in.Muid, in.Sid, in.PaymentUserAgent, rand.Intn(60000)+60000, in.StripeKey, "number%2Cexp%2Ccvc")
	var data = strings.NewReader(parsedData)

	req, err := http.NewRequest("POST", "https://api.stripe.com/v1/tokens", data)

	if err != nil {
		return err
	}

	req = StripeHeaders(true, map[string]string{
		"origin":  "https://js.stripe.com",
		"referer": "https://js.stripe.com/",
	}, req)

	resp, err := in.Client.Do(req)

	if err != nil {
		return err
	}

	if !ok(resp.StatusCode) {
		return fmt.Errorf("an error occured while getting Stripe Token [Confirm Token], Status code: %d", resp.StatusCode)
	}

	defer resp.Body.Close()
	bodyText, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return err
	}

	var jsonBody struct {
		ConfirmToken string `json:"id,omitempty"`
	}

	err = json.Unmarshal(bodyText, &jsonBody)

	if err != nil {
		return fmt.Errorf("an error occured while unmarshalling body [Confirm Token], error: %v", err)
	}

	in.ConfirmToken = jsonBody.ConfirmToken

	return nil
}

func (in *Instance) SetupIntents() error {
	req, err := http.NewRequest("POST", "https://discord.com/api/v9/users/@me/billing/stripe/setup-intents", nil)

	if err != nil {
		return err
	}

	req = DiscordHeaders(true, map[string]string{
		"X-Debug-Options":    "bugReporterEnabled",
		"X-Discord-Locale":   "en-US",
		"X-Super-Properties": in.XSuperProp,
		"Host":               "discord.com",
		"Referer":            in.PromoCode,
		"Authorization":      in.Token,
	}, req)

	resp, err := in.Client.Do(req)

	if err != nil {
		return err
	}

	if !ok(resp.StatusCode) {
		return fmt.Errorf("an error occured while setting up Discord Intents [Client Secret], Status code: %d", resp.StatusCode)
	}

	defer resp.Body.Close()
	bodyText, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return err
	}

	var jsonBody struct {
		ClientSecret string `json:"client_secret,omitempty"`
	}

	err = json.Unmarshal(bodyText, &jsonBody)

	if err != nil {
		return fmt.Errorf("an error occured while unmarshalling body [Client Secret], error: %v", err)
	}

	in.ClientSecret = jsonBody.ClientSecret

	return nil
}

func (in *Instance) ValidateBilling() error {
	payload, err := json.Marshal(in.BillingInfo)

	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", "https://discord.com/api/v9/users/@me/billing/payment-sources/validate-billing-address", bytes.NewBuffer(payload))

	if err != nil {
		return err
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
		return err
	}

	defer resp.Body.Close()
	bodyText, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return err
	}

	if !ok(resp.StatusCode) {
		return fmt.Errorf("an error occured while validating billing address [Billing Token], Status code: %d", resp.StatusCode)
	}

	var jsonBody struct {
		BillingToken string `json:"token,omitempty"`
	}

	err = json.Unmarshal(bodyText, &jsonBody)

	if err != nil {
		return fmt.Errorf("an error occured while unmarshalling [Billing Token], error: %v", err)
	}

	in.BillingToken = jsonBody.BillingToken

	return nil
}

func (in *Instance) ConfirmStripe() error {
	depractedClientSecret := strings.Split(in.ClientSecret, "_secret_")[0]
	parsedData := ParseData(fmt.Sprintf(`payment_method_data[type]=card&payment_method_data[card][token]=%v&payment_method_data[billing_details][address][line1]=%v&payment_method_data[billing_details][address][line2]=%v&payment_method_data[billing_details][address][city]=%v&payment_method_data[billing_details][address][state]=%v&payment_method_data[billing_details][address][postal_code]=%v&payment_method_data[billing_details][address][country]=%v&payment_method_data[billing_details][name]=%v&payment_method_data[guid]=%v&payment_method_data[muid]=%v&payment_method_data[sid]=%v&payment_method_data[payment_user_agent]=%v&payment_method_data[time_on_page]=%v&expected_payment_method_type=card&use_stripe_sdk=true&key=%v&client_secret=%v`, in.ConfirmToken, in.BillingInfo["billing_address"]["line_1"], in.BillingInfo["billing_address"]["line_2"], in.BillingInfo["billing_address"]["city"], in.BillingInfo["billing_address"]["state"], in.BillingInfo["billing_address"]["postal_code"], in.BillingInfo["billing_address"]["country"], in.BillingInfo["billing_address"]["name"], in.Guid, in.Muid, in.Sid, in.PaymentUserAgent, rand.Intn(450000)+250000, in.StripeKey, in.ClientSecret))
	var data = strings.NewReader(parsedData)
	link := fmt.Sprintf("https://api.stripe.com/v1/setup_intents/%v/confirm", depractedClientSecret)

	req, err := http.NewRequest("POST", link, data)

	if err != nil {
		return err
	}

	req = StripeHeaders(true, map[string]string{
		"origin":  "https://js.stripe.com",
		"referer": "https://js.stripe.com/",
	}, req)

	resp, err := in.Client.Do(req)

	if err != nil {
		return err
	}

	defer resp.Body.Close()
	bodyText, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return err
	}

	var jsonBody struct {
		PaymentId string `json:"payment_method,omitempty"`
		Error     struct {
			Message string `json:"message"`
		} `json:"error"`
	}

	err = json.Unmarshal(bodyText, &jsonBody)

	if err != nil {
		return fmt.Errorf("an error occured while unmarshalling body [Payment ID], error: %v", err)
	}

	if !ok(resp.StatusCode) {
		return fmt.Errorf("an error occured while setting up stripe intents [Payment ID], Error %v, Status code: %d", jsonBody.Error.Message, resp.StatusCode)
	}

	in.PaymentId = jsonBody.PaymentId

	return nil

}

func (in *Instance) AddPaymentMethod() error {
	jsonData := map[string]interface{}{
		"payment_gateway":       1,
		"token":                 in.PaymentId,
		"billing_address":       in.BillingInfo["billing_address"],
		"billing_address_token": in.BillingToken,
	}

	payload, err := json.Marshal(jsonData)

	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", "https://discord.com/api/v9/users/@me/billing/payment-sources", bytes.NewBuffer(payload))

	if err != nil {
		return err
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
		return err
	}

	defer resp.Body.Close()
	bodyText, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return err
	}

	if !ok(resp.StatusCode) {
		return fmt.Errorf("an error occured while adding the payment method [Payment Source ID], Status code: %d", resp.StatusCode)
	}

	var jsonBody struct {
		PaymentSourceId string `json:"id,omitempty"`
	}

	err = json.Unmarshal(bodyText, &jsonBody)

	if err != nil {
		return fmt.Errorf("an error occured while unmarshalling body [Payment Source ID], error: %v", err)
	}

	in.PaymentSourceId = jsonBody.PaymentSourceId

	return nil

}
