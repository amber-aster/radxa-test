package main

import (
	"flag"
	"fmt"
	"image/color"
	"log"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

const (
	screenW = 1000
	screenH = 600
)

// Game holds all mutable app state.
// Ebiten calls Update many times per second, then Draw whenever a frame is rendered.
type Game struct {
	vehicle    VehicleState
	source     KeyboardVehicleSource
	fullscreen bool
}

// VehicleState is the clean boundary between vehicle data and graphics.
// A real dashboard should make drawing code read this, not raw CAN messages.
type VehicleState struct {
	SpeedKPH float64
	RPM      float64
	Gear     int
	Warning  bool
}

// KeyboardVehicleSource is our temporary fake data source.
// Later this could become CANVehicleSource, OBDVehicleSource, or ReplayLogSource.
type KeyboardVehicleSource struct{}

func (KeyboardVehicleSource) Update(v *VehicleState) {
	if ebiten.IsKeyPressed(ebiten.KeyArrowUp) {
		v.SpeedKPH += 1.2
		v.RPM += 120
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowDown) {
		v.SpeedKPH -= 1.8
		v.RPM -= 180
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowRight) && v.Gear < 6 {
		v.Gear = min(v.Gear+1, 6)
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowLeft) && v.Gear > 1 {
		v.Gear = max(v.Gear-1, 1)
	}
	if ebiten.IsKeyPressed(ebiten.KeySpace) {
		v.Warning = true
	}
	if ebiten.IsKeyPressed(ebiten.KeyC) {
		v.Warning = false
	}

	v.SpeedKPH = clamp(v.SpeedKPH, 0, 260)
	v.RPM = clamp(v.RPM, 700, 8500)

	// Simple decay makes it feel alive even without real vehicle data.
	v.SpeedKPH = max(0, v.SpeedKPH-0.05)
	v.RPM = max(700, v.RPM-20)
}

func main() {
	fullscreen := flag.Bool("fullscreen", true, "start fullscreen")
	flag.Parse()

	ebiten.SetWindowSize(screenW, screenH)
	ebiten.SetWindowTitle("Ebiten Car Dashboard Demo")
	ebiten.SetFullscreen(*fullscreen)
	ebiten.SetTPS(60)
	ebiten.SetVsyncEnabled(true)

	game := &Game{vehicle: VehicleState{Gear: 1}, fullscreen: *fullscreen}
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}

// Update is where input, simulation, networking, CAN-bus reads, etc. normally happen.
// Keep it deterministic and avoid slow work here.
func (g *Game) Update() error {
	if ebiten.IsKeyPressed(ebiten.KeyEscape) || ebiten.IsKeyPressed(ebiten.KeyQ) {
		return ebiten.Termination
	}
	if inputOnce(ebiten.KeyF) {
		g.fullscreen = !g.fullscreen
		ebiten.SetFullscreen(g.fullscreen)
	}

	g.source.Update(&g.vehicle)
	return nil
}

var keyLatch = map[ebiten.Key]bool{}

func inputOnce(k ebiten.Key) bool {
	down := ebiten.IsKeyPressed(k)
	fired := down && !keyLatch[k]
	keyLatch[k] = down
	return fired
}

// Draw is where all rendering happens. It receives a fresh frame buffer each frame.
func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{8, 10, 14, 255})

	drawCard(screen, 40, 40, 920, 520)
	drawSpeedGauge(screen, 310, 310, 210, g.vehicle.SpeedKPH)
	drawRPMBar(screen, 590, 150, 320, 34, g.vehicle.RPM)
	drawStatus(screen, g.vehicle)

	ebitenutil.DebugPrintAt(screen, "Controls: Up/Down speed+rpm | Left/Right gear | Space warning | C clear | F fullscreen | Q quit", 52, 545)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("TPS: %.0f FPS: %.0f", ebiten.ActualTPS(), ebiten.ActualFPS()), 790, 545)
}

// Layout defines the logical canvas size. Ebiten scales this to the real window.
func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenW, screenH
}

func drawCard(dst *ebiten.Image, x, y, w, h float32) {
	vector.DrawFilledRect(dst, x, y, w, h, color.RGBA{18, 22, 30, 255}, false)
	vector.StrokeRect(dst, x, y, w, h, 2, color.RGBA{60, 72, 92, 255}, false)
}

func drawSpeedGauge(dst *ebiten.Image, cx, cy, radius float32, speed float64) {
	vector.StrokeCircle(dst, cx, cy, radius, 8, color.RGBA{45, 55, 70, 255}, false)

	start := math.Pi * 0.78
	end := math.Pi * 2.22
	angle := start + (end-start)*(speed/260)

	for i := 0; i <= 26; i++ {
		a := start + (end-start)*float64(i)/26
		inner := radius - 18
		outer := radius - 2
		x1 := cx + float32(math.Cos(a))*inner
		y1 := cy + float32(math.Sin(a))*inner
		x2 := cx + float32(math.Cos(a))*outer
		y2 := cy + float32(math.Sin(a))*outer
		vector.StrokeLine(dst, x1, y1, x2, y2, 3, color.RGBA{95, 115, 145, 255}, false)
	}

	needleX := cx + float32(math.Cos(angle))*(radius-45)
	needleY := cy + float32(math.Sin(angle))*(radius-45)
	vector.StrokeLine(dst, cx, cy, needleX, needleY, 6, color.RGBA{0, 220, 255, 255}, false)
	vector.DrawFilledCircle(dst, cx, cy, 10, color.RGBA{220, 245, 255, 255}, false)

	ebitenutil.DebugPrintAt(dst, fmt.Sprintf("%03.0f", speed), int(cx)-42, int(cy)+58)
	ebitenutil.DebugPrintAt(dst, "km/h", int(cx)-18, int(cy)+82)
}

func drawRPMBar(dst *ebiten.Image, x, y, w, h float32, rpm float64) {
	vector.StrokeRect(dst, x, y, w, h, 2, color.RGBA{80, 90, 110, 255}, false)
	fill := w * float32((rpm-700)/(8500-700))
	barColor := color.RGBA{80, 230, 120, 255}
	if rpm > 6500 {
		barColor = color.RGBA{255, 70, 60, 255}
	}
	vector.DrawFilledRect(dst, x, y, fill, h, barColor, false)
	ebitenutil.DebugPrintAt(dst, fmt.Sprintf("RPM %.0f", rpm), int(x), int(y)-24)
}

func drawStatus(dst *ebiten.Image, v VehicleState) {
	ebitenutil.DebugPrintAt(dst, fmt.Sprintf("GEAR %d", v.Gear), 600, 230)
	ebitenutil.DebugPrintAt(dst, "Dashboard apps are usually: state -> Update -> Draw", 600, 280)
	ebitenutil.DebugPrintAt(dst, "Rendering reads VehicleState, not keyboard/CAN directly.", 600, 305)

	if v.Warning {
		vector.DrawFilledCircle(dst, 725, 390, 42, color.RGBA{255, 45, 35, 255}, false)
		ebitenutil.DebugPrintAt(dst, "WARN", 704, 384)
	} else {
		vector.StrokeCircle(dst, 725, 390, 42, 3, color.RGBA{70, 80, 95, 255}, false)
		ebitenutil.DebugPrintAt(dst, "OK", 716, 384)
	}
}

func clamp(v, lo, hi float64) float64 {
	return min(max(v, lo), hi)
}
