package key

import (
	"crypto/ecdsa"
	"log/slog"
	"sso/pkg/lib/jwt"
	"time"
)

type App struct {
	log            *slog.Logger
	updateInterval time.Duration

	privateKey *ecdsa.PrivateKey
	publicKey  *ecdsa.PublicKey

	privateKeyChan chan *ecdsa.PrivateKey
}

func New(log *slog.Logger, updInterval time.Duration) *App {
	privKeyChan := make(chan *ecdsa.PrivateKey, 1)

	return &App{
		updateInterval: updInterval,
		privateKeyChan: privKeyChan,
	}
}

func (k *App) Run() {
	for {
		//k.log.Info("generating a new key pair")
		privkey, pubKey, err := jwt.GenerateKeys()
		if err != nil {
			k.log.Error("failed to generate keys")
		}
		//k.log.Info("generated")

		k.privateKey = privkey
		k.publicKey = pubKey

		//fmt.Println(encode(privkey, pubKey))

		//TODO: send public key to others microservices

		k.privateKeyChan <- privkey
		<-time.After(k.updateInterval)
	}
}

func (k *App) Stop() {
	close(k.privateKeyChan)
}

func (k *App) GetPrivateKeyChan() <-chan *ecdsa.PrivateKey {
	return k.privateKeyChan
}

// func encode(privateKey *ecdsa.PrivateKey, publicKey *ecdsa.PublicKey) (string, string) {
// 	x509Encoded, _ := x509.MarshalECPrivateKey(privateKey)
// 	pemEncoded := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: x509Encoded})

// 	x509EncodedPub, _ := x509.MarshalPKIXPublicKey(publicKey)
// 	pemEncodedPub := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: x509EncodedPub})

// 	return string(pemEncoded), string(pemEncodedPub)
// }

// func decode(pemEncoded string, pemEncodedPub string) (*ecdsa.PrivateKey, *ecdsa.PublicKey) {
// 	block, _ := pem.Decode([]byte(pemEncoded))
// 	x509Encoded := block.Bytes
// 	privateKey, _ := x509.ParseECPrivateKey(x509Encoded)

// 	blockPub, _ := pem.Decode([]byte(pemEncodedPub))
// 	x509EncodedPub := blockPub.Bytes
// 	genericPublicKey, _ := x509.ParsePKIXPublicKey(x509EncodedPub)
// 	publicKey := genericPublicKey.(*ecdsa.PublicKey)

// 	return privateKey, publicKey
// }
