package main

import "github.com/Bullpeen/stox"

func main() {
	stox.MakeItSo("API_KEY")

	sym := "P500"

	if stox.ValidSym(sym) {

		res, err := stox.GetQuote(sym)

		if err == nil {
			print(res)
		}
	}
}