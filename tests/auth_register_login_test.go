package tests

import (
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/pem"
	"testing"
	"time"

	"github.com/hesoyamTM/apphelper-sso/tests/suite"

	"github.com/brianvoe/gofakeit"
	"github.com/golang-jwt/jwt/v5"
	ssov1 "github.com/hesoyamTM/apphelper-protos/gen/go/sso"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	passLen   = 20
	publicKey = `-----BEGIN PUBLIC KEY-----
MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAELQSAFUnyfqA7OCojTTtUXPltZ1Bc
+dSeYmQq+Zr5qBGcUzbf7dnkZLSVeueXFyOlkPUfKVEpSKlwP6XROuEBlA==
-----END PUBLIC KEY-----`
)

func TestRegisterLoginHappyPath(t *testing.T) {
	ctx, st := suite.New(t)

	name := gofakeit.FirstName()
	surname := gofakeit.LastName()
	login := gofakeit.Word()
	pass := randPass()

	respReg, err := st.AuthClient.Register(ctx, &ssov1.RegisterRequest{
		Name:     name,
		Surname:  surname,
		Login:    login,
		Password: pass,
	})
	loginTime := time.Now()
	require.NoError(t, err)

	checkJWT(t, respReg.GetAccessToken(), name, surname, loginTime.Add(st.Cfg.AccessTokenTTL))

	require.NoError(t, err)
	assert.NotEmpty(t, respReg.GetAccessToken())
	assert.NotEmpty(t, respReg.GetRefreshToken())

	respLog, err := st.AuthClient.Login(ctx, &ssov1.LoginRequest{
		Login:    login,
		Password: pass,
	})

	loginTime = time.Now()

	require.NoError(t, err)

	checkJWT(t, respLog.GetAccessToken(), name, surname, loginTime.Add(st.Cfg.AccessTokenTTL))
}

func TestRegisterLogin_DuplicatedRegistration(t *testing.T) {
	ctx, st := suite.New(t)

	name := gofakeit.FirstName()
	surname := gofakeit.LastName()
	login := gofakeit.FirstName()
	pass := randPass()

	respReg, err := st.AuthClient.Register(ctx, &ssov1.RegisterRequest{
		Name:     name,
		Surname:  surname,
		Login:    login,
		Password: pass,
	})
	require.NoError(t, err)

	respReg, err = st.AuthClient.Register(ctx, &ssov1.RegisterRequest{
		Name:     name,
		Surname:  surname,
		Login:    login,
		Password: pass,
	})
	require.Error(t, err)
	assert.Empty(t, respReg.GetAccessToken())
	assert.Empty(t, respReg.GetRefreshToken())
	assert.ErrorContains(t, err, "user already exists")
}

func TestRegister_FailCases(t *testing.T) {
	ctx, st := suite.New(t)

	tests := []struct {
		testName string
		name     string
		surname  string
		login    string
		pass     string
		expErr   string
	}{
		{
			testName: "Register with empty name",
			name:     "",
			surname:  gofakeit.LastName(),
			login:    gofakeit.LastName(),
			pass:     randPass(),
			expErr:   "validation error",
		},
		{
			testName: "Register with empty surname",
			name:     gofakeit.FirstName(),
			surname:  "",
			login:    gofakeit.LastName(),
			pass:     randPass(),
			expErr:   "validation error",
		},
		{
			testName: "Register with empty login",
			name:     gofakeit.FirstName(),
			surname:  gofakeit.LastName(),
			login:    "",
			pass:     randPass(),
			expErr:   "validation error",
		},
		{
			testName: "Register with empty password",
			name:     gofakeit.FirstName(),
			surname:  gofakeit.LastName(),
			login:    gofakeit.LastName(),
			pass:     "",
			expErr:   "validation error",
		},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			_, err := st.AuthClient.Register(ctx, &ssov1.RegisterRequest{
				Name:     test.name,
				Surname:  test.surname,
				Login:    test.login,
				Password: test.pass,
			})
			require.Error(t, err)
			assert.ErrorContains(t, err, test.expErr)
		})
	}
}

func TestLogin_FailCases(t *testing.T) {
	ctx, st := suite.New(t)

	name := gofakeit.FirstName()
	surname := gofakeit.LastName()
	login := gofakeit.LastName()
	pass := randPass()

	_, err := st.AuthClient.Register(ctx, &ssov1.RegisterRequest{
		Name:     name,
		Surname:  surname,
		Login:    login,
		Password: pass,
	})

	require.NoError(t, err)

	tests := []struct {
		testName string
		login    string
		pass     string
		expErr   string
	}{
		{
			testName: "Login with empty login",
			login:    "",
			pass:     pass,
			expErr:   "validation error",
		},
		{
			testName: "Login with empty password",
			login:    login,
			pass:     "",
			expErr:   "validation error",
		},
		{
			testName: "Login with wrong password",
			login:    login,
			pass:     randPass(),
			expErr:   "invalid credentials",
		},
		{
			testName: "Login with empty both",
			login:    "",
			pass:     "",
			expErr:   "validation error",
		},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			_, err := st.AuthClient.Login(ctx, &ssov1.LoginRequest{
				Login:    test.login,
				Password: test.pass,
			})
			require.Error(t, err)
			assert.ErrorContains(t, err, test.expErr)
		})
	}
}

func randPass() string {
	return gofakeit.Password(true, true, true, true, true, passLen)
}

func checkJWT(t *testing.T, token, name, surname string, exp time.Time) {
	require.NotEmpty(t, token)

	tokenParsed, err := jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
		return decode(publicKey)
	})

	require.NoError(t, err)

	claims, ok := tokenParsed.Claims.(jwt.MapClaims)
	require.True(t, ok)

	assert.Equal(t, name, claims["name"].(string))
	assert.Equal(t, surname, claims["surname"].(string))

	const deltaSeconds = 1

	assert.InDelta(t, exp.Unix(), claims["exp"].(float64), deltaSeconds)
}

func decode(pemEncodedPub string) (*ecdsa.PublicKey, error) {
	blockPub, _ := pem.Decode([]byte(pemEncodedPub))
	x509EncodedPub := blockPub.Bytes
	genericPublicKey, err := x509.ParsePKIXPublicKey(x509EncodedPub)
	if err != nil {
		return nil, err
	}

	publicKey := genericPublicKey.(*ecdsa.PublicKey)

	return publicKey, nil
}
