// Copyright 2010 The Freetype-Go Authors. All rights reserved.
// Use of this source code is governed by your choice of either the
// FreeType License or the GNU General Public License version 2,
// both of which can be found in the LICENSE file.

package main

import (
	"bufio"
	"exp/draw"
	"fmt"
	"image"
	"image/png"
	"log"
	"os"

	"freetype-go.googlecode.com/hg/freetype/raster"
)

type node struct {
	x, y, degree int
}

// These contours "outside" and "inside" are from the `A' glyph from the Droid
// Serif Regular font.

var outside = []node{
	node{414, 489, 1},
	node{336, 274, 2},
	node{327, 250, 0},
	node{322, 226, 2},
	node{317, 203, 0},
	node{317, 186, 2},
	node{317, 134, 0},
	node{350, 110, 2},
	node{384, 86, 0},
	node{453, 86, 1},
	node{500, 86, 1},
	node{500, 0, 1},
	node{0, 0, 1},
	node{0, 86, 1},
	node{39, 86, 2},
	node{69, 86, 0},
	node{90, 92, 2},
	node{111, 99, 0},
	node{128, 117, 2},
	node{145, 135, 0},
	node{160, 166, 2},
	node{176, 197, 0},
	node{195, 246, 1},
	node{649, 1462, 1},
	node{809, 1462, 1},
	node{1272, 195, 2},
	node{1284, 163, 0},
	node{1296, 142, 2},
	node{1309, 121, 0},
	node{1326, 108, 2},
	node{1343, 96, 0},
	node{1365, 91, 2},
	node{1387, 86, 0},
	node{1417, 86, 1},
	node{1444, 86, 1},
	node{1444, 0, 1},
	node{881, 0, 1},
	node{881, 86, 1},
	node{928, 86, 2},
	node{1051, 86, 0},
	node{1051, 184, 2},
	node{1051, 201, 0},
	node{1046, 219, 2},
	node{1042, 237, 0},
	node{1034, 260, 1},
	node{952, 489, 1},
	node{414, 489, -1},
}

var inside = []node{
	node{686, 1274, 1},
	node{453, 592, 1},
	node{915, 592, 1},
	node{686, 1274, -1},
}

func p(n node) raster.Point {
	x, y := 20+n.x/4, 380-n.y/4
	return raster.Point{raster.Fixed(x * 256), raster.Fixed(y * 256)}
}

func contour(r *raster.Rasterizer, ns []node) {
	if len(ns) == 0 {
		return
	}
	i := 0
	r.Start(p(ns[i]))
	for {
		switch ns[i].degree {
		case -1:
			// -1 signifies end-of-contour.
			return
		case 1:
			i += 1
			r.Move1(p(ns[i]))
		case 2:
			i += 2
			r.Move2(p(ns[i-1]), p(ns[i]))
		default:
			panic("bad degree")
		}
	}
}

func showNodes(m *image.RGBA, ns []node) {
	for _, n := range ns {
		p := p(n)
		x, y := int(p.X)/256, int(p.Y)/256
		if x < 0 || x >= m.Width() || y < 0 || y >= m.Height() {
			continue
		}
		var c image.Color
		switch n.degree {
		case 0:
			c = image.Aqua
		case 1:
			c = image.Red
		case 2:
			c = image.Red
		}
		if c != nil {
			m.Set(x, y, c)
		}
	}
}

func main() {
	// Rasterize the contours to a mask image.
	const (
		w = 400
		h = 400
	)
	r := raster.New(w, h)
	contour(r, outside)
	contour(r, inside)
	mask := image.NewAlpha(w, h)
	r.Rasterize(raster.AlphaSrcPainter(mask))

	// Draw the mask image (in gray) onto an RGBA image.
	rgba := image.NewRGBA(w, h)
	gray := image.ColorImage{image.AlphaColor{0x1f}}
	draw.Draw(rgba, draw.Rect(0, 0, w, h), image.Black, draw.ZP)
	draw.DrawMask(rgba, draw.Rect(0, 0, w, h), gray, draw.ZP, mask, draw.ZP, draw.Over)
	showNodes(rgba, outside)
	showNodes(rgba, inside)

	// Save that RGBA image to disk.
	f, err := os.Open("out.png", os.O_CREAT|os.O_WRONLY, 0600)
	if err != nil {
		log.Stderr(err)
		return
	}
	defer f.Close()
	b := bufio.NewWriter(f)
	err = png.Encode(b, rgba)
	if err != nil {
		log.Stderr(err)
		return
	}
	err = b.Flush()
	if err != nil {
		log.Stderr(err)
		return
	}
	fmt.Println("Wrote out.png OK.")
}
