package redeemer

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/url"
	"strings"

	http "github.com/bogdanfinn/fhttp"
)

func (in *Instance) DiscordPaymentIntents() error {
	link := fmt.Sprintf("https://discord.com/api/v9/users/@me/billing/stripe/payment-intents/payments/%v", in.PaymentId)
	req, err := http.NewRequest("GET", link, nil)

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

	if !ok(resp.StatusCode) {
		return fmt.Errorf("an error occured while getting [StripePaymentIntentClientSecret/Id], Status code: %d", resp.StatusCode)
	}

	defer resp.Body.Close()
	bodyText, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return err
	}

	var jsonBody struct {
		StripePaymentIntentClientSecret   string `json:"stripe_payment_intent_client_secret,omitempty"`
		StripePaymentIntentClientSecretId string `json:"stripe_payment_intent_payment_method_id,omitempty"`
	}

	err = json.Unmarshal(bodyText, &jsonBody)

	if err != nil {
		return fmt.Errorf("an error occured while unmarshalling body [StripePaymentIntentClientSecret/Id], error: %v", err)
	}

	in.StripePaymentIntentClientSecret = jsonBody.StripePaymentIntentClientSecret
	in.StripePaymentIntentClientSecretId = jsonBody.StripePaymentIntentClientSecretId

	return nil
}

// Possibly Useless request
func (in *Instance) StripePaymentIntents() error {
	link := fmt.Sprintf("https://api.stripe.com/v1/payment_intents/%v?key=%v&is_stripe_sdk=false&client_secret=%v", strings.Split(in.StripePaymentIntentClientSecret, "_secret_")[0], in.StripeKey, in.StripePaymentIntentClientSecret)
	req, err := http.NewRequest("GET", link, nil)

	if err != nil {
		return err
	}

	req = StripeHeaders(true, nil, req)

	resp, err := in.Client.Do(req)

	if err != nil {
		return err
	}

	if !ok(resp.StatusCode) {
		return fmt.Errorf("an error occured while getting stripe payment intents, Status code: %d", resp.StatusCode)
	}

	return nil
}

func (in *Instance) ConfirmPaymentIntents() error {
	parsed := fmt.Sprintf(`expected_payment_method_type=card&use_stripe_sdk=true&key=%v&client_secret=%v`, in.StripeKey, in.StripePaymentIntentClientSecret)
	var data = strings.NewReader(parsed)
	link := fmt.Sprintf("https://api.stripe.com/v1/payment_intents/%v/confirm", strings.Split(in.StripePaymentIntentClientSecret, "_secret_")[0])
	req, err := http.NewRequest("POST", link, data)

	if err != nil {
		return err
	}

	req = StripeHeaders(true, nil, req)

	resp, err := in.Client.Do(req)

	if err != nil {
		return err
	}

	if !ok(resp.StatusCode) {
		return fmt.Errorf("an error occured while getting [StripePaymentIntentClientSecret/Id], Status code: %d", resp.StatusCode)
	}

	var jsonBody struct {
		NextAction struct {
			Type         string `json:"type,omitempty"`
			UseStripeSdk struct {
				ThreeDSecure2Source string `json:"three_d_secure_2_source,omitempty"`
				ServerTranscation   string `json:"server_transaction_id,omitempty"`
				Merchant            string `json:"merchant"`
				ThreeDsMethodURL    string `json:"three_ds_method_url"`
			} `json:"use_stripe_sdk,omitempty"`
		} `json:"next_action,omitempty"`
	}

	defer resp.Body.Close()
	bodyText, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return err
	}

	if !ok(resp.StatusCode) {
		return fmt.Errorf("an error occured while adding the payment method [Payment Source ID], Status code: %d", resp.StatusCode)
	}

	err = json.Unmarshal(bodyText, &jsonBody)

	if err != nil {
		return fmt.Errorf("an error occured while unmarshalling body [Payment Source ID], error: %v", err)
	}

	in.ThreeDSecure2Source = jsonBody.NextAction.UseStripeSdk.ThreeDSecure2Source
	in.ServerTranscation = jsonBody.NextAction.UseStripeSdk.ServerTranscation
	in.Merchant = jsonBody.NextAction.UseStripeSdk.Merchant
	in.ThreeDsMethodURL = jsonBody.NextAction.UseStripeSdk.ThreeDsMethodURL

	return nil
}

func (in *Instance) Authentication() error {
	browserScreen := []string{
		"1280x960",
		"1920x1080",
		"1768x992",
		"1680x1050",
		"1600x1024",
		"1600x900",
	}

	randomIndex := rand.Intn(len(browserScreen))
	browserSpecs := browserScreen[randomIndex]

	splitBrowser := strings.Split(browserSpecs, "x")
	browserScreenWidth, browserScreenHeight := splitBrowser[0], splitBrowser[1]

	params := url.Values{}
	params.Add("source", in.ThreeDSecure2Source)
	params.Add("browser", fmt.Sprintf(`{"fingerprintAttempted":false,"fingerprintData":null,"challengeWindowSize":null,"threeDSCompInd":"Y","browserJavaEnabled":false,"browserJavascriptEnabled":true,"browserLanguage":"en-US","browserColorDepth":"24","browserScreenHeight":"%s","browserScreenWidth":"%s","browserTZ":"240","browserUserAgent":"%s"}`, browserScreenHeight, browserScreenWidth, in.UserAgent))
	params.Add("one_click_authn_device_support[hosted]", "false")
	params.Add("one_click_authn_device_support[same_origin_frame]", "false")
	params.Add("one_click_authn_device_support[spc_eligible]", "true")
	params.Add("one_click_authn_device_support[webauthn_eligible]", "true")
	params.Add("one_click_authn_device_support[publickey_credentials_get_allowed]", "true")
	params.Add("key", in.StripeKey)

	var data = strings.NewReader(params.Encode())

	req, err := http.NewRequest("POST", "https://api.stripe.com/v1/3ds2/authenticate", data)

	if err != nil {
		return err
	}

	req = StripeHeaders(true, nil, req)

	resp, err := in.Client.Do(req)

	if err != nil {
		return err
	}

	if !ok(resp.StatusCode) {
		defer resp.Body.Close()
		bodyText, _ := ioutil.ReadAll(resp.Body)
		fmt.Println(string(bodyText))
		return fmt.Errorf("an error occured while getting [Auth], Status code: %d", resp.StatusCode)
	}

	var jsonBody struct {
		State string `json:"state,omitempty"`
		Error string `json:"error,omitempty"`
	}

	defer resp.Body.Close()
	bodyText, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return err
	}

	err = json.Unmarshal(bodyText, &jsonBody)

	if err != nil {
		return fmt.Errorf("an error occured while unmarshalling body [Auth], error: %v", err)
	}

	if jsonBody.State != "succeeded" {
		return fmt.Errorf("an error occured during auth")
	}

	return nil
}
