package key

import (
	"context"
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/pem"
	"log/slog"
	"time"

	"github.com/hesoyamTM/apphelper-sso/internal/clients/report"
	"github.com/hesoyamTM/apphelper-sso/internal/clients/schedule"
	"github.com/hesoyamTM/apphelper-sso/internal/lib/jwt"
)

type App struct {
	log            *slog.Logger
	updateInterval time.Duration

	privateKey *ecdsa.PrivateKey
	publicKey  *ecdsa.PublicKey

	privateKeyChan chan *ecdsa.PrivateKey

	reportClient   *report.Client
	scheduleClient *schedule.Client
}

func New(log *slog.Logger, updInterval time.Duration, repCLient *report.Client, schedClient *schedule.Client) *App {
	privKeyChan := make(chan *ecdsa.PrivateKey, 1)

	return &App{
		log: log,

		updateInterval: updInterval,
		privateKeyChan: privKeyChan,
		reportClient:   repCLient,
		scheduleClient: schedClient,
	}
}

func (k *App) Run() {
	for {
		k.log.Info("generating a new key pair")
		privkey, pubKey, err := jwt.GenerateKeys()
		if err != nil {
			k.log.Error("failed to generate keys")
		}
		k.log.Info("generated")

		k.privateKey = privkey
		k.publicKey = pubKey

		encodedKey := encode(pubKey)
		err = k.reportClient.SetPublicKey(context.Background(), encodedKey)
		if err != nil {
			k.log.Error(err.Error())
		}
		err = k.scheduleClient.SetPublicKey(context.Background(), encodedKey)
		if err != nil {
			k.log.Error(err.Error())
		}
		k.log.Info("public key has been sent to other services")

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

func encode(publicKey *ecdsa.PublicKey) string {
	x509EncodedPub, _ := x509.MarshalPKIXPublicKey(publicKey)
	pemEncodedPub := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: x509EncodedPub})

	return string(pemEncodedPub)
}

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
