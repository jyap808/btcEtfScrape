package funds

import (
	"github.com/jyap808/btcEtfScrape/funds/arkb"
	"github.com/jyap808/btcEtfScrape/funds/bitb"
	"github.com/jyap808/btcEtfScrape/funds/brrr"
	"github.com/jyap808/btcEtfScrape/funds/btcw"
	"github.com/jyap808/btcEtfScrape/funds/ezbc"
	"github.com/jyap808/btcEtfScrape/funds/fbtc"
	"github.com/jyap808/btcEtfScrape/funds/gbtc"
	"github.com/jyap808/btcEtfScrape/funds/hodl"
	"github.com/jyap808/btcEtfScrape/funds/ibit"
	"github.com/jyap808/btcEtfScrape/types"
)

func ArkbCollect() types.Result {
	return arkb.Collect()
}

func BitbCollect() types.Result {
	return bitb.Collect()
}

func BrrrCollect() types.Result {
	return brrr.Collect()
}

func BtcwCollect() types.Result {
	return btcw.Collect()
}

func EzbcCollect() types.Result {
	return ezbc.Collect()
}

func FbtcCollect() types.Result {
	return fbtc.Collect()
}

func GbtcCollect() types.Result {
	return gbtc.Collect()
}

func HodlCollect() types.Result {
	return hodl.Collect()
}

func IbitCollect() types.Result {
	return ibit.Collect()
}
