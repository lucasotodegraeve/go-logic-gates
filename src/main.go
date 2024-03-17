package main

import (
	"fmt"
	"image/color"
	"log"

	. "github.com/gen2brain/raylib-go/raylib"
)

type Screen int32

const (
	builder Screen = iota
	runner
)

const (
	And Logic = iota
	Or
	Not
	Nand
	Nor
	Xor
)

type Logic int32

var logicTypes = [...]Logic{
	And,
	Or,
	Not,
	Nand,
	Nor,
	Xor,
}

type socketID uint32
type inputSocket socketID
type outputSocket socketID

type CanvasState int32

const (
	idle CanvasState = iota
	dragging
	attached
	movingGate
	creatingLink
)

const screenWidth = 800
const screenHeight = 450
const canvasButtonHeight = 50
const gateWidth float32 = 110
const gateHeight float32 = 70
const buttonWidth = 110
const buttonMargin = 10
const socketRadius float32 = 12
const linkStroke float32 = 10

type canvasGate struct {
	logic     Logic
	n_inputs  inputSocket
	n_outputs outputSocket
	inputs    []bool
	outputs   []bool
	links     []*canvasLink
	position  Vector2
}

type canvasLink struct {
	fromGate   *canvasGate
	toGate     *canvasGate
	fromSocket outputSocket
	toSocket   inputSocket
}

type Canvas struct {
	state               CanvasState
	guiRenderTexture    RenderTexture2D
	canvasCamera        Camera2D
	canvasRenderTexture RenderTexture2D
	gates               []*canvasGate
	attached            *canvasGate
	selected            *canvasGate
	contextGate         *canvasGate
	contextLink         *canvasLink
}

func newCanvasLink() *canvasLink {
	return &canvasLink{}
}

func newCanvasGate(g Logic) *canvasGate {
	return &canvasGate{
		logic:     g,
		n_inputs:  2,
		n_outputs: 1,
		inputs:    nil,
		outputs:   nil,
		links:     make([]*canvasLink, 0),
		position:  Vector2{X: 0, Y: 0},
	}
}

func NewCanvas() Canvas {
	return Canvas{
		guiRenderTexture:    LoadRenderTexture(screenWidth, canvasButtonHeight),
		canvasCamera:        Camera2D{Zoom: 1},
		canvasRenderTexture: LoadRenderTexture(screenWidth, screenHeight),
		gates:               []*canvasGate{},
		attached:            nil,
		selected:            nil,
		contextGate:         nil,
	}
}

func (canvas *Canvas) attachGate(g Logic) {
	canvas.attached = newCanvasGate(g)
	canvas.state = attached
}

func (g Logic) String() string {
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
		panic(fmt.Sprintf("Missing String for Gate of type %#v", g))
	}
}

func (gate *canvasGate) getInputSocketPlacement(i inputSocket) Vector2 {
	var offset Vector2
	switch i {
	case 0:
		offset = Vector2{X: -gateWidth / 2, Y: -gateHeight/2 + gateHeight*1/4}
	case 1:
		offset = Vector2{X: -gateWidth / 2, Y: -gateHeight/2 + gateHeight*3/4}
	default:
		panic(fmt.Sprintf("Inputsocket with value %d out of range", i))
	}
	return Vector2Add(gate.position, offset)
}

func (gate *canvasGate) getOuputSocketPlacement(_ outputSocket) Vector2 {
	return Vector2Add(gate.position, Vector2{X: gateWidth / 2, Y: 0})
}

func drawGate(gate *canvasGate) {
	drawNamedRectangle(NewRectangle(gate.position.X-gateWidth/2, gate.position.Y-gateHeight/2, gateWidth, gateHeight), gate.logic.String(), DarkGray, Gray, Black)
	var segments int32 = 5

	for i := inputSocket(0); i < gate.n_inputs; i++ {
		pos := gate.getInputSocketPlacement(i)
		DrawCircleSector(pos, socketRadius, 90, 270, segments, DarkGray)
	}
	for i := outputSocket(0); i < gate.n_outputs; i++ {
		pos := gate.getOuputSocketPlacement(i)
		DrawCircleSector(pos, socketRadius, -90, 90, segments, DarkGray)
	}
}

func (canvas *Canvas) drawSelected() {
	if canvas.selected != nil {
		DrawRectangleLines(int32(canvas.selected.position.X-gateWidth/2), int32(canvas.selected.position.Y-gateHeight/2), int32(gateWidth), int32(gateHeight), Red)
	}
}

func drawNamedRectangle(rect Rectangle, text string, stroke color.RGBA, fill color.RGBA, textColor color.RGBA) {
	var strokeSize float32 = 5.0
	DrawRectangleV(Vector2{X: rect.X, Y: rect.Y}, Vector2{X: rect.Width, Y: rect.Height}, fill)
	DrawRectangleLinesEx(rect, strokeSize, stroke)
	var textSize float32 = 30
	v := MeasureTextEx(GetFontDefault(), text, textSize, 0)
	offset_x := (rect.Width - v.X) / 2
	offset_y := (rect.Height - v.Y) / 2
	DrawText(text, int32(rect.X+offset_x), int32(rect.Y+offset_y), int32(textSize), textColor)
}

func (canvas *Canvas) drawGates() {
	for _, g := range canvas.gates {
		drawGate(g)
	}
}

func (canvas *Canvas) drawGrid() {
	var size int32 = 50
	w, h := GetScreenWidth(), GetScreenHeight()
	v1 := GetScreenToWorld2D(Vector2{X: 0, Y: 0}, canvas.canvasCamera)
	v2 := GetScreenToWorld2D(Vector2{X: float32(w), Y: float32(h)}, canvas.canvasCamera)

	v_int := int32(v1.X)
	for i := v_int - v_int%size; i < int32(v2.X); i += size {
		DrawLine(i, int32(v1.Y), i, int32(v2.Y), Gray)
	}
	v_int = int32(v1.Y)
	for i := v_int - v_int%size; i < int32(v2.Y); i += size {
		DrawLine(int32(v1.X), i, int32(v2.X), i, Gray)
	}
}

func (canvas *Canvas) drawLinks() {
	if canvas.state == creatingLink {
		i := canvas.contextLink.fromSocket
		a := canvas.contextLink.fromGate.getOuputSocketPlacement(i)
		mouse := GetMousePosition()
		mouse = GetScreenToWorld2D(mouse, canvas.canvasCamera)
		DrawLineEx(a, mouse, linkStroke, Black)
	}

	for _, gate := range canvas.gates {
		for _, link := range gate.links {
			a := gate.getInputSocketPlacement(link.toSocket)
			b := link.fromGate.getOuputSocketPlacement(link.fromSocket)
			DrawLineEx(a, b, linkStroke, Black)
		}
	}
}

func (canvas *Canvas) builderScreen() {
	BeginTextureMode(canvas.guiRenderTexture)
	for i, gate := range logicTypes {
		button := gateButton(NewRectangle(float32(i*(buttonWidth+buttonMargin)), 0, buttonWidth, canvasButtonHeight), gate.String())

		if button && canvas.state == idle {
			canvas.attachGate(gate)
		}
	}
	EndTextureMode()

	switch canvas.state {
	case idle:
		canvas.idleState()
	case dragging:
		canvas.canvasDrag()
		if IsMouseButtonReleased(MouseButtonLeft) {
			canvas.state = idle
		}
	case attached:
		canvas.attachedState()
	case movingGate:
		canvas.moveGateState()
	case creatingLink:
		canvas.createLinkState()
	}

	BeginTextureMode(canvas.canvasRenderTexture)
	ClearBackground(RayWhite)
	BeginMode2D(canvas.canvasCamera)
	canvas.drawGrid()
	canvas.drawAttached()
	canvas.drawLinks()
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

func (canvas *Canvas) checkCanvasDrag() {
	if IsMouseButtonDown(MouseButtonLeft) && canvas.state == idle {
		canvas.state = dragging
	}
}

func (canvas *Canvas) canvasDrag() {
	delta := GetMouseDelta()
	delta = Vector2Scale(delta, -1/canvas.canvasCamera.Zoom)
	canvas.canvasCamera.Target = Vector2Add(canvas.canvasCamera.Target, delta)
}

func (canvas *Canvas) canvasZoom() {
	wheel := GetMouseWheelMove()
	if wheel == 0 {
		return
	}
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

func (canvas *Canvas) checkGateMove(gate *canvasGate) {
	if !IsMouseButtonDown(MouseButtonLeft) {
		return
	}
	canvas.selected = gate
	canvas.state = movingGate
}

func (canvas *Canvas) checkGateDelete(i int) {
	if IsMouseButtonDown(MouseButtonRight) {
		canvas.gates[i] = canvas.gates[len(canvas.gates)-1]
		canvas.gates[len(canvas.gates)-1] = nil
		canvas.gates = canvas.gates[:len(canvas.gates)-1]
	}
}

func isHoveringGate(point Vector2, gate *canvasGate) bool {
	rect := NewRectangle(gate.position.X-gateWidth/2, gate.position.Y-gateHeight/2, gateWidth, gateHeight)
	return CheckCollisionPointRec(point, rect)
}

// func checkHoverSocket(point Vector2, gate *canvasGate) bool {
// 	centers := [...]Vector2{
// 		{X: -gateWidth / 2, Y: -gateHeight/2 + gateHeight*1/4},
// 		{X: -gateWidth / 2, Y: -gateHeight/2 + gateHeight*3/4},
// 		{X: +gateWidth / 2, Y: 0},
// 	}

// 	for _, center := range centers {
// 		hover := CheckCollisionPointCircle(point, Vector2Add(gate.position, center), socketRadius)
// 		if hover {
// 			return true
// 		}
// 	}

// 	return false
// }

func (canvas *Canvas) isHoveringInputSocket(point Vector2, gate *canvasGate) *inputSocket {
	for i := inputSocket(0); i < gate.n_inputs; i++ {
		center := gate.getInputSocketPlacement(i)
		hover := CheckCollisionPointCircle(point, center, socketRadius)
		if hover {
			return &i
		}
	}
	return nil

}

func (canvas *Canvas) isHoveringOutputSocket(point Vector2, gate *canvasGate) *outputSocket {
	for i := outputSocket(0); i < gate.n_outputs; i++ {
		center := gate.getOuputSocketPlacement(i)
		hover := CheckCollisionPointCircle(point, center, socketRadius)
		if hover {
			return &i
		}
	}
	return nil
}

// func (canvas *Canvas) checkLinkDrag(gate *canvasGate, socket socketID) {
// 	if IsMouseButtonDown(MouseButtonLeft) {
// 		canvas.state = creatingLink
// 		canvas.contextLink = newCanvasLink(gate, socket)
// 	}
// }

func (canvas *Canvas) idleState() {
	mouse := GetMousePosition()
	mouse = GetScreenToWorld2D(mouse, canvas.canvasCamera)
	for i, gate := range canvas.gates {
		hoveringGate := isHoveringGate(mouse, gate)
		if hoveringGate {
			canvas.checkGateMove(gate)
			canvas.checkGateDelete(i)
		}

		p := canvas.isHoveringOutputSocket(mouse, gate)
		if p != nil && !hoveringGate {
			// canvas.checkLinkDrag()

			if IsMouseButtonDown(MouseButtonLeft) {
				canvas.state = creatingLink
				canvas.contextLink = &canvasLink{fromGate: gate, fromSocket: *p}
				log.Println("New link")
			}
		}
	}

	canvas.checkCanvasDrag()
	canvas.canvasZoom()
}

func (canvas *Canvas) attachedState() {
	if IsMouseButtonReleased(MouseButtonLeft) {
		canvas.placeAttached()
		canvas.state = idle
	}
}

func (canvas *Canvas) moveGateState() {
	delta := GetMouseDelta()
	delta = Vector2Scale(delta, 1/canvas.canvasCamera.Zoom)
	canvas.selected.position = Vector2Add(canvas.selected.position, delta)
	if IsMouseButtonReleased(MouseButtonLeft) {
		canvas.selected = nil
		canvas.state = idle
	}
}

func (canvas *Canvas) createLinkState() {
	if !IsMouseButtonReleased(MouseButtonLeft) {
		return
	}
	//option 2: link conntected to other node
	mouse := GetMousePosition()
	mouse = GetScreenToWorld2D(mouse, canvas.canvasCamera)
	for _, gate := range canvas.gates {
		i := canvas.isHoveringInputSocket(mouse, gate)
		if i != nil {
			canvas.contextLink.toGate = gate
			canvas.contextLink.toSocket = *i
			gate.links = append(gate.links, canvas.contextLink)
		}
	}

	// option 1: link not valid
	canvas.contextLink = nil
	canvas.state = idle

}

func (canvas *Canvas) placeAttached() {
	canvas.attached.position = GetScreenToWorld2D(GetMousePosition(), canvas.canvasCamera)
	canvas.gates = append(canvas.gates, canvas.attached)
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
