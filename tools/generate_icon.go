package main

import (
	"image"
	"image/color"
	"image/png"
	"math"
	"os"

	"golang.org/x/image/vector"
)

type point struct{ x, y float32 }

func fillPolygon(dst *image.NRGBA, points []point, col color.NRGBA) {
	r := vector.NewRasterizer(dst.Bounds().Dx(), dst.Bounds().Dy())
	r.MoveTo(points[0].x, points[0].y)
	for _, p := range points[1:] {
		r.LineTo(p.x, p.y)
	}
	r.ClosePath()
	r.Draw(dst, dst.Bounds(), image.NewUniform(col), image.Point{})
}

func roundedRect(dst *image.NRGBA, col color.NRGBA) {
	const w, h, radius float32 = 1024, 1024, 220
	r := vector.NewRasterizer(int(w), int(h))
	r.MoveTo(radius, 0)
	r.LineTo(w-radius, 0)
	r.QuadTo(w, 0, w, radius)
	r.LineTo(w, h-radius)
	r.QuadTo(w, h, w-radius, h)
	r.LineTo(radius, h)
	r.QuadTo(0, h, 0, h-radius)
	r.LineTo(0, radius)
	r.QuadTo(0, 0, radius, 0)
	r.ClosePath()
	r.Draw(dst, dst.Bounds(), image.NewUniform(col), image.Point{})
}

func thickLine(dst *image.NRGBA, a, b point, width float32, col color.NRGBA) {
	dx, dy := b.x-a.x, b.y-a.y
	length := float32(math.Sqrt(float64(dx*dx + dy*dy)))
	if length == 0 {
		return
	}
	nx, ny := -dy/length*width/2, dx/length*width/2
	fillPolygon(dst, []point{
		{a.x + nx, a.y + ny}, {b.x + nx, b.y + ny},
		{b.x - nx, b.y - ny}, {a.x - nx, a.y - ny},
	}, col)
}

func main() {
	img := image.NewNRGBA(image.Rect(0, 0, 1024, 1024))
	roundedRect(img, color.NRGBA{R: 0x14, G: 0x99, B: 0x98, A: 255})
	fillPolygon(img, []point{
		{160, 500}, {870, 155}, {625, 875}, {500, 625},
	}, color.NRGBA{R: 255, G: 255, B: 255, A: 255})
	thickLine(img, point{870, 155}, point{500, 625}, 22, color.NRGBA{R: 0x14, G: 0x99, B: 0x98, A: 255})

	for _, path := range []string{"resources/icon.png", "cmd/nanoapi/app-icon.png"} {
		file, err := os.Create(path)
		if err != nil {
			panic(err)
		}
		err = png.Encode(file, img)
		closeErr := file.Close()
		if err != nil {
			panic(err)
		}
		if closeErr != nil {
			panic(closeErr)
		}
	}
}
