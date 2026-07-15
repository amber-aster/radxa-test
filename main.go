package main

import (
	"flag"
	"fmt"
	"image/color"
	"log"
	"math"
	"runtime"
	"strings"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

const (
	logicalW = 1280
	logicalH = 720
	graphN   = 240
)

type game struct {
	start       time.Time
	last        time.Time
	frameMS     float64
	maxFrameMS  float64
	stress      bool
	showHelp    bool
	fullscreen  bool
	history     [graphN]float64
	historyHead int
}

func main() {
	fullscreen := flag.Bool("fullscreen", true, "start fullscreen")
	stress := flag.Bool("stress", true, "draw extra load")
	flag.Parse()

	ebiten.SetWindowTitle("Radxa Ebiten Dashboard Render Test")
	ebiten.SetWindowSize(logicalW, logicalH)
	ebiten.SetFullscreen(*fullscreen)
	ebiten.SetTPS(60)
	ebiten.SetVsyncEnabled(true)

	g := &game{start: time.Now(), last: time.Now(), stress: *stress, fullscreen: *fullscreen, showHelp: true}
	if err := ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}
}

func (g *game) Layout(int, int) (int, int) { return logicalW, logicalH }

func (g *game) Update() error {
	now := time.Now()
	g.frameMS = float64(now.Sub(g.last).Microseconds()) / 1000
	g.last = now
	if g.frameMS > g.maxFrameMS || time.Since(g.start) < 2*time.Second {
		g.maxFrameMS = g.frameMS
	}
	g.history[g.historyHead] = g.frameMS
	g.historyHead = (g.historyHead + 1) % graphN

	if ebiten.IsKeyPressed(ebiten.KeyEscape) || ebiten.IsKeyPressed(ebiten.KeyQ) {
		return ebiten.Termination
	}
	if inputOnce(ebiten.KeyF) {
		g.fullscreen = !g.fullscreen
		ebiten.SetFullscreen(g.fullscreen)
	}
	if inputOnce(ebiten.KeyS) {
		g.stress = !g.stress
	}
	if inputOnce(ebiten.KeyH) {
		g.showHelp = !g.showHelp
	}
	return nil
}

var keyLatch = map[ebiten.Key]bool{}

func inputOnce(k ebiten.Key) bool {
	down := ebiten.IsKeyPressed(k)
	fired := down && !keyLatch[k]
	keyLatch[k] = down
	return fired
}

func (g *game) Draw(dst *ebiten.Image) {
	t := time.Since(g.start).Seconds()
	drawBackground(dst, t)
	drawGrid(dst, t)

	rpm := 900 + 5200*(0.5+0.5*math.Sin(t*0.85))
	speed := 12 + 118*(0.5+0.5*math.Sin(t*0.32-0.8))
	temp := 72 + 34*(0.5+0.5*math.Sin(t*0.18))
	boost := -8 + 20*(0.5+0.5*math.Sin(t*1.25+1.4))

	drawGauge(dst, 310, 365, 230, "RPM", rpm, 0, 7000, "x1", col(55, 220, 255, 255))
	drawGauge(dst, 970, 365, 230, "MPH", speed, 0, 160, "", col(255, 190, 60, 255))
	drawSmallCard(dst, 510, 86, 260, 130, "COOLANT", fmt.Sprintf("%.0f C", temp), temp > 100)
	drawSmallCard(dst, 510, 234, 260, 130, "MAP/BOOST", fmt.Sprintf("%.1f psi", boost), boost > 8)
	drawSmallCard(dst, 510, 382, 260, 130, "AFR", fmt.Sprintf("%.2f", 13.8+math.Sin(t*1.7)*1.4), false)
	drawSmallCard(dst, 510, 530, 260, 104, "VOLTAGE", fmt.Sprintf("%.1f V", 13.8+math.Sin(t*2.2)*0.4), false)

	drawFrameGraph(dst, 32, 596, 420, 90, g)
	if g.stress {
		drawStress(dst, t)
	}
	drawMetrics(dst, g)
}

func drawBackground(dst *ebiten.Image, t float64) {
	dst.Fill(col(5, 7, 13, 255))
	for i := 0; i < 28; i++ {
		x := float32(math.Mod(float64(i*67)+t*22, logicalW+120) - 60)
		y := float32(40 + i*23%logicalH)
		vector.DrawFilledCircle(dst, x, y, float32(40+i%6*18), col(10, 35, 55, 25), false)
	}
}

func drawGrid(dst *ebiten.Image, t float64) {
	for x := -80; x < logicalW+80; x += 40 {
		fx := float32(x) + float32(math.Mod(t*18, 40))
		vector.StrokeLine(dst, fx, 0, fx-300, logicalH, 1, col(30, 70, 90, 55), false)
	}
	for y := 0; y < logicalH; y += 40 {
		vector.StrokeLine(dst, 0, float32(y), logicalW, float32(y), 1, col(20, 45, 65, 45), false)
	}
}

func drawGauge(dst *ebiten.Image, cx, cy, r float32, label string, val, min, max float64, suffix string, accent color.Color) {
	vector.DrawFilledCircle(dst, cx, cy, r+18, col(2, 8, 15, 210), false)
	for i := 0; i <= 60; i++ {
		ang := math.Pi*0.78 + float64(i)/60*math.Pi*1.44
		inner := r - 18
		if i%5 == 0 { inner = r - 34 }
		x1, y1 := cx+float32(math.Cos(ang))*inner, cy+float32(math.Sin(ang))*inner
		x2, y2 := cx+float32(math.Cos(ang))*r, cy+float32(math.Sin(ang))*r
		vector.StrokeLine(dst, x1, y1, x2, y2, 2, col(130, 190, 210, 190), false)
	}
	pct := clamp((val-min)/(max-min), 0, 1)
	end := math.Pi*0.78 + pct*math.Pi*1.44
	for a := math.Pi * 0.78; a < end; a += 0.018 {
		vector.StrokeLine(dst, cx+float32(math.Cos(a))*float32(r-62), cy+float32(math.Sin(a))*float32(r-62), cx+float32(math.Cos(a))*float32(r-48), cy+float32(math.Sin(a))*float32(r-48), 5, accent, false)
	}
	needle := r - 72
	vector.StrokeLine(dst, cx, cy, cx+float32(math.Cos(end))*needle, cy+float32(math.Sin(end))*needle, 8, col(245, 250, 255, 255), false)
	vector.DrawFilledCircle(dst, cx, cy, 15, accent, false)
	ebitenutil.DebugPrintAt(dst, label, int(cx)-28, int(cy)-42)
	ebitenutil.DebugPrintAt(dst, fmt.Sprintf("%04.0f %s", val, suffix), int(cx)-54, int(cy)+18)
}

func drawSmallCard(dst *ebiten.Image, x, y, w, h float32, title, value string, warn bool) {
	c := col(12, 24, 34, 230)
	if warn { c = col(95, 18, 18, 235) }
	vector.DrawFilledRect(dst, x, y, w, h, c, false)
	vector.StrokeRect(dst, x, y, w, h, 2, col(85, 160, 190, 180), false)
	ebitenutil.DebugPrintAt(dst, title, int(x)+18, int(y)+18)
	ebitenutil.DebugPrintAt(dst, value, int(x)+18, int(y)+62)
}

func drawFrameGraph(dst *ebiten.Image, x, y, w, h float32, g *game) {
	vector.DrawFilledRect(dst, x, y, w, h, col(0, 0, 0, 135), false)
	for i := 1; i < graphN; i++ {
		a := g.history[(g.historyHead+i-1)%graphN]
		b := g.history[(g.historyHead+i)%graphN]
		x1 := x + float32(i-1)/graphN*w
		x2 := x + float32(i)/graphN*w
		y1 := y + h - float32(clamp(a/33.3, 0, 1))*h
		y2 := y + h - float32(clamp(b/33.3, 0, 1))*h
		vector.StrokeLine(dst, x1, y1, x2, y2, 2, col(60, 255, 140, 220), false)
	}
	ebitenutil.DebugPrintAt(dst, "FRAME TIME 16.7ms target / 33.3ms bad", int(x)+8, int(y)+8)
}

func drawStress(dst *ebiten.Image, t float64) {
	for i := 0; i < 650; i++ {
		a := t*float64(0.7+float64(i%9)*0.08) + float64(i)*1.618
		x := float32(640 + math.Cos(a)*float64(110+i%190))
		y := float32(360 + math.Sin(a*1.31)*float64(80+i%160))
		vector.DrawFilledCircle(dst, x, y, float32(2+i%5), col(uint8(60+i%150), uint8(120+i%100), 255, 95), false)
	}
}

func drawMetrics(dst *ebiten.Image, g *game) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	mode := "normal"
	if g.stress { mode = "stress" }
	lines := []string{
		"RADXA / EBITEN RENDER TEST",
		fmt.Sprintf("FPS %.1f  TPS %.1f  frame %.2fms  max %.2fms", ebiten.ActualFPS(), ebiten.ActualTPS(), g.frameMS, g.maxFrameMS),
		fmt.Sprintf("heap %.1fMB  goroutines %d  mode %s", float64(m.Alloc)/1024/1024, runtime.NumGoroutine(), mode),
	}
	if g.showHelp {
		lines = append(lines, "F fullscreen  S stress  H help  Q/Esc quit")
	}
	ebitenutil.DebugPrintAt(dst, strings.Join(lines, "\n"), 32, 28)
}

func col(r, g, b, a uint8) color.RGBA { return color.RGBA{R: r, G: g, B: b, A: a} }
func clamp(v, lo, hi float64) float64 { return math.Max(lo, math.Min(hi, v)) }
