package transport

import (
	"fmt"
	"time"

	"github.com/layerhq/go-client/common"

	jwt "github.com/dgrijalva/jwt-go"
)

func localCredentialTokenFactory(credentials *common.ClientCredentials, nonce string) (token string, err error) {
	if credentials.Key == nil {
		return "", fmt.Errorf("No key data specified")
	}

	// Set claims
	claims := jwt.MapClaims{}
	claims["iss"] = credentials.ProviderID
	claims["prn"] = credentials.User
	claims["iat"] = time.Now().Unix()
	claims["exp"] = time.Now().Add(time.Hour * 72).Unix()
	claims["nce"] = nonce

	// Create a token
	jwtToken := jwt.NewWithClaims(jwt.GetSigningMethod("RS256"), claims)

	// Set header values
	jwtToken.Header["typ"] = "JWT"
	jwtToken.Header["alg"] = "RS256"
	jwtToken.Header["cty"] = "layer-eit;v=1"
	jwtToken.Header["kid"] = credentials.Key.ID

	// Build the keypair from the key
	key, err := jwt.ParseRSAPrivateKeyFromPEM([]byte(credentials.Key.KeyPair.Private))
	if err != nil {
		return "", fmt.Errorf("Error getting private key from keypair - %v", err)
	}

	// Sign and get the complete encoded token as a string
	return jwtToken.SignedString(key)
}
