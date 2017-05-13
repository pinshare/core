package lib

import (
	"fmt"

	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"io/ioutil"

	jwt "github.com/dgrijalva/jwt-go"
)

var CURVE = elliptic.P256()

type key struct {
	X []byte `json:"X"`
	Y []byte `json:"Y"`
	D []byte `json:"D"`
}

func GenerateToken(userId) string {
	privKey, _ := ecdsa.GenerateKey(CURVE, rand.Reader)
	key := key{
		X: privKey.PublicKey.X.Bytes(),
		Y: privKey.PublicKey.Y.Bytes(),
		D: privKey.D.Bytes(),
	}

	buf, _ := json.Marshal(key)
	encoded := base64.StdEncoding.EncodeToString(buf)
	ioutil.WriteFile("./serverKey", []byte(encoded), 0644)

	claims := &jwt.StandardClaims{
		Audience: "SampleAudience",
		Subject:  "SampleSubject",
		Issuer:   "SampleIssuer",
		Id:       data,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodES256, claims)
	ss, _ := token.SignedString(privKey)
	fmt.Printf("Singed String: %s¥n", ss)

	verifyToken(ss, privKey.PublicKey)
}

func verifyToken(ss string, pubKey ecdsa.PublicKey) {
	token, err := jwt.Parse(ss, func(token *jwt.Token) (interface{}, error) {
		_, ok := token.Method.(*jwt.SigningMethodECDSA)
		if ok {
			return &pubKey, nil
		} else {
			return nil, fmt.Errorf("Token parse error")
		}
	})

	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Printf("Verify success: %v¥n", token)
}
