package stealth

import (
	"context"
	"math/rand"
	"time"

	"linkedin-automation-poc/internal/config"
	"linkedin-automation-poc/internal/logger"
)

type Controller struct {
	cfg config.StealthConfig
	rng *rand.Rand
	log *logger.Logger
}

func New(cfg config.StealthConfig, log *logger.Logger) *Controller {
	return &Controller{
		cfg: cfg,
		rng: rand.New(rand.NewSource(time.Now().UnixNano())),
		log: log,
	}
}

func (c *Controller) Think(ctx context.Context) {
	c.sleep(ctx, c.cfg.ThinkTimeMinMs, c.cfg.ThinkTimeMaxMs)
}

func (c *Controller) ActionPause(ctx context.Context) {
	c.sleep(ctx, c.cfg.ActionIntervalMinMs, c.cfg.ActionIntervalMaxMs)
}

func (c *Controller) TypingDelay() time.Duration {
	return c.randomDuration(c.cfg.TypingDelayMinMs, c.cfg.TypingDelayMaxMs)
}

func (c *Controller) HoverPause() time.Duration {
	return c.randomDuration(c.cfg.HoverPauseMinMs, c.cfg.HoverPauseMaxMs)
}

func (c *Controller) ScrollStep() int {
	return c.randomInt(c.cfg.ScrollStepMin, c.cfg.ScrollStepMax)
}

func (c *Controller) ShouldScrollBack() bool {
	return c.rollPercent(c.cfg.ScrollBackChance)
}

func (c *Controller) ShouldWanderMouse() bool {
	return c.rollPercent(c.cfg.MouseWanderChance)
}

func (c *Controller) ShouldTypo() bool {
	return c.rollPercent(c.cfg.TypoChance)
}

func (c *Controller) RandomizeViewport(width, height int) (int, int) {
	if !c.cfg.RandomizeViewport || c.cfg.ViewportVariancePx <= 0 {
		return width, height
	}
	delta := c.rng.Intn(c.cfg.ViewportVariancePx*2+1) - c.cfg.ViewportVariancePx
	return width + delta, height + delta
}

func (c *Controller) randomDuration(minMs, maxMs int) time.Duration {
	if maxMs <= minMs {
		return time.Duration(minMs) * time.Millisecond
	}
	value := c.randomInt(minMs, maxMs)
	return time.Duration(value) * time.Millisecond
}

func (c *Controller) randomInt(min, max int) int {
	if max <= min {
		return min
	}
	return min + c.rng.Intn(max-min+1)
}

func (c *Controller) rollPercent(chance int) bool {
	if chance <= 0 {
		return false
	}
	if chance >= 100 {
		return true
	}
	return c.rng.Intn(100) < chance
}

func (c *Controller) sleep(ctx context.Context, minMs, maxMs int) {
	delay := c.randomDuration(minMs, maxMs)
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
	case <-timer.C:
	}
}
