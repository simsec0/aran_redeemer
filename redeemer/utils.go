package redeemer

import (
	"encoding/base64"
	"fmt"
	"io"
	shttp "net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	http "github.com/bogdanfinn/fhttp"
	tls_client "github.com/bogdanfinn/tls-client"
)

type Instance struct {
	Client                            tls_client.HttpClient
	UserAgent                         string
	ChromeVersion                     string
	Token                             string
	Vcc                               string
	Proxy                             string
	PromoCode                         string
	XSuperProp                        string
	StripeKey                         string
	Muid                              string
	Guid                              string
	Sid                               string
	PaymentUserAgent                  string
	ClientSecret                      string
	ConfirmToken                      string
	BillingToken                      string
	BillingInfo                       map[string]map[string]string
	PaymentId                         string
	PaymentSourceId                   string
	StripePaymentIntentClientSecret   string
	StripePaymentIntentClientSecretId string
	ThreeDSecure2Source               string
	XFingerprint                      string
	ServerTranscation                 string
	Merchant                          string
	ThreeDsMethodURL                  string
}

// Checks if status code is in {200, 201, 204}
func ok(statusCode int) bool {
	for _, value := range [3]int{200, 201, 204} {
		if statusCode == value {
			return true
		}
	}

	return false
}

func ParseData(content string) string {
	return strings.ReplaceAll(content, " ", "+")
}

func DiscordHeaders(useCommonHeaders bool, headers map[string]string, req *http.Request) *http.Request {
	if useCommonHeaders {
		for k, v := range map[string]string{
			"authority":                 "discord.com",
			"accept":                    "*/*",
			"accept-language":           "en-US,en;q=0.9",
			"dnt":                       "1",
			"sec-ch-ua":                 `"Chromium";v="104", " Not A;Brand";v="99", "Google Chrome";v="104"`,
			"sec-ch-ua-mobile":          "?0",
			"sec-ch-ua-platform":        `"Windows"`,
			"sec-fetch-dest":            "document",
			"sec-fetch-mode":            "navigate",
			"sec-fetch-site":            "none",
			"sec-fetch-user":            "?1",
			"upgrade-insecure-requests": "1",
			"user-agent":                "Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/104.0.0.0 Safari/537.36",
		} {
			req.Header.Set(k, v)
		}
	}

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	return req
}

func StripeHeaders(useCommonHeaders bool, headers map[string]string, req *http.Request) *http.Request {
	if useCommonHeaders {
		for k, v := range map[string]string{
			"accept":             "application/json",
			"accept-language":    "en-US,en;q=0.9",
			"content-type":       "application/x-www-form-urlencoded",
			"dnt":                "1",
			"origin":             "https://m.stripe.network",
			"referer":            "https://m.stripe.network/",
			"sec-ch-ua":          `".Not/A)Brand";v="99", "Google Chrome";v="103", "Chromium";v="103"`,
			"sec-ch-ua-mobile":   "?0",
			"sec-ch-ua-platform": `"Windows"`,
			"sec-fetch-dest":     "empty",
			"sec-fetch-mode":     "cors",
			"sec-fetch-site":     "cross-site",
			"user-agent":         "Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/104.0.0.0 Safari/537.36",
		} {
			req.Header.Set(k, v)
		}
	}

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	return req
}

var (
	buildNumber string
)

func UpdateDiscordBuildInfo() error {
	jsFileRegex := regexp.MustCompile(`([a-zA-z0-9]+)\.js`)
	buildInfoRegex := regexp.MustCompile(`Build Number: [0-9]+, Version Hash: [A-Za-z0-9]+`)
	req, err := shttp.NewRequest("GET", "https://discord.com/app", nil)
	if err != nil {
		return err
	}
	req.Header = shttp.Header{
		"accept":             {`application/json, text/plain, */*`},
		"accept-language":    {`en-US,en;q=0.9`},
		"cache-control":      {`no-cache`},
		"origin":             {`https://discord.com`},
		"pragma":             {`no-cache`},
		"referer":            {`https://discord.com/`},
		"sec-ch-ua":          {`".Not/A)Brand";v="99", "Google Chrome";v="103", "Chromium";v="103"`},
		"sec-ch-ua-mobile":   {`?0`},
		"sec-ch-ua-platform": {`"Windows"`},
		"sec-fetch-dest":     {`empty`},
		"sec-fetch-mode":     {`cors`},
		"sec-fetch-site":     {`same-site`},
		"user-agent":         {`Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/103.0.0.0 Safari/537.36`},
	}
	client := &shttp.Client{
		Timeout:   10 * time.Second,
		Transport: shttp.DefaultTransport,
	}
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}
	index := 0
	if strings.Contains(string(body), "/alpha/invisible.js") {
		index = 2
	} else {
		index = 1
	}

	r := jsFileRegex.FindAllString(string(body), -1)
	asset := r[len(r)-index]
	resp, err := client.Get("https://discord.com/assets/" + asset)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	z := buildInfoRegex.FindAllString(string(b), -1)
	e := strings.ReplaceAll(z[0], " ", "")
	buildInfos := strings.Split(e, ",")
	buildNum := strings.Split(buildInfos[0], ":")
	buildNumber = buildNum[len(buildNum)-1]

	fmt.Println("Fetched Latest Build Info")

	return nil
}

func GetDiscordBuildNumber() string {
	return buildNumber
}

func BuildSuperProps(userAgent string, browserVersion string) string {
	err := UpdateDiscordBuildInfo()

	if err != nil {
		return "eyJvcyI6IldpbmRvd3MiLCJicm93c2VyIjoiQ2hyb21lIiwiZGV2aWNlIjoiIiwic3lzdGVtX2xvY2FsZSI6ImVuLVVTIiwiYnJvd3Nlcl91c2VyX2FnZW50IjoiTW96aWxsYS81LjAgKFdpbmRvd3MgTlQgMTAuMDsgV09XNjQpIEFwcGxlV2ViS2l0LzUzNy4zNiAoS0hUTUwsIGxpa2UgR2Vja28pIENocm9tZS8xMDUuMC4wLjAgU2FmYXJpLzUzNy4zNiIsImJyb3dzZXJfdmVyc2lvbiI6IjEwNS4wLjAuMCIsIm9zX3ZlcnNpb24iOiIxMCIsInJlZmVycmVyIjoiIiwicmVmZXJyaW5nX2RvbWFpbiI6IiIsInJlZmVycmVyX2N1cnJlbnQiOiIiLCJyZWZlcnJpbmdfZG9tYWluX2N1cnJlbnQiOiIiLCJyZWxlYXNlX2NoYW5uZWwiOiJzdGFibGUiLCJjbGllbnRfYnVpbGRfbnVtYmVyIjowLCJjbGllbnRfZXZlbnRfc291cmNlIg=="
	}

	buildNum, _ := strconv.Atoi(GetDiscordBuildNumber())

	toEncode := fmt.Sprintf(`{"os":"Windows","browser":"Chrome","device":"","system_locale":"en-US","browser_user_agent":"%v","browser_version":"%v","os_version":"10","referrer":"","referring_domain":"","referrer_current":"","referring_domain_current":"","release_channel":"stable","client_build_number":%d,"client_event_source":null}`, userAgent, browserVersion, buildNum)

	return base64.StdEncoding.EncodeToString([]byte(toEncode))
}
