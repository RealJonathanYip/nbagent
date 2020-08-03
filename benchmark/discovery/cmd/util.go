package cmd

import (
	"fmt"
	"strings"
	"time"
)

func GenerateInstanceID(szName string, nIndex int) string {

	return fmt.Sprintf("%s_%d_%d", szName, nIndex, time.Now().UnixNano())
}

func GenerateRegisterURI(szName string, nIndex, nSize int) string {

	szUri := fmt.Sprintf("%s_%d", szName, nIndex)
	if len(szUri) >= nSize {
		panic("GenerateRegisterURI exceed uri size")
	}

	nLen := nSize - len(szUri)
	szUri += strings.Repeat("u", nLen)
	return szUri
}

func GenerateRegisterURIWithSpace(szName string, nIndex, nSize, nSpace int) string {

	szUri := fmt.Sprintf("%s_%d", szName, nIndex%nSpace)
	if len(szUri) >= nSize {
		panic("GenerateRegisterURI exceed uri size")
	}

	nLen := nSize - len(szUri)
	szUri += strings.Repeat("u", nLen)
	return szUri
}
