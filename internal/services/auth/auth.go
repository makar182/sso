package auth

import (
	"context"
	"errors"
	"golang.org/x/crypto/bcrypt"
	"log/slog"
	"sso/internal/domain/models"
	"sso/internal/lib/jwt"
	"sso/internal/lib/logger/sl"
	"sso/internal/storage"
	"time"
)

type Auth struct {
	log          *slog.Logger
	userSaver    UserSaver
	userProvider UserProvider
	appProvider  AppProvider
	tokenTTL     time.Duration
}

type UserSaver interface {
	SaveUser(email string, passHash []byte) (int64, error)
}

type UserProvider interface {
	GetUserByEmail(email string) (*models.User, error)
	IsAdmin(userId int64) (bool, error)
}

type AppProvider interface {
	GetAppById(appId int) (*models.App, error)
}

var (
	ErrInvalidCredentials  = errors.New("invalid credentials")
	ErrInternalServerError = errors.New("internal server error")
)

// NewAuthService creates a new instance of Auth with the provided dependencies.
func NewAuthService(
	log *slog.Logger,
	userSaver UserSaver,
	userProvider UserProvider,
	appProvider AppProvider,
	tokenTTL time.Duration) *Auth {
	return &Auth{
		log:          log,
		userSaver:    userSaver,
		userProvider: userProvider,
		appProvider:  appProvider,
		tokenTTL:     tokenTTL,
	}
}

func (a *Auth) Login(ctx context.Context, email string, password string, appId int) (string, error) {
	const op = "Auth.Login"
	log := a.log.With(slog.String("op", op), slog.String("email", email), slog.Int("appId", appId))
	user, err := a.userProvider.GetUserByEmail(email)
	//зарефакторить этот блок по итогу реализации стореджа, потому что не ясно как будет выглядеть ненайденный юзер
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) || user == nil {
			log.Info("user not found", sl.Err(err))
			return "", ErrInvalidCredentials
		}
		log.Error("failed to get user by email", sl.Err(err))
		return "", ErrInternalServerError
	}

	err = bcrypt.CompareHashAndPassword(user.PassHash, []byte(password))
	if err != nil {
		log.Info("password mismatch", sl.Err(err))
		return "", ErrInternalServerError
	}

	app, err := a.appProvider.GetAppById(appId)
	if err != nil {
		if errors.Is(err, storage.ErrAppNotFound) {
			log.Info("app not found", sl.Err(err))
			return "", ErrInvalidCredentials
		}

		log.Error("failed to get app by id", sl.Err(err))
		return "", ErrInternalServerError
	}

	token, err := jwt.NewToken(user, app, a.tokenTTL)

	log.Info("user logged in successfully", slog.Int64("userId", user.Id), slog.String("appName", app.Name))
	return token, nil
}

func (a *Auth) Logout(ctx context.Context, token string) (bool, error) {
	// Implement logout logic here
	return false, nil
}

func (a *Auth) Register(ctx context.Context, email string, password string) (int64, error) {
	const op = "Auth.RegisterNewUser"
	log := a.log.With(slog.String("op", op), slog.String("email", email))

	passHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Error("failed to hash password", sl.Err(err))
		return 0, ErrInternalServerError
	}

	_, err = a.userProvider.GetUserByEmail(email)
	if err != nil {
		if errors.Is(err, storage.ErrUserAlreadyExists) {
			log.Info("user already exists", sl.Err(err))
			return 0, ErrInvalidCredentials
		} else if !errors.Is(err, storage.ErrUserNotFound) {
			log.Error("failed to get user by email", sl.Err(err))
			return 0, ErrInternalServerError
		}
	}

	userId, err := a.userSaver.SaveUser(email, passHash)
	if err != nil {
		log.Error("failed to save user", sl.Err(err))
		return 0, ErrInternalServerError
	}

	log.Info("user registered successfully", slog.Int64("userId", userId))
	return userId, nil
}

func (a *Auth) IsAdmin(ctx context.Context, userId int64) (bool, error) {
	// Implement admin check logic here
	const op = "Auth.IsAdmin"
	log := a.log.With(slog.String("op", op), slog.Int64("userId", userId))

	isAdmin, err := a.userProvider.IsAdmin(userId)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			log.Info("user not found", sl.Err(err))
			return false, ErrInvalidCredentials
		}
		log.Error("failed to check if user is admin", sl.Err(err))
		return false, ErrInternalServerError
	}
	return isAdmin, nil
}
