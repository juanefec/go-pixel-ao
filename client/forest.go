package main

import (
	"math/rand"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
)

type TreeType int

const (
	ZombieTree TreeType = iota
	TallTree
	NoLeafsTallTree
)

type Tree struct {
	Frame pixel.Rect
	Batch *pixel.Batch
	Pic   *pixel.Picture
}

type Trees []*Tree

func (s Trees) Load(kind TreeType, imagPath string) {
	sheet := Pictures[imagPath]
	tree := &Tree{
		Pic:   &sheet,
		Batch: pixel.NewBatch(&pixel.TrianglesData{}, sheet),
		Frame: sheet.Bounds(),
	}
	s[kind] = tree
}

func (s Trees) Draw(win *pixelgl.Window) {
	for i := range s {
		s[i].Batch.Draw(win)
	}
}

type Forest struct {
	Pic, GrassPic, FencePicH, FencePicV                            pixel.Picture
	Frames                                                         []pixel.Rect
	FenceFrameH, FenceFrameV, GrassFrames                          pixel.Rect
	Batch, GrassBatch, FenceBatchHTOP, FenceBatchHBOT, FenceBatchV *pixel.Batch
	Trees                                                          Trees
}

func generateRandomPositions(c, bot, top, left, rigth int64) []pixel.Vec {
	position := make([]pixel.Vec, c)
	rand.Seed(c + top + bot + left + rigth)
	for i := range position {
		position[i] = pixel.V(random(left, rigth), random(bot, top))
	}
	return position
}
func random(min, max int64) float64 {
	return float64(rand.Int63n(max-min) + min)
}

func NewForest(cs *CollisionSystem) *Forest {
	treeSheet := Pictures["./images/trees.png"]
	treeBatch := pixel.NewBatch(&pixel.TrianglesData{}, treeSheet)
	treeFrames := getFrames(treeSheet, 32, 32, 3, 3)

	trees := make(Trees, 3)
	trees.Load(ZombieTree, "./images/arbolmuerto.png")
	trees.Load(TallTree, "./images/talltree.png")
	trees.Load(NoLeafsTallTree, "./images/tallnoleafstree.png")

	grassSheet := Pictures["./images/newGrass.png"]
	grassBatch := pixel.NewBatch(&pixel.TrianglesData{}, grassSheet)
	grassFrames := grassSheet.Bounds()

	vfenceSheet := Pictures["./images/verticalfence.png"]
	vfenceBatch := pixel.NewBatch(&pixel.TrianglesData{}, vfenceSheet)
	vfenceFrame := vfenceSheet.Bounds()
	hfenceSheet := Pictures["./images/horizontalfence.png"]
	hfenceBatchBot := pixel.NewBatch(&pixel.TrianglesData{}, hfenceSheet)
	hfenceBatchTop := pixel.NewBatch(&pixel.TrianglesData{}, hfenceSheet)
	hfenceFrame := hfenceSheet.Bounds()

	for x := 0; x <= 13; x++ {
		for y := 0; y <= 13; y++ {
			pos := pixel.V(float64(x*320)-160, float64(y*320)-160)
			bread := pixel.NewSprite(grassSheet, grassFrames)
			bread.Draw(grassBatch, pixel.IM.Moved(pos))
		}
	}

	for x := 0; x < 31; x++ {
		top := pixel.V(float64(x*128)+64, 4000)
		bottom := pixel.V(float64(x*128)+64, 0)
		fence := pixel.NewSprite(hfenceSheet, hfenceFrame)
		fence.Draw(hfenceBatchTop, pixel.IM.Moved(top))
		fence.Draw(hfenceBatchBot, pixel.IM.Moved(bottom))
	}

	for x := 0; x < 31; x++ {
		left := pixel.V(-9, float64(x*128)+64)
		rigth := pixel.V(4010, float64(x*128)+64)
		fence := pixel.NewSprite(vfenceSheet, vfenceFrame)
		fence.Draw(vfenceBatch, pixel.IM.Moved(left))
		fence.Draw(vfenceBatch, pixel.IM.Moved(rigth))
	}

	pathTreeLength := 22
	treeSeparation := 110
	pathTop := .0
	// Make path
	for i := 0; i <= pathTreeLength; i++ {
		h := float64(300 + i*treeSeparation)
		pos1 := pixel.V(1850, h)
		pos2 := pixel.V(2150, h)

		tree := pixel.NewSprite(treeSheet, treeFrames[rand.Intn(len(treeFrames))])
		tree.Draw(treeBatch, pixel.IM.Scaled(pixel.ZV, 3.5).Moved(pos1))
		tree.Draw(treeBatch, pixel.IM.Scaled(pixel.ZV, 3.5).Moved(pos2))
		pathTop = h
	}

	// Make "arena"
	for i := 0; i <= 18; i++ {
		w := float64(1000 + i*treeSeparation)
		top := pixel.V(w, 3850)
		bottom := pixel.V(w, pathTop)
		tree := pixel.NewSprite(treeSheet, treeFrames[rand.Intn(len(treeFrames))])
		tree.Draw(treeBatch, pixel.IM.Scaled(pixel.ZV, 3.5).Moved(top))
		if bottom.X < 1800 || bottom.X > 2200 {
			tree.Draw(treeBatch, pixel.IM.Scaled(pixel.ZV, 3.5).Moved(bottom))
		}
	}
	arenaTop := .0
	for i := 0; i <= 10; i++ {
		h := pathTop + float64(i*treeSeparation)
		left := pixel.V(1000, h)
		rigth := pixel.V(3000, h)
		tree := pixel.NewSprite(treeSheet, treeFrames[rand.Intn(len(treeFrames))])
		tree.Draw(treeBatch, pixel.IM.Scaled(pixel.ZV, 3.5).Moved(left))
		tree.Draw(treeBatch, pixel.IM.Scaled(pixel.ZV, 3.5).Moved(rigth))
		arenaTop = h
	}

	// Fill outside
	position := generateRandomPositions(200, 0, int64(pathTop)-100, 0, 1800)
	position = append(position, generateRandomPositions(200, 0, int64(pathTop)-100, 2200, 4000)...)
	for i := range position {
		tree := pixel.NewSprite(treeSheet, treeFrames[rand.Intn(len(treeFrames))])
		tree.Draw(treeBatch, pixel.IM.Scaled(pixel.ZV, 3.5).Moved(position[i]))
	}

	// fill with zombie trees
	zombieTrees := generateRandomPositions(20, int64(pathTop)+100, int64(arenaTop), 1100, 2900)
	ztree := trees[ZombieTree]
	for i := 0; i <= len(zombieTrees)-1; i++ {
		tree := pixel.NewSprite(*ztree.Pic, ztree.Frame)
		tree.Draw(ztree.Batch, pixel.IM.Moved(zombieTrees[i]))
		cs.Insert(&Bounds{
			Pos:    zombieTrees[i],
			Offset: pixel.V(-18, -60),
			Height: 24,
			Width:  46,
		})
	}

	// fill with tall trees
	// commented because they were ugly

	// tallTress := generateRandomPositions(10, int64(pathTop)+100, int64(arenaTop), 1100, 2900)
	// ttree := trees[TallTree]
	// for i := 0; i <= len(tallTress)-1; i++ {
	// 	tree := pixel.NewSprite(*ttree.Pic, ttree.Frame)
	// 	tree.Draw(ttree.Batch, pixel.IM.Scaled(pixel.ZV, 1.2).Moved(tallTress[i]))
	// }

	nltallTress := generateRandomPositions(20, int64(pathTop)+100, int64(arenaTop), 1100, 2900)
	nlttree := trees[NoLeafsTallTree]
	for i := 0; i <= len(nltallTress)-1; i++ {
		tree := pixel.NewSprite(*nlttree.Pic, nlttree.Frame)
		tree.Draw(nlttree.Batch, pixel.IM.Scaled(pixel.ZV, 1.2).Moved(nltallTress[i]))
		cs.Insert(&Bounds{
			Pos:    nltallTress[i],
			Offset: pixel.V(-20, -129),
			Height: 20,
			Width:  26,
		})
	}

	return &Forest{
		Trees:          trees,
		Pic:            treeSheet,
		Frames:         treeFrames,
		Batch:          treeBatch,
		GrassBatch:     grassBatch,
		GrassFrames:    grassFrames,
		GrassPic:       grassSheet,
		FenceBatchHTOP: hfenceBatchTop,
		FenceBatchHBOT: hfenceBatchBot,
		FenceFrameH:    hfenceFrame,
		FencePicH:      hfenceSheet,
		FenceBatchV:    vfenceBatch,
		FenceFrameV:    vfenceFrame,
		FencePicV:      vfenceSheet,
	}
}
