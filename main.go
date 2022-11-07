package main

import (
	auth "NitroRedeemerV3/auth"
	redeemer "NitroRedeemerV3/redeemer"
	"encoding/json"
	"fmt"

	"github.com/its-vichy/GoCycle"

	modules "NitroRedeemerV3/modules"

	"os"
	"strings"
	"sync"
	"time"

	"github.com/bogdanfinn/fhttp/cookiejar"
	tls_client "github.com/bogdanfinn/tls-client"
)

var (
	name      = "Nitro Redeemer V3"
	ownerid   = "LCDmGSlZxE"
	version   = "1.0"
	key       = ""
	successes int
	proxies   *GoCycle.Cycle
	config    Config
	tokens    []string
	vccs      []string
	promos    []string
	wg        sync.WaitGroup
)

type Config struct {
	License       string `json:"license,omitempty"`
	UseOnVcc      int    `json:"use_on_vcc,omitempty"`
	Proxyless     bool   `json:"proxyless,omitempty"`
	AmtOfWorkers  int    `json:"amt_of_workers,omitempty"`
	ClientTimeout int    `json:"client_timeout,omitempty"`
	Fingerprint   struct {
		UseClientProfile bool   `json:"use_client_profile,omitempty"`
		UserAgent        string `json:"user_agent,omitempty"`
		BrowserVersion   string `json:"browser_version,omitempty"`
		Ja3              string `json:"ja3,omitempty"`
	} `json:"fingerprint,omitempty"`
}

func createIterator(file string) (*GoCycle.Cycle, error) {
	iterator, err := GoCycle.NewFromFile(file)

	if err != nil {
		return nil, err
	}

	return iterator, nil
}

func loadConfig() (Config, error) {
	var config Config
	configFile, err := os.Open("config.json")
	if err != nil {
		return config, err
	}
	defer configFile.Close()
	jsonParser := json.NewDecoder(configFile)
	jsonParser.Decode(&config)
	return config, nil
}

func init() {
	modules.Clear()
	load_config, err := loadConfig()
	if err != nil {
		panic(fmt.Sprintf("an error occured while opening config.json, error: %v", err.Error()))
	}

	config = load_config

	auth.Api(name, ownerid, version)
	auth.Init()
	auth.License(config.License)

	if !config.Proxyless {
		proxyIterator, err := createIterator("data/proxies.txt")

		if err != nil {
			panic(fmt.Sprintf("Could not create a proxy cycle, error: %v", err.Error()))
		}

		proxies = proxyIterator
	}

	all_vccs, err := modules.ReadLines("data/vccs.txt")

	if err != nil {
		panic(fmt.Sprintf("Could not read vccs.txt, error: %v", err.Error()))
	}

	if len(all_vccs) <= 0 {
		panic("there are no vccs in vccs.txt")
	} else {
		for i := 0; i < config.UseOnVcc; i++ {
			vccs = append(vccs, all_vccs...)
		}
	}

	all_tokens, err := modules.ReadLines("data/tokens.txt")

	if err != nil {
		panic(fmt.Sprintf("Could not read tokens.txt, error: %v", err.Error()))
	} else if len(all_tokens) <= 0 {
		panic("there are no tokens in tokens.txt")
	}

	all_promos, err := modules.ReadLines("data/promos.txt")

	if err != nil {
		panic(fmt.Sprintf("Could not read promos.txt, error: %v", err.Error()))
	}

	promos = all_promos
	tokens = all_tokens
}

func createThread(promoCode, token, vcc, proxy string) string {
	defer wg.Done()
	// Checking for bad variables, if one is found it will panic
	if !strings.Contains(config.Fingerprint.UserAgent, "Chrome") {
		panic(fmt.Sprintf("Currently useragent %v is not supported, please use a chrome user agent version 103-105", config.Fingerprint.UserAgent))
	} else if !config.Fingerprint.UseClientProfile && config.Fingerprint.Ja3 == "" {
		panic("You have set UseClientProfile to false but no Ja3 was provided")
	}

	jar, _ := cookiejar.New(nil)

	options := []tls_client.HttpClientOption{
		tls_client.WithTimeout(config.ClientTimeout),

		func() tls_client.HttpClientOption {
			if config.Fingerprint.UseClientProfile {
				if strings.Contains(config.Fingerprint.BrowserVersion, "103") {
					return tls_client.WithClientProfile(tls_client.Chrome_103)

				} else if strings.Contains(config.Fingerprint.BrowserVersion, "104") {
					return tls_client.WithClientProfile(tls_client.Chrome_104)

				} else if strings.Contains(config.Fingerprint.BrowserVersion, "105") {
					return tls_client.WithClientProfile(tls_client.Chrome_105)

				} else {
					panic(fmt.Sprintf("Currently version %v is not supported, please use a chrome useragent/version", config.Fingerprint.UseClientProfile))
				}
			} else {
				return tls_client.WithJa3String(config.Fingerprint.Ja3)
			}
		}(),

		func() tls_client.HttpClientOption {
			if proxy != "" {
				return tls_client.WithProxyUrl(fmt.Sprintf("https://%v", proxy))
			} else {
				return nil
			}
		}(),
		tls_client.WithNotFollowRedirects(),
		tls_client.WithInsecureSkipVerify(),
		tls_client.WithCookieJar(jar),
	}

	client, err := tls_client.NewHttpClient(tls_client.NewNoopLogger(), options...)

	if err != nil {
		modules.Logger(false, "Could not create tls client, err:", err.Error())
		return "err"
	}

	in := redeemer.Instance{
		Client:           client,
		PromoCode:        promoCode,
		Vcc:              vcc,
		PaymentUserAgent: `stripe.js%2Ff2ecd562b%3B+stripe-js-v3%2Ff2ecd562b`,
		Token:            token,
		UserAgent:        config.Fingerprint.UserAgent,
		ChromeVersion:    config.Fingerprint.BrowserVersion,
		BillingInfo: map[string]map[string]string{
			"billing_address": {
				"name":        "Joseph Wong",
				"line_1":      "605 Shannon Rd",
				"line_2":      "",
				"city":        "Papillion",
				"state":       "NE",
				"postal_code": "68046",
				"country":     "US",
				"email":       "",
			},
		},
	}
	err = in.Session()

	if err != nil {
		modules.Logger(false, "An error occured:", err.Error())
		return "err"
	}

	err = in.Stripe()

	if err != nil {
		modules.Logger(false, "An error occured:", err.Error())
		return "errr"
	}

	err = in.StripeToken()

	if err != nil {
		modules.Logger(false, "An error occured:", err.Error())
		return "err"
	}

	err = in.SetupIntents()

	if err != nil {
		modules.Logger(false, "An error occured:", err.Error())
		return "err"
	}

	err = in.ValidateBilling()

	if err != nil {
		modules.Logger(false, "An error occured:", err.Error())
		return "err"
	}

	err = in.ConfirmStripe()

	if err != nil {
		modules.Logger(false, "An error occured:", err.Error())
		return "err"
	}

	err = in.AddPaymentMethod()

	if err != nil {
		modules.Logger(false, "An error occured:", err.Error())
		return "err"
	}

	resp, err := in.Redeem()

	if err != nil {
		modules.Logger(false, "An error occured:", err.Error())
		return "err"
	}

	if resp == "success" {
		modules.AppendLine("data/success.txt", in.Token)
		return "success"

	} else if resp != "auth" {
		modules.Logger(false, fmt.Sprintf("An error occured while redeeming, error: %v", resp))
		return "err"

	} else {
		err = in.DiscordPaymentIntents()

		if err != nil {
			modules.Logger(false, "An error occured:", err.Error())
			return "err"
		}

		time.Sleep(503 * time.Millisecond)

		err = in.StripePaymentIntents()

		if err != nil {
			modules.Logger(false, "An error occured:", err.Error())
			return "err"
		}

		time.Sleep(503 * time.Millisecond)

		err = in.ConfirmPaymentIntents()

		if err != nil {
			modules.Logger(false, "An error occured:", err.Error())
			return "err"
		}

		time.Sleep(503 * time.Millisecond)

		err = in.Authentication()

		if err != nil {
			modules.Logger(false, "An error occured:", err.Error())
			return "err"
		}

		time.Sleep(503 * time.Millisecond)

		err = in.StripePaymentIntents()

		if err != nil {
			modules.Logger(false, "An error occured:", err.Error())
			return "err"
		}

		time.Sleep(503 * time.Millisecond)

		resp, err := in.Redeem()

		if resp == "success" && err == nil {
			modules.AppendLine("data/success.txt", in.Token)
			modules.Logger(true, "Successfully redeemed nitro on ->", in.Token)
			return "success"
		} else {
			modules.Logger(false, `Goroutine ended with message:`, resp, `| error:`, err)
			return "err"
		}

	}
}

func main() {
	tokensLength, promoLength, vccLength := len(tokens), len(promos), len(vccs)/config.UseOnVcc

	go func() {
		for {
			modules.SetTitle(fmt.Sprintf("Opti Boosts ^| Tokens: %v ^| Promos: %v ^| Vccs: %v ^| Success: %v", tokensLength, promoLength, vccLength, successes))
		}
	}()

	for tokensLength >= 1 && promoLength >= 1 && vccLength >= 1 {
		wg.Add(config.AmtOfWorkers)

		for i := 0; i < config.AmtOfWorkers; i++ {
			if tokensLength <= 0 || promoLength <= 0 || vccLength <= 0 {
				modules.Logger(false, "Ran out of materials to use!")
				break
			}

			var proxy string
			if !config.Proxyless {
				proxy = proxies.Next()
			} else {
				proxy = ""
			}
			fmt.Println(promos)
			token, promo, vcc := tokens[0], promos[0], vccs[0]
			tokens, promos, vccs = modules.RemoveFromArray(tokens, token), modules.RemoveFromArray(promos, promo), modules.RemoveFromArray(vccs, vcc)

			tokensLength--
			promoLength--

			if !modules.Contains(vccs, vcc) {
				vccLength--
			}

			go createThread(promo, token, vcc, proxy)
		}

		wg.Wait()
	}
}
