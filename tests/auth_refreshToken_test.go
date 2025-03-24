package tests

import (
	"testing"
	"time"

	"github.com/hesoyamTM/apphelper-sso/tests/suite"

	"github.com/brianvoe/gofakeit"
	ssov1 "github.com/hesoyamTM/apphelper-protos/gen/go/sso"
	"github.com/stretchr/testify/require"
)

func TestRefreshToken_HappyPath(t *testing.T) {
	ctx, st := suite.New(t)

	name := gofakeit.FirstName()
	surname := gofakeit.LastName()

	respReg, err := st.AuthClient.Register(ctx, &ssov1.RegisterRequest{
		Name:     name,
		Surname:  surname,
		Login:    gofakeit.BeerName(),
		Password: randPass(),
	})

	require.NoError(t, err)
	require.NotEmpty(t, respReg.GetAccessToken())
	require.NotEmpty(t, respReg.GetRefreshToken())

	refreshToken := respReg.GetRefreshToken()

	respRefToken, err := st.AuthClient.RefreshToken(ctx, &ssov1.RefreshTokenRequest{
		RefreshToken: refreshToken,
	})

	refreshTime := time.Now()

	require.NoError(t, err)
	require.NotEmpty(t, respRefToken.GetAccessToken())
	require.NotEmpty(t, respRefToken.GetRefreshToken())

	checkJWT(t, respRefToken.GetAccessToken(), name, surname, refreshTime.Add(st.Cfg.AccessTokenTTL))
}

func TestRefreshToken_TokenExpired(t *testing.T) {
	ctx, st := suite.New(t)

	respReg, err := st.AuthClient.Register(ctx, &ssov1.RegisterRequest{
		Name:     gofakeit.FirstName(),
		Surname:  gofakeit.LastName(),
		Login:    gofakeit.LastName(),
		Password: randPass(),
	})

	require.NoError(t, err)
	require.NotEmpty(t, respReg.GetAccessToken())
	require.NotEmpty(t, respReg.GetRefreshToken())

	refreshToken := respReg.GetRefreshToken()

	time.Sleep(time.Second * 3)

	respRefToken, err := st.AuthClient.RefreshToken(ctx, &ssov1.RefreshTokenRequest{
		RefreshToken: refreshToken,
	})

	require.Error(t, err)
	require.ErrorContains(t, err, "not authorized")
	require.Empty(t, respRefToken.GetAccessToken())
	require.Empty(t, respRefToken.GetRefreshToken())
}
