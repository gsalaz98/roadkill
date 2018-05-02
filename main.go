package main

import (
	"gitlab.com/CuteQ/roadkill/exchanges/slow/bitmex"
	"gitlab.com/CuteQ/roadkill/orderbook"
)

func main() {
	var receiver = make(chan orderbook.Delta, 1<<16)

	//polo := poloniex.DefaultSettings
	//polo.Initialize("BTC_BCN", "BTC_BELA", "BTC_BLK", "BTC_BTCD", "BTC_BTM", "BTC_BTS", "BTC_BURST", "BTC_CLAM", "BTC_DGB", "BTC_DOGE", "BTC_DASH", "BTC_EMC2", "BTC_FLDC", "BTC_GAME", "BTC_HUC", "BTC_LTC", "BTC_MAID", "BTC_OMNI", "BTC_NAV", "BTC_NEOS", "BTC_NMC", "BTC_NXT", "BTC_PINK", "BTC_POT", "BTC_PPC", "BTC_RIC", "BTC_STR", "BTC_SYS", "BTC_VIA", "BTC_VRC", "BTC_VTC", "BTC_XBC", "BTC_XCP", "BTC_XMR", "BTC_XPM", "BTC_XRP", "BTC_XVC", "BTC_FLO", "BTC_XEM", "BTC_GRC", "BTC_ETH", "BTC_SC", "BTC_BCY", "BTC_EXP", "BTC_FCT", "BTC_RADS", "BTC_AMP", "BTC_DCR", "BTC_LSK", "BTC_LBC", "BTC_STEEM", "BTC_SBD", "BTC_ETC", "BTC_REP", "BTC_ARDR", "BTC_ZEC", "BTC_STRAT", "BTC_NXC", "BTC_PASC", "BTC_GNT", "BTC_GNO", "BTC_BCH", "BTC_ZRX", "BTC_CVC", "BTC_OMG", "BTC_GAS", "BTC_STORJ", "USDT_BTC", "USDT_DASH", "USDT_LTC", "USDT_NXT", "USDT_STR", "USDT_XMR", "USDT_XRP", "USDT_ETH", "USDT_ETC", "USDT_REP", "USDT_ZEC", "USDT_BCH", "XMR_BCN", "XMR_BLK", "XMR_BTCD", "XMR_DASH", "XMR_LTC", "XMR_MAID", "XMR_NXT", "XMR_ZEC", "ETH_LSK", "ETH_STEEM", "ETH_ETC", "ETH_REP", "ETH_ZEC", "ETH_GNT", "ETH_GNO", "ETH_BCH", "ETH_ZRX", "ETH_CVC", "ETH_OMG", "ETH_GAS")
	//polo.Initialize("BTC_ETH", "USDT_BTC")
	//polo.ReceiveMessageLoop(receiver)

	//go polo.ReceiveMessageLoop(receiver)
	//
	//go func(recv chan orderbook.Delta) {
	//	for {
	//		fmt.Println(<-recv)
	//	}
	//}(receiver)
	//
	//for {
	//}

	//bitm := bitmex.DefaultSettings
	//bitm.Initialize("XBTUSD", "ETHM18")
	//bitm.ReceiveMessageLoop(receiver)

	bitm := bitmexslow.DefaultSettings
	bitm.ChannelType = []string{"orderBookL2", "trade"}

	bitm.Initialize("XBTUSD", "ETHM18", "XBT7D_U110")
	bitm.ReceiveMessageLoop(&receiver)

}
