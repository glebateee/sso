package app

import (
	"log/slog"
	"sso/internal/app/grpcapp"
	"time"
)

type App struct {
	GRPCSrv *grpcapp.App
}

func New(
	log *slog.Logger,
	port int,
	storagePath string,
	tokenTTL time.Duration,
) *App {
	// init storage
	// init auth service
	grpcApp := grpcapp.New(log, port)
	return &App{
		GRPCSrv: grpcApp,
	}
}
