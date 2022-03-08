package license

import (
	"encoding/base64"
	"errors"
	"github.com/destructiqn/kogtevran/generic"
	"log"
	"os"
	"time"

	"github.com/golang-jwt/jwt"
)

var SigningKey = getSigningKey()

func getSigningKey() []byte {
	raw, ok := os.LookupEnv("KV_SIGNING_KEY")
	if !ok {
		log.Println("signing key is not available")
		return nil
	}

	bytes, err := base64.StdEncoding.DecodeString(raw)
	if err != nil {
		log.Println("unable to decode signing key: " + err.Error())
		return nil
	}

	return bytes
}

type KogtevranClaims struct {
	jwt.StandardClaims
	Features uint64 `json:"fts"`
}

func (k *KogtevranClaims) IsRelated(tunnel generic.Tunnel) bool {
	return tunnel.GetRemoteAddr() == k.Subject
}

func (k *KogtevranClaims) HasFeature(feature Feature) bool {
	return k.Features&uint64(feature) > 0
}

func (k *KogtevranClaims) GetFeatures() uint64 {
	return k.Features
}

func GetLicense(key string) (License, error) {
	if generic.IsDevelopmentEnvironment() {
		return &DevelopmentLicense{}, nil
	}

	var claims KogtevranClaims
	token, err := jwt.ParseWithClaims(key, &claims, func(token *jwt.Token) (interface{}, error) {
		if SigningKey == nil {
			return nil, errors.New("signing key is not available")
		}
		return SigningKey, nil
	})
	if err != nil {
		return nil, err
	}

	if !token.Valid || claims.ExpiresAt < time.Now().Unix() {
		return nil, errors.New("invalid token")
	}

	return &claims, nil
}
