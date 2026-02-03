package utils

import "strings"

func GetComparableMnemonic(mnemonic string) string {
	var tbr = strings.ToLower(mnemonic)
	tbr = strings.ReplaceAll(tbr, "_", "")
	tbr = strings.ReplaceAll(tbr, "-", "")
	return tbr
}
