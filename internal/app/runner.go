package app

import (
	"context"

	"github.com/go-rod/rod/lib/proto"

	"linkedin-automation-poc/internal/auth"
	"linkedin-automation-poc/internal/browser"
	"linkedin-automation-poc/internal/config"
	"linkedin-automation-poc/internal/connect"
	"linkedin-automation-poc/internal/limits"
	"linkedin-automation-poc/internal/logger"
	"linkedin-automation-poc/internal/messaging"
	"linkedin-automation-poc/internal/search"
	"linkedin-automation-poc/internal/stealth"
	"linkedin-automation-poc/internal/storage"
)

type Runner struct {
	cfg   config.Config
	log   *logger.Logger
	store *storage.Storage
	dry   bool
}

func NewRunner(cfg config.Config, log *logger.Logger, store *storage.Storage, dryRun bool) *Runner {
	return &Runner{
		cfg:   cfg,
		log:   log,
		store: store,
		dry:   dryRun,
	}
}

func (r *Runner) Run(ctx context.Context) error {
	r.log.Info("runner started", map[string]any{"headless": r.cfg.Stealth.Headless})

	scheduler, err := stealth.NewScheduler(r.cfg.Schedule)
	if err != nil {
		return err
	}
	if err := scheduler.Enforce(ctx, r.log); err != nil {
		return err
	}

	stealthCtrl := stealth.New(r.cfg.Stealth, r.log)
	session, err := browser.NewSession(ctx, r.cfg, r.log, stealthCtrl)
	if err != nil {
		return err
	}
	defer func() {
		_ = session.Close()
	}()

	if r.dry {
		return r.runDry(ctx, session)
	}

	authenticator := auth.New(r.cfg.Auth, r.log)
	if err := authenticator.Login(ctx, session); err != nil {
		return err
	}

	limiter := limits.New(r.cfg, r.store, r.log)

	finder := search.New(r.cfg.Search, r.log)
	leads, err := finder.Search(ctx, session)
	if err != nil {
		return err
	}
	r.log.Info("search complete", map[string]any{"leads": len(leads)})

	connector := connect.New(r.cfg.Connect, r.store, r.log)
	if err := connector.SendRequests(ctx, session, limiter, leads); err != nil {
		return err
	}

	messenger := messaging.New(r.cfg.Messaging, r.store, r.log)
	if err := messenger.SyncAccepted(ctx, session); err != nil {
		return err
	}
	if err := messenger.SendFollowUps(ctx, session, limiter); err != nil {
		return err
	}

	r.log.Info("runner finished", nil)
	return nil
}

func (r *Runner) runDry(ctx context.Context, session *browser.Session) error {
	r.log.Warn("dry-run enabled: not logging into LinkedIn and not performing any platform actions", nil)
	if err := session.Page.SetDocumentContent(demoHTML); err != nil {
		return err
	}

	session.Stealth.Think(ctx)

	inputEl, err := session.Page.Element("#demo-input")
	if err == nil {
		_ = session.HoverElement(ctx, inputEl)
		_ = inputEl.Click(proto.InputMouseButtonLeft, 1)
		_ = session.Stealth.TypeHuman(ctx, session.Page, "Hello! This is a local dry-run demo.\nTyping includes pauses, typos, and corrections.")
	}
	session.Stealth.ActionPause(ctx)

	btn, err := session.Page.Element("#demo-button")
	if err == nil {
		_ = session.HoverElement(ctx, btn)
		_ = btn.Click(proto.InputMouseButtonLeft, 1)
	}
	session.Stealth.ActionPause(ctx)

	_ = session.Stealth.ScrollHuman(ctx, session.Page, session.Stealth.ScrollStep()*4)
	session.Stealth.ActionPause(ctx)
	_ = session.Stealth.ScrollHuman(ctx, session.Page, -session.Stealth.ScrollStep()*2)

	session.MaybeWander(ctx)
	r.log.Info("dry-run finished (local demo page)", nil)
	return nil
}
