package app

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	identitymod "github.com/go-mentorship-platform/backend/internal/identity"
	profilemod "github.com/go-mentorship-platform/backend/internal/profile"
	achievementmod "github.com/go-mentorship-platform/backend/internal/achievement"
	bonusmod "github.com/go-mentorship-platform/backend/internal/bonus"
	oneononemod "github.com/go-mentorship-platform/backend/internal/oneonone"
	interviewmod "github.com/go-mentorship-platform/backend/internal/interview"
	finalcheckmod "github.com/go-mentorship-platform/backend/internal/finalcheck"
	calendarmod "github.com/go-mentorship-platform/backend/internal/calendar"
	activitymod "github.com/go-mentorship-platform/backend/internal/activity"
	notificationmod "github.com/go-mentorship-platform/backend/internal/notification"
	progressmod "github.com/go-mentorship-platform/backend/internal/progress"
	roadmapmod "github.com/go-mentorship-platform/backend/internal/roadmap"
	usermod "github.com/go-mentorship-platform/backend/internal/user"
	"github.com/go-mentorship-platform/backend/internal/platform/authorization"
	"github.com/go-mentorship-platform/backend/internal/platform/httpserver/routes"
	"github.com/go-mentorship-platform/backend/internal/platform/config"
	"github.com/go-mentorship-platform/backend/internal/platform/eventbus"
	"github.com/go-mentorship-platform/backend/internal/platform/httpserver"
	"github.com/go-mentorship-platform/backend/internal/platform/httpserver/health"
	"github.com/go-mentorship-platform/backend/internal/platform/postgres"
	redisclient "github.com/go-mentorship-platform/backend/internal/platform/redis"
)

// App is the composition root: wires infrastructure and HTTP delivery.
type App struct {
	Config config.Config
	Logger *slog.Logger

	Postgres *postgres.Pool
	Redis    *redisclient.Client
	EventBus *eventbus.Bus
	Authz    *authorization.Service
	Identity *identitymod.Module
	User     *usermod.Module
	Profile  *profilemod.Module
	Roadmap  *roadmapmod.Module
	Progress    *progressmod.Module
	Achievement *achievementmod.Module
	Bonus       *bonusmod.Module
	OneOnOne    *oneononemod.Module
	Interview   *interviewmod.Module
	FinalCheck  *finalcheckmod.Module
	Calendar    *calendarmod.Module
	Activity    *activitymod.Module
	Notification *notificationmod.Module

	HTTPServer *httpserver.Server
}

// New builds the application container and connects external dependencies.
func New(ctx context.Context) (*App, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}

	log := newLogger(cfg.App.LogLevel)

	pgPool, err := postgres.New(ctx, cfg.Database)
	if err != nil {
		return nil, fmt.Errorf("postgres: %w", err)
	}

	redisClient, err := redisclient.New(ctx, cfg.Redis)
	if err != nil {
		pgPool.Close()
		return nil, fmt.Errorf("redis: %w", err)
	}

	bus := eventbus.New()

	identityModule, err := identitymod.NewModule(cfg.Auth, pgPool)
	if err != nil {
		pgPool.Close()
		_ = redisClient.Close()
		return nil, fmt.Errorf("identity: %w", err)
	}

	userModule := usermod.NewModule(pgPool, identityModule.HTTP)
	profileModule := profilemod.NewModule(pgPool, identityModule.HTTP, userModule.Assignments)
	roadmapModule := roadmapmod.NewModule(pgPool, identityModule.HTTP)
	progressModule := progressmod.NewModule(pgPool, identityModule.HTTP)
	achievementModule := achievementmod.NewModule(pgPool, identityModule.HTTP, bus, log)
	bonusModule := bonusmod.NewModule(pgPool, identityModule.HTTP, log)
	oneOnOneModule := oneononemod.NewModule(pgPool, identityModule.HTTP, bonusModule.Ledger)
	interviewModule := interviewmod.NewModule(pgPool, identityModule.HTTP)
	finalCheckModule := finalcheckmod.NewModule(pgPool, identityModule.HTTP)
	calendarModule := calendarmod.NewModule(pgPool, identityModule.HTTP)
	activityModule := activitymod.NewModule(pgPool, identityModule.HTTP, log)
	notificationModule := notificationmod.NewModule(pgPool, identityModule.HTTP, log)
	userModule.HTTP.ExtraBuddyRegister = activityModule.HTTP.RegisterBuddy

	authz := authorization.NewService(userModule.Assignments)

	healthHandler := health.NewHandler(pgPool, redisClient)

	router := httpserver.NewRouter(httpserver.Deps{
		Config: cfg,
		Logger: log,
		Health: healthHandler,
		V1: routes.V1Deps{
			Identity: identityModule.HTTP,
			User:     userModule.HTTP,
			Profile:  profileModule.HTTP,
			Roadmap:  roadmapModule.HTTP,
			Progress:    progressModule.HTTP,
			Achievement: achievementModule.HTTP,
			Bonus:       bonusModule.HTTP,
			OneOnOne:    oneOnOneModule.HTTP,
			Interview:   interviewModule.HTTP,
			FinalCheck:  finalCheckModule.HTTP,
			Calendar:    calendarModule.HTTP,
			Activity:    activityModule.HTTP,
			Notification: notificationModule.HTTP,
		},
	})

	httpServer := httpserver.New(cfg.HTTP, log, router)

	return &App{
		Config:     cfg,
		Logger:     log,
		Postgres:   pgPool,
		Redis:      redisClient,
		EventBus:   bus,
		Authz:      authz,
		Identity:   identityModule,
		User:       userModule,
		Profile:    profileModule,
		Roadmap:    roadmapModule,
		Progress:    progressModule,
		Achievement: achievementModule,
		Bonus:       bonusModule,
		OneOnOne:    oneOnOneModule,
		Interview:   interviewModule,
		FinalCheck:  finalCheckModule,
		Calendar:    calendarModule,
		Activity:    activityModule,
		Notification: notificationModule,
		HTTPServer:  httpServer,
	}, nil
}

// Run starts the HTTP server until the context is cancelled.
func (a *App) Run(ctx context.Context) error {
	if a.Achievement != nil {
		go a.Achievement.RunOutboxWorker(ctx)
	}
	if a.Bonus != nil {
		go a.Bonus.RunOutboxWorker(ctx)
	}
	if a.Activity != nil {
		go a.Activity.RunOutboxWorker(ctx)
	}
	if a.Notification != nil {
		go a.Notification.RunOutboxWorker(ctx)
	}
	return a.HTTPServer.Start(ctx)
}

// Close releases infrastructure resources.
func (a *App) Close() {
	if a.Postgres != nil {
		a.Postgres.Close()
	}
	if a.Redis != nil {
		_ = a.Redis.Close()
	}
}

func newLogger(level string) *slog.Logger {
	var lvl slog.Level
	switch level {
	case "debug":
		lvl = slog.LevelDebug
	case "warn":
		lvl = slog.LevelWarn
	case "error":
		lvl = slog.LevelError
	default:
		lvl = slog.LevelInfo
	}

	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: lvl})
	return slog.New(handler)
}
