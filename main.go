package main

import . "github.com/gen2brain/raylib-go/raylib"

func main() {
	InitWindow(800, 450, "raylib [core] example - basic window")
	defer CloseWindow()

	SetWindowPosition(0, 0)
	SetTargetFPS(60)

	for !WindowShouldClose() {
		BeginDrawing()

		ClearBackground(RayWhite)
		DrawGateAnd(Vector2{X: 100, Y: 200})

		EndDrawing()
	}
}

func DrawGateAnd(pos Vector2) {
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

	dots := [3]Vector2{
		{X: -stroke / 2, Y: h / 4},
		{X: -stroke / 2, Y: h * 3 / 4},
		{X: 2*w + stroke/2, Y: h / 2},
	}

	for _, dot := range dots {
		c := Vector2Add(pos, dot)
		DrawCircle(int32(c.X), int32(c.Y), 10, Black)
	}
}
