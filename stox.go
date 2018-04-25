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

	// Results are returned oldest first
	// For `today` grab the last item
	qq := res[len(res) - 1]

	// Previous close can default to current open
	pc := qq.Open
	if (len(res) > 1) {
		pc = res[len(res) - 2].Close
	}

	// Find pct gain/loss from open
	cng := qq.Close - pc
	cngs := fmt.Sprintf("-$%.2f", math.Abs(cng))
	if cng >= 0 {
		cngs = fmt.Sprintf("+$%.2f", cng)
	}

	pct := (cng/pc) * 100
	pcts := fmt.Sprintf("%.2f%%", pct)

	return fmt.Sprintf("%s: Cur: $%.2f: %s (%s)\nOpen: $%.2f, Low: $%.2f, High: $%.2f",
		sym, qq.Close, cngs, pcts, qq.Open, qq.Low, qq.High), nil

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

