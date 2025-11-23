package tests

import (
	"sso/services/auth"
	"sso/tests/suite"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v6"
	ssov1 "github.com/glebateee/protos/gen/go/sso"
	"github.com/golang-jwt/jwt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	emptyAppID     = 0
	appID          = 1
	appSecret      = "test-secret"
	passDefaultLen = 10
)

func TestRegisterLogin_Login_HappyPath(t *testing.T) {
	ctx, st := suite.New(t)
	email := gofakeit.Email()
	pass := randomFakePassword()
	regResp, err := st.AuthClient.Register(ctx, &ssov1.RegisterRequest{
		Email:    email,
		Password: pass,
	})
	require.NoError(t, err)
	assert.NotEmpty(t, regResp.GetUserId())
	loginResp, err := st.AuthClient.Login(ctx, &ssov1.LoginRequest{
		Email:    email,
		Password: pass,
		AppId:    appID,
	})
	require.NoError(t, err)
	loginTime := time.Now()

	token := loginResp.GetToken()
	assert.NotEmpty(t, token)
	tokenParsed, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		return []byte(appSecret), nil
	})
	require.NoError(t, err)
	claims, ok := tokenParsed.Claims.(jwt.MapClaims)
	assert.True(t, ok)
	assert.Equal(t, regResp.GetUserId(), int64(claims["uid"].(float64)))
	assert.Equal(t, email, claims["email"].(string))
	assert.Equal(t, appID, int(claims["app_id"].(float64)))
	const deltaSecond = 1
	assert.InDelta(t, loginTime.Add(st.Cfg.TokenTTL).Unix(), int64(claims["exp"].(float64)), deltaSecond)
}

func TestRegisterDouble(t *testing.T) {
	ctx, st := suite.New(t)
	email := gofakeit.Email()
	pass := randomFakePassword()
	regResp, err := st.AuthClient.Register(ctx, &ssov1.RegisterRequest{
		Email:    email,
		Password: pass,
	})
	require.NoError(t, err)
	assert.NotEmpty(t, regResp.GetUserId())

	regResp, err = st.AuthClient.Register(ctx, &ssov1.RegisterRequest{
		Email:    email,
		Password: pass,
	})
	require.Error(t, err)
	assert.Empty(t, regResp.GetUserId())
	assert.ErrorContains(t, err, auth.ErrUserExists.Error())
}

func randomFakePassword() string {
	return gofakeit.Password(true, true, true, true, false, passDefaultLen)
}
