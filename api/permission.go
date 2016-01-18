package api

import (
	"crypto/sha256"
	"encoding/hex"

	"github.com/labstack/echo"
)

var (
	hashedSecret string
)

func init() {
	hashedSecret = authSecret("kis.json")["secret"].(string)
	data := sha256.Sum256([]byte(hashedSecret))
	hashedSecret = hex.EncodeToString(data[:])
}

// admin check
func isAdmin(ctx *echo.Context) (res bool) {
	res = hashCompare(ctx.Query("secret"), hashedSecret)
	return
}
