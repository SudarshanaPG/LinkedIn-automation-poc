package stealth

import (
	"context"
	"math"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
)

type Point struct {
	X float64
	Y float64
}

func (c *Controller) MoveMouseHuman(ctx context.Context, page *rod.Page, from, to Point) (Point, error) {
	path := c.buildPath(from, to)
	for _, step := range path {
		if err := ctx.Err(); err != nil {
			return from, err
		}
		if err := page.Mouse.MoveTo(proto.Point{X: step.X, Y: step.Y}); err != nil {
			return from, err
		}
		time.Sleep(step.Delay)
		from = Point{X: step.X, Y: step.Y}
	}
	return to, nil
}

func (c *Controller) WanderMouse(ctx context.Context, page *rod.Page, from Point, width, height int) (Point, error) {
	if width <= 0 || height <= 0 {
		return from, nil
	}
	target := Point{
		X: float64(c.randomInt(30, width-30)),
		Y: float64(c.randomInt(30, height-30)),
	}
	return c.MoveMouseHuman(ctx, page, from, target)
}

type mouseStep struct {
	X     float64
	Y     float64
	Delay time.Duration
}

func (c *Controller) buildPath(from, to Point) []mouseStep {
	distance := math.Hypot(to.X-from.X, to.Y-from.Y)
	steps := clampInt(int(distance/8)+10, 18, 64)

	jitter := c.cfg.MouseMoveJitter
	if jitter < 0 {
		jitter = 0
	}
	control1 := Point{
		X: from.X + (to.X-from.X)*0.3 + float64(c.randomInt(-40-jitter, 40+jitter)),
		Y: from.Y + (to.Y-from.Y)*0.3 + float64(c.randomInt(-40-jitter, 40+jitter)),
	}
	control2 := Point{
		X: from.X + (to.X-from.X)*0.7 + float64(c.randomInt(-40-jitter, 40+jitter)),
		Y: from.Y + (to.Y-from.Y)*0.7 + float64(c.randomInt(-40-jitter, 40+jitter)),
	}

	path := make([]mouseStep, 0, steps+6)
	for i := 0; i <= steps; i++ {
		t := float64(i) / float64(steps)
		p := cubicBezier(from, control1, control2, to, t)
		delay := time.Duration(6+math.Sin(t*math.Pi)*9+float64(c.randomInt(0, 6))) * time.Millisecond
		path = append(path, mouseStep{X: p.X, Y: p.Y, Delay: delay})
	}

	overshoot := Point{
		X: to.X + float64(c.randomInt(-12, 12)),
		Y: to.Y + float64(c.randomInt(-12, 12)),
	}
	path = append(path, mouseStep{X: overshoot.X, Y: overshoot.Y, Delay: 14 * time.Millisecond})

	for i := 0; i < 3; i++ {
		jitter := Point{
			X: to.X + float64(c.randomInt(-2, 2)),
			Y: to.Y + float64(c.randomInt(-2, 2)),
		}
		path = append(path, mouseStep{X: jitter.X, Y: jitter.Y, Delay: 12 * time.Millisecond})
	}

	return path
}

func cubicBezier(p0, p1, p2, p3 Point, t float64) Point {
	u := 1 - t
	tt := t * t
	uu := u * u
	uuu := uu * u
	ttt := tt * t
	return Point{
		X: uuu*p0.X + 3*uu*t*p1.X + 3*u*tt*p2.X + ttt*p3.X,
		Y: uuu*p0.Y + 3*uu*t*p1.Y + 3*u*tt*p2.Y + ttt*p3.Y,
	}
}

func clampInt(value, min, max int) int {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}
