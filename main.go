package main

import (
	"fmt"
	"math"
	"os"
	"sync"

	"github.com/cznic/mathutil"
	"github.com/veandco/go-sdl2/sdl"
)

func Map(n, start1, stop1, start2, stop2 int32) int32 {
	return ((n-start1)/(stop1-start1))*(stop2-start2) + start2
}

func Distance(a, b Vector2D) float64 {
	first := math.Pow(float64(a.x-b.x), 2)
	second := math.Pow(float64(a.y-b.y), 2)
	return math.Sqrt(first + second)
}

type Vector2D struct {
	x float64
	y float64
}

type Player struct {
	pos        Vector2D
	rotation   int32
	headingRay *Ray
	rays       []*Ray
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
	headingRadian := float64(player.rotation+30) * math.Pi / 180
	player.headingRay.pos.x = player.pos.x
	player.headingRay.pos.y = player.pos.y
	player.headingRay.dir.x = math.Cos(headingRadian) * 300
	player.headingRay.dir.y = math.Sin(headingRadian) * 300
	player.headingRay.show(renderer)
	for i, ray := range player.rays {
		radian := float64(player.rotation+int32(i)) * math.Pi / 180
		ray.dir.x = math.Cos(radian) * 300
		ray.dir.y = math.Sin(radian) * 300
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

var player Player
var walls []Boundary
var distances []float64

var runningMutex sync.Mutex

func run() int {
	var window *sdl.Window
	var secondWindow *sdl.Window
	var renderer *sdl.Renderer
	var secondRenderer *sdl.Renderer
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
		secondWindow, err = sdl.CreateWindow(
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
			secondWindow.Destroy()
		})
	}()

	sdl.Do(func() {
		renderer, err = sdl.CreateRenderer(
			window,
			-1,
			sdl.RENDERER_ACCELERATED,
		)
		secondRenderer, err = sdl.CreateRenderer(
			secondWindow,
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
			secondRenderer.Destroy()
		})
	}()

	sdl.Do(func() {
		renderer.Clear()
		secondRenderer.Clear()
	})

	running := true
	for running {
		sdl.Do(func() {
			for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
				switch t := event.(type) {
				case *sdl.QuitEvent:
					runningMutex.Lock()
					running = false
					runningMutex.Unlock()
				case *sdl.KeyboardEvent:
					if t.Keysym.Sym == sdl.K_ESCAPE {
						runningMutex.Lock()
						running = false
						runningMutex.Unlock()
					}
					if t.Keysym.Sym == sdl.K_RIGHT {
						player.rotation += 2
					}
					if t.Keysym.Sym == sdl.K_LEFT {
						player.rotation -= 2
					}
				}
			}
			x, y, _ := sdl.GetMouseState()
			player.move(float64(x), float64(y))

			renderer.Clear()
			renderer.SetDrawColor(0, 0, 0, 255)
			renderer.FillRect(&sdl.Rect{0, 0, WindowWidth, WindowHeight})

			secondRenderer.Clear()
			secondRenderer.SetDrawColor(0, 0, 0, 255)
			secondRenderer.FillRect(&sdl.Rect{0, 0, WindowWidth, WindowHeight})
		})

		// First View
		sdl.Do(func() {
			renderer.SetDrawColor(255, 255, 255, 255)
			for _, wall := range walls {
				wall.show(renderer)
			}
			player.show(renderer)
		})
		// Second View
		sdl.Do(func() {
			distances = make([]float64, len(player.rays))
			var playerHeadingHit *Vector2D
			for _, wall := range walls {
				playerHeadingHit = player.headingRay.cast(wall)
				if playerHeadingHit != nil {
					break
				}
			}
			for i, ray := range player.rays {
				for _, wall := range walls {
					hit := ray.cast(wall)
					if hit != nil && playerHeadingHit != nil {
						eucDistance := Distance(ray.pos, *hit)
						dotProduct := (hit.x * playerHeadingHit.x) + (hit.y * playerHeadingHit.y)
						aMag := math.Sqrt(math.Pow(hit.x, 2) + math.Pow(hit.y, 2))
						bMag := math.Sqrt(math.Pow(playerHeadingHit.x, 2) + math.Pow(playerHeadingHit.y, 2))
						angle := dotProduct / (math.Abs(aMag) * math.Abs(bMag))

						radian := float64(angle) * math.Pi / 180
						distances[i] = math.Cos(radian) * eucDistance
						fmt.Fprintf(os.Stdout, "Euc Distance: %v\n", eucDistance)
						fmt.Fprintf(os.Stdout, "Fish Distance: %v\n", math.Cos(radian)*eucDistance)

						renderer.SetDrawColor(255, 0, 0, 255)
						renderer.DrawLine(
							int32(ray.pos.x),
							int32(ray.pos.y),
							int32(hit.x),
							int32(hit.y),
						)
						break
					} else {
						distances[i] = WindowWidth
					}
				}
			}
			w := WindowWidth / len(distances)
			for i, p := range distances {
				if p == WindowWidth {
					secondRenderer.SetDrawColor(0, 0, 0, 255)
					secondRenderer.FillRect(&sdl.Rect{int32(i * w), 0, int32(w), WindowHeight})
				} else {
					clamped := uint8(mathutil.ClampInt32(int32(p), 0, 255))
					secondRenderer.SetDrawColor(255-clamped, 255-clamped, 255-clamped, 255)
					secondRenderer.FillRect(&sdl.Rect{int32(i * w), WindowHeight / 4, int32(w), 255 - int32(clamped)})
				}
			}
		})

		sdl.Do(func() {
			renderer.Present()
			secondRenderer.Present()
			sdl.Delay(1000 / FrameRate)
		})
	}

	return 0
}

func main() {
	walls = append(walls, Boundary{
		a: Vector2D{x: 300, y: 100},
		b: Vector2D{x: 300, y: 300},
	})
	walls = append(walls, Boundary{
		a: Vector2D{x: 300, y: 100},
		b: Vector2D{x: 100, y: 100},
	})
	walls = append(walls, Boundary{
		a: Vector2D{x: 100, y: 100},
		b: Vector2D{x: 100, y: 300},
	})
	walls = append(walls, Boundary{
		a: Vector2D{x: 100, y: 300},
		b: Vector2D{x: 300, y: 300},
	})

	var rays []*Ray
	for i := -30; i < 30; i += 1 {
		radian := float64(i) * math.Pi / 180
		rays = append(rays, &Ray{
			pos: Vector2D{x: 100, y: 200},
			dir: Vector2D{x: math.Cos(radian) * 300, y: math.Sin(radian) * 300},
		})
	}
	player.headingRay = &Ray{
		pos: Vector2D{x: 100, y: 200},
	}
	player.rotation = -30
	player.pos = Vector2D{x: 100, y: 200}
	player.rays = rays

	var exitcode int
	sdl.Main(func() {
		exitcode = run()
	})
	os.Exit(exitcode)
}
