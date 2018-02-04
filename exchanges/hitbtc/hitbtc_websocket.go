package hitbtc

import (
	"log"
	"strconv"

	"github.com/beatgammit/turnpike"
)

const (
	HITBTC_WEBSOCKET_ADDRESS  = "wss://api.hitbtc.com"
	HITBTC_WEBSOCKET_REALM    = "realm1"
	HITBTC_WEBSOCKET_TICKER   = "ticker"
	HITBTC_WEBSOCKET_TROLLBOX = "trollbox"
)

type HitBTCWebsocketTicker struct {
	CurrencyPair  string
	Last          float64
	LowestAsk     float64
	HighestBid    float64
	PercentChange float64
	BaseVolume    float64
	QuoteVolume   float64
	IsFrozen      bool
	High          float64
	Low           float64
}

func HitBTCOnTicker(args []interface{}, kwargs map[string]interface{}) {
	ticker := HitBTCWebsocketTicker{}
	ticker.CurrencyPair = args[0].(string)
	ticker.Last, _ = strconv.ParseFloat(args[1].(string), 64)
	ticker.LowestAsk, _ = strconv.ParseFloat(args[2].(string), 64)
	ticker.HighestBid, _ = strconv.ParseFloat(args[3].(string), 64)
	ticker.PercentChange, _ = strconv.ParseFloat(args[4].(string), 64)
	ticker.BaseVolume, _ = strconv.ParseFloat(args[5].(string), 64)
	ticker.QuoteVolume, _ = strconv.ParseFloat(args[6].(string), 64)

	if args[7].(float64) != 0 {
		ticker.IsFrozen = true
	} else {
		ticker.IsFrozen = false
	}

	ticker.High, _ = strconv.ParseFloat(args[8].(string), 64)
	ticker.Low, _ = strconv.ParseFloat(args[9].(string), 64)
}

type HitBTCWebsocketTrollboxMessage struct {
	MessageNumber float64
	Username      string
	Message       string
	Reputation    float64
}

func HitBTCOnTrollbox(args []interface{}, kwargs map[string]interface{}) {
	message := HitBTCWebsocketTrollboxMessage{}
	message.MessageNumber, _ = args[1].(float64)
	message.Username = args[2].(string)
	message.Message = args[3].(string)
	if len(args) == 5 {
		message.Reputation = args[4].(float64)
	}
}

func HitBTCOnDepthOrTrade(args []interface{}, kwargs map[string]interface{}) {
	for x := range args {
		data := args[x].(map[string]interface{})
		msgData := data["data"].(map[string]interface{})
		msgType := data["type"].(string)

		switch msgType {
		case "orderBookModify":
			{
				type HitBTCWebsocketOrderbookModify struct {
					Type   string
					Rate   float64
					Amount float64
				}

				orderModify := HitBTCWebsocketOrderbookModify{}
				orderModify.Type = msgData["type"].(string)

				rateStr := msgData["rate"].(string)
				orderModify.Rate, _ = strconv.ParseFloat(rateStr, 64)

				amountStr := msgData["amount"].(string)
				orderModify.Amount, _ = strconv.ParseFloat(amountStr, 64)
			}
		case "orderBookRemove":
			{
				type HitBTCWebsocketOrderbookRemove struct {
					Type string
					Rate float64
				}

				orderRemoval := HitBTCWebsocketOrderbookRemove{}
				orderRemoval.Type = msgData["type"].(string)

				rateStr := msgData["rate"].(string)
				orderRemoval.Rate, _ = strconv.ParseFloat(rateStr, 64)
			}
		case "newTrade":
			{
				type HitBTCWebsocketNewTrade struct {
					Type    string
					TradeID int64
					Rate    float64
					Amount  float64
					Date    string
					Total   float64
				}

				trade := HitBTCWebsocketNewTrade{}
				trade.Type = msgData["type"].(string)

				tradeIDstr := msgData["tradeID"].(string)
				trade.TradeID, _ = strconv.ParseInt(tradeIDstr, 10, 64)

				rateStr := msgData["rate"].(string)
				trade.Rate, _ = strconv.ParseFloat(rateStr, 64)

				amountStr := msgData["amount"].(string)
				trade.Amount, _ = strconv.ParseFloat(amountStr, 64)

				totalStr := msgData["total"].(string)
				trade.Rate, _ = strconv.ParseFloat(totalStr, 64)

				trade.Date = msgData["date"].(string)
			}
		}
	}
}

func (p *HitBTC) WebsocketClient() {
	for p.Enabled && p.Websocket {
		c, err := turnpike.NewWebsocketClient(turnpike.JSON, HITBTC_WEBSOCKET_ADDRESS, nil)
		if err != nil {
			log.Printf("%s Unable to connect to Websocket. Error: %s\n", p.GetName(), err)
			continue
		}

		if p.Verbose {
			log.Printf("%s Connected to Websocket.\n", p.GetName())
		}

		_, err = c.JoinRealm(HITBTC_WEBSOCKET_REALM, nil)
		if err != nil {
			log.Printf("%s Unable to join realm. Error: %s\n", p.GetName(), err)
			continue
		}

		if p.Verbose {
			log.Printf("%s Joined Websocket realm.\n", p.GetName())
		}

		c.ReceiveDone = make(chan bool)

		if err := c.Subscribe(HITBTC_WEBSOCKET_TICKER, HitBTCOnTicker); err != nil {
			log.Printf("%s Error subscribing to ticker channel: %s\n", p.GetName(), err)
		}

		if err := c.Subscribe(HITBTC_WEBSOCKET_TROLLBOX, HitBTCOnTrollbox); err != nil {
			log.Printf("%s Error subscribing to trollbox channel: %s\n", p.GetName(), err)
		}

		for x := range p.EnabledPairs {
			currency := p.EnabledPairs[x]
			if err := c.Subscribe(currency, HitBTCOnDepthOrTrade); err != nil {
				log.Printf("%s Error subscribing to %s channel: %s\n", p.GetName(), currency, err)
			}
		}

		if p.Verbose {
			log.Printf("%s Subscribed to websocket channels.\n", p.GetName())
		}

		<-c.ReceiveDone
		log.Printf("%s Websocket client disconnected.\n", p.GetName())
	}
}
