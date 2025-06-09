package tests

import (
	"github.com/brianvoe/gofakeit/v7"
	"github.com/golang-jwt/jwt/v5"
	ssov1 "github.com/makar182/protos/gen/sso"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"sso/internal/services/auth"
	"sso/tests/suite"
	"testing"
	"time"
)

const (
	emptyAppId = 0
	appId      = 1
	appSecret  = "default_secret"

	passDefaultLen = 10
)

func TestRegisterLogin_Login_HappyPath(t *testing.T) {
	ctx, st := suite.NewSuite(t)

	email := gofakeit.Email()
	password := randomPassword()

	regResp, err := st.AuthClient.Register(ctx, &ssov1.RegisterRequest{
		Email:    email,
		Password: password,
	})
	require.NoError(t, err)
	require.NotEmpty(t, regResp.GetUserId())

	loginResp, err := st.AuthClient.Login(ctx, &ssov1.LoginRequest{
		Email:    email,
		Password: password,
		AppId:    appId,
	})

	require.NoError(t, err)
	require.NotEmpty(t, loginResp.Token)

	loginTime := time.Now()
	token, err := jwt.Parse(loginResp.GetToken(), func(token *jwt.Token) (interface{}, error) {
		return []byte(appSecret), nil
	})
	require.NoError(t, err)

	claims, ok := token.Claims.(jwt.MapClaims)
	assert.True(t, ok)

	assert.Equal(t, regResp.GetUserId(), int64(claims["user_id"].(float64)))
	assert.Equal(t, email, claims["email"].(string))
	assert.Equal(t, appId, int(claims["app_id"].(float64)))

	deltaSeconds := 1
	assert.InDelta(t, loginTime.Add(st.Cfg.TokenTTL).Unix(), int64(claims["exp"].(float64)), float64(deltaSeconds), "Token expiration time should be within 1 second of expected value")

}

func randomPassword() string {
	return gofakeit.Password(true, true, true, true, false, passDefaultLen)
}

func TestRegister_FailCases(t *testing.T) {
	ctx, st := suite.NewSuite(t)

	tests := []struct {
		name        string
		email       string
		password    string
		expectedErr string
	}{
		{
			name:        "Empty email",
			email:       "",
			password:    randomPassword(),
			expectedErr: "email and password must be provided",
		},
		{
			name:        "Empty password",
			email:       gofakeit.Email(),
			password:    "",
			expectedErr: "email and password must be provided",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := st.AuthClient.Register(ctx, &ssov1.RegisterRequest{
				Email:    tt.email,
				Password: tt.password})

			assert.Contains(t, err.Error(), tt.expectedErr, "Expected error message to contain: %s", tt.expectedErr)

		})
	}
}

func TestLogin_FailCases(t *testing.T) {
	ctx, st := suite.NewSuite(t)

	tests := []struct {
		name        string
		email       string
		password    string
		appId       int32
		expectedErr string
	}{
		{
			name:        "Empty email",
			email:       "",
			password:    randomPassword(),
			appId:       appId,
			expectedErr: "email, password and app_id must be provided",
		},
		{
			name:        "Empty password",
			email:       gofakeit.Email(),
			password:    "",
			appId:       appId,
			expectedErr: "email, password and app_id must be provided",
		},
		{
			name:        "Empty appId",
			email:       gofakeit.Email(),
			password:    randomPassword(),
			appId:       emptyAppId,
			expectedErr: "email, password and app_id must be provided",
		},
		{
			name:        "Wrong password",
			email:       gofakeit.Email(),
			password:    randomPassword(),
			appId:       appId,
			expectedErr: "failed to login: " + auth.ErrInternalServerError.Error(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if tt.name == "Wrong password" {
				_, err := st.AuthClient.Register(ctx, &ssov1.RegisterRequest{
					Email:    tt.email,
					Password: tt.password})

				require.NoError(t, err, "Registration should succeed for test setup")
				tt.password = "wrong-password"
			}

			_, err := st.AuthClient.Login(ctx, &ssov1.LoginRequest{
				Email:    tt.email,
				Password: tt.password,
				AppId:    tt.appId})

			assert.Contains(t, err.Error(), tt.expectedErr, "Expected error message to contain: %s", tt.expectedErr)

		})
	}
}

func TestRegister_Duplication(t *testing.T) {
	ctx, st := suite.NewSuite(t)

	email := gofakeit.Email()
	password := randomPassword()

	// First registration should succeed
	regResp, err := st.AuthClient.Register(ctx, &ssov1.RegisterRequest{
		Email:    email,
		Password: password,
	})
	require.NoError(t, err)
	require.NotEmpty(t, regResp.GetUserId())

	// Second registration with the same email should fail
	_, err = st.AuthClient.Register(ctx, &ssov1.RegisterRequest{
		Email:    email,
		Password: password,
	})

	assert.ErrorContains(t, err, "failed to register: "+auth.ErrInternalServerError.Error())
}
