package main

import (
	"fmt"
	. "github.com/gen2brain/raylib-go/raylib"
)

type Gate int32
type Screen int32
type CanvasState int32

const (
	builder Screen = iota
	runner
)

const (
	And Gate = iota
	Or
	Not
	Nand
	Nor
	Xor
)

const (
	normal CanvasState = iota
	attached
)

type RuntimeNetwork struct {
	gates       []Gate
	inputs      [][2]bool
	output      []bool
	connections []*bool
}

type CanvasGate struct {
	gate        Gate
	inputs      [2]bool
	output      bool
	connections []*CanvasGate
	position    Vector2
}

type Canvas struct {
	camera   Camera2D
	gates    []CanvasGate
	attached *CanvasGate
	state    CanvasState
}

func newCanvasGate(g Gate) *CanvasGate {
	return &CanvasGate{
		gate: g,
	}
}

func NewCanvas() Canvas {
	return Canvas{
		camera:   Camera2D{Zoom: 1},
		gates:    []CanvasGate{},
		attached: nil,
	}
}

func (canvas *Canvas) attachGate(g Gate) {
	canvas.attached = newCanvasGate(g)
	canvas.state = attached
}

func (g Gate) String() string {
	switch g {
	case And:
		return "AND"
	case Or:
		return "OR"
	case Not:
		return "NOT"
	case Nand:
		return "NAND"
	case Nor:
		return "NOR"
	case Xor:
		return "XOR"
	default:
		panic(fmt.Sprintf("Missing String for Gate of type", g))
	}
}

func drawGate(g Gate, pos Vector2) {
	switch g {
	case And:
		drawGateAnd(pos)
	default:
		panic(fmt.Sprintf("Drawing not implemented for Gate of type %s", g))
	}
}

func drawGateAnd(pos Vector2) {
	var h, stroke float32 = 100, 10
	w := h / 2
	lines := [3][2]Vector2{
		{Vector2{X: 0, Y: 0}, Vector2{X: 0, Y: h}},
		{Vector2{X: -stroke / 2, Y: 0}, Vector2{X: w, Y: 0}},
		{Vector2{X: -stroke / 2, Y: h}, Vector2{X: w, Y: h}},
	}

	for _, line := range lines {
		x := Vector2Add(pos, line[0])
		y := Vector2Add(pos, line[1])
		DrawLineEx(x, y, stroke, Black)
	}

	DrawRing(Vector2Add(pos, Vector2{X: w, Y: h / 2}), w-stroke/2, w+stroke/2, -90, 90, 10, Black)

	dots := [2]Vector2{
		{X: -stroke / 2, Y: h / 4},
		{X: -stroke / 2, Y: h * 3 / 4},
	}

	for _, dot := range dots {
		c := Vector2Add(pos, dot)
		DrawCircle(int32(c.X), int32(c.Y), 10, Black)
	}
}

func (canvas *Canvas) drawGates() {
	for _, g := range canvas.gates {
		drawGate(g.gate, g.position)
	}
}

func (canvas *Canvas) builderScreen() {
	switch canvas.state {
	case normal:
		canvas.normalState()
	case attached:
		canvas.attachedState()
	}

	BeginMode2D(canvas.camera)
	canvas.drawAttached()
	canvas.drawGates()
	EndMode2D()

	button := gateButton(NewRectangle(0, 0, 100, 50), "AND")

	if button && canvas.attached == nil {
		canvas.attachGate(And)
	}
}

func gateButton(rect Rectangle, s string) bool {
	r := rect.ToInt32()
	mouse := GetMousePosition()
	inside := CheckCollisionPointRec(mouse, rect)
	down := IsMouseButtonDown(MouseButtonLeft)
	color := LightGray
	lineColor := Gray
	var stroke float32 = 5.0

	if inside {
		color = SkyBlue
		lineColor = Blue

		if down {
			color = Blue
			lineColor = DarkBlue
		}

	}

	DrawRectangle(r.X, r.Y, r.Width, r.Height, color)
	DrawRectangleLinesEx(rect, stroke, lineColor)
	DrawText(s, r.X+int32(stroke), r.Y+int32(stroke), 20, DarkGray)

	return inside && down
}

func (canvas *Canvas) normalState() {
	if IsMouseButtonDown(MouseButtonLeft) {
		delta := GetMouseDelta()
		delta = Vector2Scale(delta, -1/canvas.camera.Zoom)
		canvas.camera.Target = Vector2Add(canvas.camera.Target, delta)
	}
	wheel := GetMouseWheelMove()
	if wheel != 0 {
		mouseWorldPos := GetScreenToWorld2D(GetMousePosition(), canvas.camera)

		// Set the offset to where the mouse is
		canvas.camera.Offset = GetMousePosition()

		// Set the target to match, so that the camera maps the world space point
		// under the cursor to the screen space point under the cursor at any zoom
		canvas.camera.Target = mouseWorldPos

		// Zoom increment
		var zoomIncrement float32 = 0.125

		canvas.camera.Zoom += (wheel * zoomIncrement)
		if canvas.camera.Zoom < zoomIncrement {
			canvas.camera.Zoom = zoomIncrement
		}
	}
}

func (canvas *Canvas) attachedState() {
	if IsMouseButtonReleased(MouseButtonLeft) {
		canvas.placeAttached()
		canvas.state = normal
	}
}

func (canvas *Canvas) placeAttached() {
	canvas.attached.position = GetMousePosition()
	canvas.gates = append(canvas.gates, *canvas.attached)
	canvas.attached = nil
}

func (canvas *Canvas) drawAttached() {
	if canvas.attached == nil {
		return
	}
	mouse := GetMousePosition()
	drawGate(canvas.attached.gate, mouse)
}

func runnerScreen() {}

func main() {
	InitWindow(800, 450, "raylib [core] example - basic window")
	defer CloseWindow()

	SetTargetFPS(60)

	currentScreen := builder
	canvas := NewCanvas()

	for !WindowShouldClose() {
		BeginDrawing()
		ClearBackground(RayWhite)

		switch currentScreen {
		case builder:
			canvas.builderScreen()
		case runner:
			runnerScreen()
		}

		EndDrawing()
	}
}
