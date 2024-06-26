package main

import (
	"fmt"
	"image/color"
	"math"

	. "github.com/gen2brain/raylib-go/raylib"
)

type Screen int32

const (
	builder Screen = iota
	runner
)

type Logic int32

const (
	And Logic = iota
	Or
	Not
	Nand
	Nor
	Xor
	Switch
	Out
)

type gateTemplate struct {
	logic     Logic
	n_inputs  inputSocketIndex
	n_outputs outputSocketIndex
}

var logicTypes = [...]Logic{
	And,
	Or,
	Not,
	Nand,
	Nor,
	Xor,
	Switch,
	Out,
}

var gateTemplates = [...]gateTemplate{
	{And, 2, 1},
	{Or, 2, 1},
	{Nand, 2, 1},
	{Nor, 2, 1},
	{Xor, 2, 1},
	{Not, 1, 1},
	{Switch, 0, 1},
	{Out, 1, 0},
}

type socketIndex uint32
type inputSocketIndex socketIndex
type outputSocketIndex socketIndex

type CanvasState int32
type runnerState int32

const (
	idle CanvasState = iota
	dragging
	attached
	movingGate
	creatingLink
)

const (
	evaluateNext runnerState = iota
	propagateNext
)

const screenWidth = 800
const screenHeight = 450
const canvasButtonHeight = 50
const gateWidth float32 = 110
const gateHeight float32 = 70
const buttonWidth = 90
const buttonMargin = 10
const socketRadius float32 = 12
const linkStroke float32 = 10
const linkLead = 40
const currentSpacing = 20
const currentRadius = 3

type canvasInputSocket struct {
	gate  *canvasGate
	index inputSocketIndex
	link  *canvasOutputSocket
}

type canvasOutputSocket struct {
	gate  *canvasGate
	index outputSocketIndex
	links []*canvasInputSocket
}

type canvasGate struct {
	logic         Logic
	n_inputs      inputSocketIndex
	n_outputs     outputSocketIndex
	outputSockets []canvasOutputSocket
	inputSockets  []canvasInputSocket
	position      Vector2
	inputs        []bool
	outputs       []bool
}

type canvasLink struct {
	fromSocket *canvasOutputSocket
	toSocket   *canvasInputSocket
}

type Canvas struct {
	currentScreen       Screen
	state               CanvasState
	runnerState         runnerState
	guiRenderTexture    RenderTexture2D
	canvasCamera        Camera2D
	canvasRenderTexture RenderTexture2D
	gates               []*canvasGate
	contextGate         *canvasGate
	contextLink         *canvasLink
	frame               int
}

func newCanvasLink() *canvasLink {
	return &canvasLink{}
}

func newInputSocket(gate *canvasGate, index inputSocketIndex) canvasInputSocket {
	return canvasInputSocket{gate, index, nil}
}

func newOutputSocket(gate *canvasGate, index outputSocketIndex) canvasOutputSocket {
	return canvasOutputSocket{gate, index, make([]*canvasInputSocket, 0)}
}

func newCanvasGate(g Logic, n_inputs inputSocketIndex, n_outputs outputSocketIndex) *canvasGate {
	// n_inputs := inputSocketIndex(2)
	// n_outputs := outputSocketIndex(1)
	gate := &canvasGate{
		logic:         g,
		n_inputs:      n_inputs,
		n_outputs:     n_outputs,
		inputSockets:  make([]canvasInputSocket, n_inputs),
		outputSockets: make([]canvasOutputSocket, n_outputs),
		position:      Vector2{X: 0, Y: 0},
		inputs:        make([]bool, n_inputs),
		outputs:       make([]bool, n_outputs),
	}

	for i := inputSocketIndex(0); i < n_inputs; i++ {
		gate.inputSockets[i] = newInputSocket(gate, i)
	}
	for i := outputSocketIndex(0); i < n_outputs; i++ {
		gate.outputSockets[i] = newOutputSocket(gate, i)
	}

	return gate
}

func NewCanvas() Canvas {
	return Canvas{
		currentScreen:       builder,
		guiRenderTexture:    LoadRenderTexture(screenWidth, canvasButtonHeight),
		canvasCamera:        Camera2D{Zoom: 1},
		canvasRenderTexture: LoadRenderTexture(screenWidth, screenHeight),
		gates:               []*canvasGate{},
		contextGate:         nil,
	}
}

func (canvas *Canvas) attachGate(template gateTemplate) {
	canvas.contextGate = newCanvasGate(template.logic, template.n_inputs, template.n_outputs)
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
	case Switch:
		return "SWITCH"
	case Out:
		return "OUT"
	default:
		panic(fmt.Sprintf("Missing String for Gate of type %#v", g))
	}
}

func (gate *canvasGate) getInputSocketPlacement(i inputSocketIndex) Vector2 {
	unit := gateHeight / float32(gate.n_inputs)
	offset := Vector2{X: -gateWidth / 2, Y: -gateHeight/2 + unit/2 + float32(i)*unit}
	return Vector2Add(gate.position, offset)
}

func (gate *canvasGate) getOuputSocketPlacement(_ outputSocketIndex) Vector2 {
	return Vector2Add(gate.position, Vector2{X: gateWidth / 2, Y: 0})
}

func removeElement[T any](list []*T, i int) []*T {
	l := len(list)
	list[i] = list[l-1]
	list[l-1] = nil
	list = list[:l-1]
	return list
}

func (canvas *Canvas) drawGate(gate *canvasGate) {
	var fillColor color.RGBA
	var strokeColor color.RGBA
	var textColor color.RGBA
	if canvas.currentScreen == builder {
		fillColor = DarkGray
		strokeColor = Gray
		textColor = Black
	} else {
		textColor = Black
		fillColor = DarkGray
		strokeColor = Gray
	}
	drawNamedRectangle(NewRectangle(gate.position.X-gateWidth/2, gate.position.Y-gateHeight/2, gateWidth, gateHeight), gate.logic.String(), fillColor, strokeColor, textColor)
	canvas.drawSockets(gate)
}

func (canvas *Canvas) getSocketColor(values []bool, i socketIndex) color.RGBA {
	var color color.RGBA
	if canvas.currentScreen == runner && values[i] == true {
		color = Green
	}
	if canvas.currentScreen == runner && values[i] == false {
		color = Red
	}
	if canvas.currentScreen == builder {
		color = DarkGray
	}
	return color
}

func (canvas *Canvas) drawSockets(gate *canvasGate) {
	var segments int32 = 5
	for i := inputSocketIndex(0); i < gate.n_inputs; i++ {
		pos := gate.getInputSocketPlacement(i)
		color := canvas.getSocketColor(gate.inputs, socketIndex(i))
		DrawCircleSector(pos, socketRadius, 90, 270, segments, color)
	}
	for i := outputSocketIndex(0); i < gate.n_outputs; i++ {
		pos := gate.getOuputSocketPlacement(i)
		color := canvas.getSocketColor(gate.outputs, socketIndex(i))
		DrawCircleSector(pos, socketRadius, -90, 90, segments, color)
	}

}

func drawNamedRectangle(rect Rectangle, text string, stroke color.RGBA, fill color.RGBA, textColor color.RGBA) {
	var strokeSize float32 = 5.0
	DrawRectangleV(Vector2{X: rect.X, Y: rect.Y}, Vector2{X: rect.Width, Y: rect.Height}, fill)
	DrawRectangleLinesEx(rect, strokeSize, stroke)
	var textSize float32 = 20
	v := MeasureTextEx(GetFontDefault(), text, textSize, 0)
	offset_x := (rect.Width - v.X) / 2
	offset_y := (rect.Height - v.Y) / 2
	DrawText(text, int32(rect.X+offset_x), int32(rect.Y+offset_y), int32(textSize), textColor)
}

func (canvas *Canvas) drawGates() {
	for _, g := range canvas.gates {
		if canvas.currentScreen == runner && (g.logic == Switch || g.logic == Out) {
			continue
		}
		canvas.drawGate(g)
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
		i := canvas.contextLink.fromSocket.index
		from := canvas.contextLink.fromSocket.gate.getOuputSocketPlacement(i)
		to := GetMousePosition()
		to = GetScreenToWorld2D(to, canvas.canvasCamera)
		canvas.drawAngledLine(from, to, Red)
	}

	for _, gate := range canvas.gates {
		for _, fromSocket := range gate.outputSockets {
			from := gate.getOuputSocketPlacement(fromSocket.index)

			color := Purple

			if canvas.currentScreen == runner {
				output := gate.outputs[fromSocket.index]
				switch output {
				case true:
					color = Green
				case false:
					color = Red
				}
			}

			for _, toSocket := range fromSocket.links {
				to := toSocket.gate.getInputSocketPlacement(toSocket.index)

				canvas.drawAngledLine(from, to, color)
			}
		}
	}
}

func (canvas *Canvas) drawAngledLine(from Vector2, to Vector2, color color.RGBA) {
	factor := Vector2Distance(from, to) / (2 * linkLead)
	factor = Clamp(factor, 0, 1)
	from_offset := Vector2Add(from, Vector2{X: factor * linkLead, Y: 0})
	to_offset := Vector2Add(to, Vector2{X: -factor * linkLead, Y: 0})
	DrawLineEx(from, from_offset, linkStroke, Black)
	DrawLineEx(to, to_offset, linkStroke, Black)
	DrawLineEx(from_offset, to_offset, linkStroke, Black)

	if canvas.currentScreen == runner {
		head := Vector2Distance(from, from_offset)
		tail := Vector2Distance(to, to_offset)
		body := Vector2Distance(from_offset, to_offset)
		total_distance := head + body + tail

		n_points := int(total_distance / currentSpacing)
		for i := 0; i < n_points; i++ {
			t_offset := float32(i) / float32(n_points)
			t := t_offset + float32(canvas.frame)/1000
			t = float32(float64(t) - math.Floor(float64(t)))
			offset := total_distance * t

			var c Vector2
			if offset < head {
				c = Vector2Lerp(from, from_offset, offset/head)
			}
			if head <= offset && offset < head+body {
				c = Vector2Lerp(from_offset, to_offset, (offset-head)/body)
			}
			if head+body < offset {
				c = Vector2Lerp(to_offset, to, (offset-head-body)/tail)
			}
			DrawCircle(int32(c.X), int32(c.Y), currentRadius, color)
		}

	}

}

func (canvas *Canvas) builderScreen() {
	BeginTextureMode(canvas.guiRenderTexture)
	for i, template := range gateTemplates {
		button := gateButton(NewRectangle(float32(i*(buttonWidth+buttonMargin)), 0, buttonWidth, canvasButtonHeight), template.logic.String())

		if button && canvas.state == idle {
			canvas.attachGate(template)
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

	canvas.draw()
}

func (canvas *Canvas) draw() {
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

func Button(rect Rectangle, s string) bool {
	mouse := GetMousePosition()
	inside := CheckCollisionPointRec(mouse, rect)
	release := IsMouseButtonReleased(MouseButtonLeft)
	color := LightGray
	lineColor := Gray

	if inside {
		color = SkyBlue
		lineColor = Blue

		if release {
			color = Blue
			lineColor = DarkBlue
		}

	}
	drawNamedRectangle(rect, s, lineColor, color, DarkGray)
	return inside && release
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
	canvas.contextGate = gate
	canvas.state = movingGate
}

func (canvas *Canvas) checkGateDelete(i int) {
	if IsMouseButtonDown(MouseButtonRight) {
		canvas.gates = removeElement(canvas.gates, i)
	}
}

func isHoveringGate(point Vector2, gate *canvasGate) bool {
	rect := NewRectangle(gate.position.X-gateWidth/2, gate.position.Y-gateHeight/2, gateWidth, gateHeight)
	return CheckCollisionPointRec(point, rect)
}

func (canvas *Canvas) isHoveringInputSocket(point Vector2, gate *canvasGate) *inputSocketIndex {
	for i := inputSocketIndex(0); i < gate.n_inputs; i++ {
		center := gate.getInputSocketPlacement(i)
		hover := CheckCollisionPointCircle(point, center, socketRadius)
		if hover {
			return &i
		}
	}
	return nil

}

func (canvas *Canvas) isHoveringOutputSocket(point Vector2, gate *canvasGate) *outputSocketIndex {
	for i := outputSocketIndex(0); i < gate.n_outputs; i++ {
		center := gate.getOuputSocketPlacement(i)
		hover := CheckCollisionPointCircle(point, center, socketRadius)
		if hover {
			return &i
		}
	}
	return nil
}

func (canvas *Canvas) idleState() {
	mouse := GetMousePosition()
	mouse = GetScreenToWorld2D(mouse, canvas.canvasCamera)
	for i, gate := range canvas.gates {
		hoveringGate := isHoveringGate(mouse, gate)
		if hoveringGate {
			canvas.checkGateMove(gate)
			canvas.checkGateDelete(i)
		}

		outputSocketIndex := canvas.isHoveringOutputSocket(mouse, gate)
		if outputSocketIndex != nil && !hoveringGate {
			if IsMouseButtonDown(MouseButtonLeft) {
				canvas.state = creatingLink
				canvas.contextLink = &canvasLink{fromSocket: &gate.outputSockets[*outputSocketIndex]}
			}
		}

		inputSocketIndex := canvas.isHoveringInputSocket(mouse, gate)
		if inputSocketIndex != nil && !hoveringGate {
			if IsMouseButtonDown(MouseButtonLeft) {

				toSocket := &gate.inputSockets[*inputSocketIndex]

				// option 1: socket does not have link
				if toSocket.link == nil {
					return
				}

				// option 2: socket has link
				fromSocket := toSocket.link
				for i, link := range fromSocket.links {
					if link == toSocket {
						fromSocket.links = removeElement(fromSocket.links, i)
						toSocket.link = nil
						canvas.contextLink = &canvasLink{fromSocket: fromSocket}
						canvas.state = creatingLink
						break
					}
				}
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
	canvas.contextGate.position = Vector2Add(canvas.contextGate.position, delta)
	if IsMouseButtonReleased(MouseButtonLeft) {
		canvas.contextGate = nil
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
		inputSocketIndex := canvas.isHoveringInputSocket(mouse, gate)
		if inputSocketIndex != nil {
			toSocket := &gate.inputSockets[*inputSocketIndex]
			fromSocket := canvas.contextLink.fromSocket

			fromSocket.links = append(fromSocket.links, toSocket)
			toSocket.link = fromSocket
			canvas.contextLink = nil
			break
		}
	}

	// option 1: link not valid
	canvas.contextLink = nil
	canvas.state = idle

}

func (canvas *Canvas) placeAttached() {
	canvas.contextGate.position = GetScreenToWorld2D(GetMousePosition(), canvas.canvasCamera)
	canvas.gates = append(canvas.gates, canvas.contextGate)
	canvas.contextGate = nil
}

func (canvas *Canvas) drawAttached() {
	if canvas.contextGate == nil {
		return
	}
	mouse := GetScreenToWorld2D(GetMousePosition(), canvas.canvasCamera)
	canvas.contextGate.position = mouse
	canvas.drawGate(canvas.contextGate)
}

func (canvas *Canvas) step() {
	for _, gate := range canvas.gates {
		gate.propagate()
	}
	for _, gate := range canvas.gates {
		gate.evaluate()
	}
}

func (gate *canvasGate) evaluate() {
	switch gate.logic {
	case And:
		gate.outputs[0] = gate.inputs[0] && gate.inputs[1]
	case Or:
		gate.outputs[0] = gate.inputs[0] || gate.inputs[1]
	case Not:
		gate.outputs[0] = !gate.inputs[0]
	case Nand:
		gate.outputs[0] = !(gate.inputs[0] && gate.inputs[1])
	case Nor:
		gate.outputs[0] = !(gate.inputs[0] || gate.inputs[1])
	case Xor:
		gate.outputs[0] = gate.inputs[0] != gate.inputs[1]
	case Switch:
	case Out:
	default:
		panic(fmt.Sprintf("Gate of type %v, does not implement evaluate!", gate.logic))
	}
}

func (gate *canvasGate) propagate() {
	for _, fromSocket := range gate.outputSockets {
		for _, toSocket := range fromSocket.links {
			i := toSocket.index
			j := fromSocket.index
			toSocket.gate.inputs[i] = fromSocket.gate.outputs[j]
		}
	}
}

func (canvas *Canvas) runnerScreen() {

	BeginTextureMode(canvas.guiRenderTexture)
	button := Button(NewRectangle(0, 0, buttonWidth, canvasButtonHeight), "Step")
	EndTextureMode()

	if button {
		canvas.step()
	}

	switch canvas.state {
	case idle:
		canvas.checkCanvasDrag()
		canvas.canvasZoom()
	case dragging:
		canvas.canvasDrag()
		if IsMouseButtonReleased(MouseButtonLeft) {
			canvas.state = idle
		}
	}

	canvas.checkSwitches()
	canvas.draw()
	canvas.drawSpecialGates()
}

func (canvas *Canvas) checkSwitches() {
	mouse := GetMousePosition()
	mouse = GetScreenToWorld2D(mouse, canvas.canvasCamera)
	for _, gate := range canvas.gates {
		if gate.logic == Switch {
			rect := NewRectangle(gate.position.X-gateWidth/2, gate.position.Y-gateHeight/2, gateWidth, gateHeight)
			hover := CheckCollisionPointRec(mouse, rect)
			click := IsMouseButtonPressed(MouseButtonLeft)
			if hover && click {
				gate.outputs[0] = !gate.outputs[0]
			}
		}
	}

}

func (canvas *Canvas) drawSpecialGates() {
	BeginTextureMode(canvas.canvasRenderTexture)
	BeginMode2D(canvas.canvasCamera)
	for _, gate := range canvas.gates {
		if gate.logic == Switch || gate.logic == Out {
			rect := NewRectangle(gate.position.X-gateWidth/2, gate.position.Y-gateHeight/2, gateWidth, gateHeight)
			canvas.drawSockets(gate)

			var strokeColor color.RGBA
			if gate.logic == Switch {
				strokeColor = canvas.getSocketColor(gate.outputs, 0)
			}
			if gate.logic == Out {
				strokeColor = canvas.getSocketColor(gate.inputs, 0)
			}
			fillColor := Gray
			drawNamedRectangle(rect, gate.logic.String(), strokeColor, fillColor, Black)
		}
	}
	EndMode2D()
	EndTextureMode()
}

func (canvas *Canvas) clearGui() {
	BeginTextureMode(canvas.guiRenderTexture)
	ClearBackground(color.RGBA{0, 0, 0, 0})
	EndTextureMode()
}

func main() {
	InitWindow(screenWidth, screenHeight, "Logic gates")
	defer CloseWindow()

	SetTargetFPS(60)

	canvas := NewCanvas()

	for !WindowShouldClose() {

		if IsKeyPressed(KeyEnter) {
			canvas.clearGui()
			switch canvas.currentScreen {
			case builder:
				canvas.currentScreen = runner
			case runner:
				canvas.currentScreen = builder
			}
		}

		switch canvas.currentScreen {
		case builder:
			canvas.builderScreen()
		case runner:
			canvas.runnerScreen()
		}

		BeginDrawing()
		ClearBackground(Black)
		DrawTextureRec(canvas.canvasRenderTexture.Texture, NewRectangle(0, 0, screenWidth, -screenHeight), Vector2{X: 0, Y: 0}, White)
		DrawTextureRec(canvas.guiRenderTexture.Texture, NewRectangle(0, 0, screenWidth, -canvasButtonHeight), Vector2{X: 0, Y: 0}, White)
		EndDrawing()

		canvas.frame += 1
	}
}
