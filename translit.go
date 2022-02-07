package main

import (
	"math/rand"
	"strings"
	"time"

	"pkg.re/essentialkaos/translit.v2"
)

func transliterate(str string) string {
	symbols := make([]string, len(str))
	symbols = strings.Split(str, "")

	msg := make([]string, len(str))

	for i := 0; i < len(symbols); i++ {
		rand.Seed(time.Now().UnixNano())
		ran := rand.Intn(8)
		switch ran {
		case 0:
			msg[i] = translit.EncodeToALALC(symbols[i])
		case 1:
			msg[i] = translit.EncodeToBGN(symbols[i])
		case 2:
			msg[i] = translit.EncodeToBS(symbols[i])
		case 3:
			msg[i] = translit.EncodeToICAO(symbols[i])
		case 4:
			msg[i] = translit.EncodeToISO9A(symbols[i])
		case 5:
			msg[i] = translit.EncodeToISO9B(symbols[i])
		case 6:
			msg[i] = translit.EncodeToPCGN(symbols[i])
		case 7:
			msg[i] = translit.EncodeToScientific(symbols[i])
		}
	}
	return strings.Join(msg, "")
}
