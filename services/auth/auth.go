package auth

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sso/domain/models"
	"sso/internal/lib/jwt"
	"sso/internal/lib/logger/sl"
	"sso/internal/storage"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type Auth struct {
	log          *slog.Logger
	userSaver    UserSaver
	userProvider UserProvider
	appProvider  AppProvider
	TokenTTL     time.Duration
}

type UserSaver interface {
	SaveUser(
		ctx context.Context,
		email string,
		passHash []byte,
	) (uid int64, err error)
}

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidAppID       = errors.New("invalid application id")
	ErrUserExists         = errors.New("user already exists")
)

type UserProvider interface {
	User(
		ctx context.Context,
		email string,
	) (user models.User, err error)
	IsAdmin(
		ctx context.Context,
		userID int64,
	) (isAdmin bool, err error)
}

type AppProvider interface {
	App(
		ctx context.Context,
		appID int,
	) (app models.App, err error)
}

func New(
	log *slog.Logger,
	userSaver UserSaver,
	userProvider UserProvider,
	appProvider AppProvider,
	tokenTTL time.Duration,
) *Auth {
	return &Auth{
		log:          log,
		userSaver:    userSaver,
		userProvider: userProvider,
		appProvider:  appProvider,
		TokenTTL:     tokenTTL,
	}
}

func (a *Auth) LoginUser(
	ctx context.Context,
	email,
	password string,
	appID int,
) (string, error) {
	const op = "auth.LoginUser"
	log := a.log.With(
		slog.String("op", op),
		slog.String("email", email),
	)

	user, err := a.userProvider.User(ctx, email)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			log.Warn("user not found", sl.Err(err))
			return "", fmt.Errorf("%s : %w", op, ErrInvalidCredentials)
		}
		log.Error("failed to get user", sl.Err(err))
		return "", fmt.Errorf("%s : %w", op, err)
	}

	if err := bcrypt.CompareHashAndPassword(user.PassHash, []byte(password)); err != nil {
		log.Error("password hash doesn't match", sl.Err(err))
		return "", fmt.Errorf("%s : %w", op, ErrInvalidCredentials)
	}

	app, err := a.appProvider.App(ctx, appID)
	if err != nil {
		return "", fmt.Errorf("%s : %w", op, err)
	}

	token, err := jwt.NewToken(user, app, a.TokenTTL)
	if err != nil {
		log.Error("failed to generate token", sl.Err(err))
		return "", fmt.Errorf("%s : %w", op, err)
	}
	return token, nil
}

func (a *Auth) RegisterUser(ctx context.Context, email, password string) (userID int64, err error) {
	const op = "auth.RegisterUser"
	log := a.log.With(
		slog.String("op", op),
		slog.String("email", email),
	)

	passHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {

		log.Error("failed to generate password hash", sl.Err(err))
		return 0, fmt.Errorf("%s : %w", op, err)
	}
	uid, err := a.userSaver.SaveUser(ctx, email, passHash)
	if err != nil {
		if errors.Is(err, storage.ErrUserExists) {
			log.Warn("user exists", sl.Err(err))
			return 0, fmt.Errorf("%s : %w", op, ErrUserExists)
		}
		log.Error("failed to save user", sl.Err(err))
		return 0, fmt.Errorf("%s : %w", op, err)
	}
	return uid, nil
}
func (a *Auth) IsAdminUser(ctx context.Context, userID int64) (bool, error) {
	const op = "auth.IsAdminUser"
	log := a.log.With(
		slog.String("op", op),
		slog.Int64("user_id", userID),
	)
	isAdmin, err := a.userProvider.IsAdmin(ctx, userID)
	if err != nil {
		if errors.Is(err, storage.ErrAppNotFound) {
			log.Warn("app not found ", sl.Err(err))
			return false, fmt.Errorf("%s : %w", op, ErrInvalidAppID)
		}
		return false, fmt.Errorf("%s : %w", op, err)
	}
	log.Info("user isAdmin", slog.Bool("is_admin", isAdmin))
	return isAdmin, nil
}
