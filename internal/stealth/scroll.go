package stealth

import (
	"context"
	"math"
	"time"

	"github.com/go-rod/rod"
)

func (c *Controller) ScrollHuman(ctx context.Context, page *rod.Page, totalPx int) error {
	remaining := totalPx
	for math.Abs(float64(remaining)) > 0 {
		if err := ctx.Err(); err != nil {
			return err
		}
		step := c.ScrollStep()
		if step <= 0 {
			step = 220
		}
		if math.Abs(float64(step)) > math.Abs(float64(remaining)) {
			step = int(math.Abs(float64(remaining)))
		}
		if remaining < 0 {
			step = -step
		}
		steps := c.randomInt(3, 10)
		if err := page.Mouse.Scroll(0, float64(step), steps); err != nil {
			return err
		}
		remaining -= step
		time.Sleep(c.randomDuration(60, 220))

		if c.ShouldScrollBack() {
			back := c.randomInt(40, 120)
			if step < 0 {
				back = -back
			}
			if err := page.Mouse.Scroll(0, float64(-back), c.randomInt(2, 6)); err != nil {
				return err
			}
			time.Sleep(c.randomDuration(80, 240))
		}
	}
	return nil
}

func (c *Controller) HoverElement(ctx context.Context, page *rod.Page, current Point, el *rod.Element) (Point, error) {
	pt, err := el.WaitInteractable()
	if err != nil {
		return current, err
	}
	target := Point{X: pt.X, Y: pt.Y}
	next, err := c.MoveMouseHuman(ctx, page, current, target)
	if err != nil {
		return current, err
	}
	time.Sleep(c.HoverPause())
	return next, nil
}
