package stox

import (
	"fmt"
	"strings"
	"context"
	"math"

	log "github.com/Sirupsen/logrus"
	"github.com/jirwin/quadlek/quadlek"
	"github.com/cmckee-dev/go-alpha-vantage"
)

var client *av.Client = nil

func isAllUC(s string) bool {
	up := strings.ToUpper(s)
	return strings.Compare(s, up) == 0
}

func getQuote(sym string) (string, error) {
	res, err := client.StockTimeSeries(av.TimeSeriesDaily, sym)
	if err != nil {
		return "", err
	}

	// Find pct gain/loss from open
	cng := res[0].Close - res[0].Open
	cngs := fmt.Sprintf("-$%.2f", math.Abs(cng))
	if cng >= 0 {
		cngs = fmt.Sprintf("+$%.2f", cng)
	}

	pct := math.RoundToEven(((res[0].Close/res[0].Open)-1.0) * 100)
	pcts := fmt.Sprintf("%.2f%%", pct)

	return fmt.Sprintf("%s: Cur: $%.2f: %s (%s)\nOpen: $%.2f, Low: $%.2f, High: $%.2f",
		sym, res[0].Close, cngs, pcts, res[0].Open, res[0].Low, res[0].High), nil

}

func stoxHook(ctx context.Context, hookChan <-chan *quadlek.HookMsg) {
	for {
		select {
		case hook := <-hookChan:
			tokens := strings.Split(hook.Msg.Text, " ")

			for _, t := range tokens {
				if strings.HasPrefix(t, "$") {
					symbol := t[1:]
					if isAllUC(symbol) {
						log.Info(fmt.Sprintf("Symobl lookup triggered: %s", symbol))
						quote, err := getQuote(symbol)
						if err != nil {
							hook.Bot.Say(hook.Msg.Channel, fmt.Sprintf("Could not fetch quote for: %s", symbol))
						} else {
							hook.Bot.Say(hook.Msg.Channel, quote)
						}
					}
				}
			}

		case <-ctx.Done():
			log.Info("Exiting Stox")
		return
		}
	}
}

func Register(apiKey string) quadlek.Plugin {
	client = av.NewClient(apiKey)

	return quadlek.MakePlugin(
		"stox",
		nil,
		[]quadlek.Hook{
			quadlek.MakeHook(stoxHook),
		},
		nil,
		nil,
		nil,
	)
}

