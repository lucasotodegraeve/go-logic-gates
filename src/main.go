package main

import (
	"fmt"
	"image/color"

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

const screenWidth = 800
const screenHeight = 450
const canvasButtonHeight = 50
const gateWidth = 100
const gateHeight = 70

type RuntimeNetwork struct {
	gates       []Gate
	inputs      [][2]bool
	output      []bool
	connections []*bool
}

type CanvasGate struct {
	logic       Gate
	inputs      [2]bool
	output      bool
	connections []*CanvasGate
	position    Vector2
}

type Canvas struct {
	guiRenderTexture    RenderTexture2D
	canvasCamera        Camera2D
	canvasRenderTexture RenderTexture2D
	gates               []CanvasGate
	attached            *CanvasGate
	state               CanvasState
}

func newCanvasGate(g Gate) *CanvasGate {
	return &CanvasGate{
		logic: g,
	}
}

func NewCanvas() Canvas {
	return Canvas{
		guiRenderTexture:    LoadRenderTexture(screenWidth, canvasButtonHeight),
		canvasCamera:        Camera2D{Zoom: 1},
		canvasRenderTexture: LoadRenderTexture(screenWidth, screenHeight),
		gates:               []CanvasGate{},
		attached:            nil,
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

func drawGate(gate *CanvasGate) {
	drawNamedRectangle(NewRectangle(gate.position.X, gate.position.Y, gateWidth, gateHeight), gate.logic.String(), DarkGray, Gray, Black)
	var size float32 = 10
	var segments int32 = 5
	DrawCircleSector(Vector2{X: gate.position.X, Y: gate.position.Y + gateHeight*1/4}, size, 90, 270, segments, DarkGray)
	DrawCircleSector(Vector2{X: gate.position.X, Y: gate.position.Y + gateHeight*3/4}, size, 90, 270, segments, DarkGray)
	DrawCircleSector(Vector2{X: gate.position.X + gateWidth, Y: gate.position.Y + gateHeight*1/2}, size, -90, 90, segments, DarkGray)
}

func drawNamedRectangle(rect Rectangle, text string, stroke color.RGBA, fill color.RGBA, textColor color.RGBA) {
	var strokeSize float32 = 5.0
	DrawRectangle(int32(rect.X), int32(rect.Y), int32(rect.Width), int32(rect.Height), fill)
	DrawRectangleLinesEx(rect, strokeSize, stroke)
	var textSize float32 = 30
	v := MeasureTextEx(GetFontDefault(), text, textSize, 0)
	offset_x := (rect.Width - v.X) / 2
	offset_y := (rect.Height - v.Y) / 2
	DrawText(text, int32(rect.X+offset_x), int32(rect.Y+offset_y), int32(textSize), textColor)
}

func (canvas *Canvas) drawGates() {
	for _, g := range canvas.gates {
		drawGate(&g)
	}
}

func (canvas *Canvas) drawGrid() {
	var size int32 = 50
	w, h := GetScreenWidth(), GetScreenHeight()
	v1 := GetScreenToWorld2D(Vector2{0, 0}, canvas.canvasCamera)
	v2 := GetScreenToWorld2D(Vector2{float32(w), float32(h)}, canvas.canvasCamera)

	v_int := int32(v1.X)
	for i := v_int - v_int%size; i < int32(v2.X); i += size {
		DrawLine(i, int32(v1.Y), i, int32(v2.Y), Gray)
	}
	v_int = int32(v1.Y)
	for i := v_int - v_int%size; i < int32(v2.Y); i += size {
		DrawLine(int32(v1.X), i, int32(v2.X), i, Gray)
	}
}

func (canvas *Canvas) builderScreen() {
	BeginTextureMode(canvas.guiRenderTexture)
	button := gateButton(NewRectangle(0, 0, 100, canvasButtonHeight), "AND")
	EndTextureMode()

	if button && canvas.attached == nil {
		canvas.attachGate(And)
	}

	switch canvas.state {
	case normal:
		canvas.normalState()
	case attached:
		canvas.attachedState()
	}

	BeginTextureMode(canvas.canvasRenderTexture)
	ClearBackground(RayWhite)
	BeginMode2D(canvas.canvasCamera)
	canvas.drawGrid()
	canvas.drawAttached()
	canvas.drawGates()
	EndMode2D()
	EndTextureMode()
}

func gateButton(rect Rectangle, s string) bool {
	mouse := GetMousePosition()
	inside := CheckCollisionPointRec(mouse, rect)
	down := IsMouseButtonDown(MouseButtonLeft)
	color := LightGray
	lineColor := Gray

	if inside {
		color = SkyBlue
		lineColor = Blue

		if down {
			color = Blue
			lineColor = DarkBlue
		}

	}

	drawNamedRectangle(rect, s, lineColor, color, DarkGray)

	return inside && down
}

func (canvas *Canvas) normalState() {
	if IsMouseButtonDown(MouseButtonLeft) {
		delta := GetMouseDelta()
		delta = Vector2Scale(delta, -1/canvas.canvasCamera.Zoom)
		canvas.canvasCamera.Target = Vector2Add(canvas.canvasCamera.Target, delta)
	}
	wheel := GetMouseWheelMove()
	if wheel != 0 {
		mouseWorldPos := GetScreenToWorld2D(GetMousePosition(), canvas.canvasCamera)

		// Set the offset to where the mouse is
		canvas.canvasCamera.Offset = GetMousePosition()

		// Set the target to match, so that the camera maps the world space point
		// under the cursor to the screen space point under the cursor at any zoom
		canvas.canvasCamera.Target = mouseWorldPos

		// Zoom increment
		var zoomIncrement float32 = 0.125

		canvas.canvasCamera.Zoom += (wheel * zoomIncrement)
		if canvas.canvasCamera.Zoom < zoomIncrement {
			canvas.canvasCamera.Zoom = zoomIncrement
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
	canvas.attached.position = GetScreenToWorld2D(GetMousePosition(), canvas.canvasCamera)
	canvas.gates = append(canvas.gates, *canvas.attached)
	canvas.attached = nil
}

func (canvas *Canvas) drawAttached() {
	if canvas.attached == nil {
		return
	}
	mouse := GetScreenToWorld2D(GetMousePosition(), canvas.canvasCamera)
	canvas.attached.position = mouse
	drawGate(canvas.attached)
}

func runnerScreen() {}

func main() {
	InitWindow(screenWidth, screenHeight, "Logic gates")
	defer CloseWindow()

	SetTargetFPS(60)

	currentScreen := builder
	canvas := NewCanvas()

	for !WindowShouldClose() {
		switch currentScreen {
		case builder:
			canvas.builderScreen()
		case runner:
			runnerScreen()
		}

		BeginDrawing()
		ClearBackground(Black)
		DrawTextureRec(canvas.canvasRenderTexture.Texture, NewRectangle(0, 0, screenWidth, -screenHeight), Vector2{X: 0, Y: 0}, White)
		DrawTextureRec(canvas.guiRenderTexture.Texture, NewRectangle(0, 0, screenWidth, -canvasButtonHeight), Vector2{X: 0, Y: 0}, White)
		EndDrawing()
	}
}
