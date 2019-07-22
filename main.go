package main

import (
	"fmt"
	"math"
	"os"
	"sync"

	"github.com/veandco/go-sdl2/sdl"
)

type Vector2D struct {
	x float64
	y float64
}

type Player struct {
	pos  Vector2D
	rays []*Ray
}

func (player *Player) move(x, y float64) {
	player.pos.x = x
	player.pos.y = y
	for _, ray := range player.rays {
		ray.pos.x = x
		ray.pos.y = y
	}
}

func (player *Player) show(renderer *sdl.Renderer) {
	for _, ray := range player.rays {
		ray.show(renderer)
	}
}

type Boundary struct {
	a Vector2D
	b Vector2D
}

func (boundary *Boundary) show(renderer *sdl.Renderer) {
	renderer.DrawLine(
		int32(boundary.a.x),
		int32(boundary.a.y),
		int32(boundary.b.x),
		int32(boundary.b.y),
	)
}

type Ray struct {
	pos Vector2D
	dir Vector2D
}

func (ray *Ray) show(renderer *sdl.Renderer) {
	renderer.DrawLine(
		int32(ray.pos.x),
		int32(ray.pos.y),
		int32(ray.pos.x+ray.dir.x),
		int32(ray.pos.y+ray.dir.y),
	)
}

func (ray *Ray) cast(wall Boundary) *Vector2D {
	x1 := ray.pos.x
	y1 := ray.pos.y
	x2 := ray.pos.x + ray.dir.x
	y2 := ray.pos.y + ray.dir.y

	x3 := wall.a.x
	y3 := wall.a.y
	x4 := wall.b.x
	y4 := wall.b.y

	uA := ((x4-x3)*(y1-y3) - (y4-y3)*(x1-x3)) / ((y4-y3)*(x2-x1) - (x4-x3)*(y2-y1))
	uB := ((x2-x1)*(y1-y3) - (y2-y1)*(x1-x3)) / ((y4-y3)*(x2-x1) - (x4-x3)*(y2-y1))

	if uA >= 0 && uA <= 1 && uB >= 0 && uB <= 1 {
		return &Vector2D{x: x1 + (uA * (x2 - x1)), y: y1 + (uA * (y2 - y1))}
	}
	return nil
}

const (
	WindowTitle  = "Raycasting"
	WindowWidth  = 800
	WindowHeight = 600
	FrameRate    = 60
)

var wall Boundary
var player Player

var runningMutex sync.Mutex

func run() int {
	var window *sdl.Window
	var renderer *sdl.Renderer
	var err error

	sdl.Do(func() {
		window, err = sdl.CreateWindow(
			WindowTitle,
			sdl.WINDOWPOS_UNDEFINED,
			sdl.WINDOWPOS_UNDEFINED,
			WindowWidth,
			WindowHeight,
			sdl.WINDOW_OPENGL,
		)
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create window: %s\n", err)
		return 1
	}
	defer func() {
		sdl.Do(func() {
			window.Destroy()
		})
	}()

	sdl.Do(func() {
		renderer, err = sdl.CreateRenderer(
			window,
			-1,
			sdl.RENDERER_ACCELERATED,
		)
	})
	if err != nil {
		fmt.Fprint(os.Stderr, "Failed to create renderer: %s\n", err)
		return 2
	}
	defer func() {
		sdl.Do(func() {
			renderer.Destroy()
		})
	}()

	sdl.Do(func() {
		renderer.Clear()
	})

	running := true
	for running {
		sdl.Do(func() {
			for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
				switch event.(type) {
				case *sdl.QuitEvent:
					runningMutex.Lock()
					running = false
					runningMutex.Unlock()
				}
			}
			x, y, _ := sdl.GetMouseState()
			player.move(float64(x), float64(y))

			renderer.Clear()
			renderer.SetDrawColor(0, 0, 0, 255)
			renderer.FillRect(&sdl.Rect{0, 0, WindowWidth, WindowHeight})
		})

		sdl.Do(func() {
			renderer.SetDrawColor(255, 255, 255, 255)
			wall.show(renderer)
			player.show(renderer)
			for _, ray := range player.rays {
				hit := ray.cast(wall)
				if hit != nil {
					renderer.SetDrawColor(255, 0, 0, 255)
					renderer.DrawLine(
						int32(ray.pos.x),
						int32(ray.pos.y),
						int32(hit.x),
						int32(hit.y),
					)
				}
			}
		})

		sdl.Do(func() {
			renderer.Present()
			sdl.Delay(1000 / FrameRate)
		})
	}

	return 0
}

func main() {
	wall = Boundary{a: Vector2D{x: 300, y: 100}, b: Vector2D{x: 300, y: 300}}

	var rays []*Ray
	for i := 0; i < 60; i += 2 {
		radian := float64(i) * math.Pi / 180
		rays = append(rays, &Ray{
			pos: Vector2D{x: 100, y: 200},
			dir: Vector2D{x: math.Cos(radian) * 300, y: math.Sin(radian) * 300},
		})
	}
	player.pos = Vector2D{x: 100, y: 200}
	player.rays = rays

	var exitcode int
	sdl.Main(func() {
		exitcode = run()
	})
	os.Exit(exitcode)
}
