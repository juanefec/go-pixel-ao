package main

import (
	"github.com/faiface/pixel"
)

type CollisionSystem struct {
	Bounds     Bounds
	MaxObjects int
	MaxLevels  int
	Level      int
	Objects    []*Bounds
	Nodes      []CollisionSystem
	Total      int
}

type Bounds struct {
	Pos    pixel.Vec
	Offset pixel.Vec
	Width  float64
	Height float64
}

func (b *Bounds) GetHitBoxX() float64 {
	return b.Pos.X + b.Offset.X
}

func (b *Bounds) GetHitBoxY() float64 {
	return b.Pos.Y + b.Offset.Y
}

func (b *Bounds) IsPoint() bool {
	if b.Width == 0 && b.Height == 0 {
		return true
	}

	return false
}

func (b *Bounds) Intersects(a *Bounds) bool {
	aMaxX := a.GetHitBoxX() + a.Width
	aMaxY := a.GetHitBoxY() + a.Height
	bMaxX := b.GetHitBoxX() + b.Width
	bMaxY := b.GetHitBoxY() + b.Height

	if a == b {
		return false
	}

	if aMaxX < b.GetHitBoxX() {
		return false
	}

	if a.GetHitBoxX() > bMaxX {
		return false
	}

	if aMaxY < b.GetHitBoxY() {
		return false
	}

	if a.GetHitBoxY() > bMaxY {
		return false
	}

	return true
}

func (cs *CollisionSystem) TotalNodes() int {
	total := 0

	if len(cs.Nodes) > 0 {
		for i := 0; i < len(cs.Nodes); i++ {
			total += 1
			total += cs.Nodes[i].TotalNodes()
		}
	}

	return total
}

func (cs *CollisionSystem) split() {
	if len(cs.Nodes) == 4 {
		return
	}

	nextLevel := cs.Level + 1
	subWidth := cs.Bounds.Width / 2
	subHeight := cs.Bounds.Height / 2
	x := cs.Bounds.Pos.X
	y := cs.Bounds.Pos.Y

	cs.Nodes = append(cs.Nodes, CollisionSystem{
		Bounds: Bounds{
			Pos:    pixel.V(x+subWidth, y),
			Width:  subWidth,
			Height: subHeight,
		},
		MaxObjects: cs.MaxObjects,
		MaxLevels:  cs.MaxLevels,
		Level:      nextLevel,
		Objects:    make([]*Bounds, 0),
		Nodes:      make([]CollisionSystem, 0, 4),
	})

	cs.Nodes = append(cs.Nodes, CollisionSystem{
		Bounds: Bounds{
			Pos:    pixel.V(x, y),
			Width:  subWidth,
			Height: subHeight,
		},
		MaxObjects: cs.MaxObjects,
		MaxLevels:  cs.MaxLevels,
		Level:      nextLevel,
		Objects:    make([]*Bounds, 0),
		Nodes:      make([]CollisionSystem, 0, 4),
	})

	cs.Nodes = append(cs.Nodes, CollisionSystem{
		Bounds: Bounds{
			Pos:    pixel.V(x, y+subHeight),
			Width:  subWidth,
			Height: subHeight,
		},
		MaxObjects: cs.MaxObjects,
		MaxLevels:  cs.MaxLevels,
		Level:      nextLevel,
		Objects:    make([]*Bounds, 0),
		Nodes:      make([]CollisionSystem, 0, 4),
	})

	cs.Nodes = append(cs.Nodes, CollisionSystem{
		Bounds: Bounds{
			Pos:    pixel.V(x+subWidth, y+subHeight),
			Width:  subWidth,
			Height: subHeight,
		},
		MaxObjects: cs.MaxObjects,
		MaxLevels:  cs.MaxLevels,
		Level:      nextLevel,
		Objects:    make([]*Bounds, 0),
		Nodes:      make([]CollisionSystem, 0, 4),
	})
}

func (cs *CollisionSystem) getIndex(pRect *Bounds) int {
	index := -1

	verticalMidpoint := cs.Bounds.Pos.X + (cs.Bounds.Width / 2)
	horizontalMidpoint := cs.Bounds.Pos.Y + (cs.Bounds.Height / 2)

	topQuadrant := (pRect.GetHitBoxY() < horizontalMidpoint) && (pRect.GetHitBoxY()+pRect.Height < horizontalMidpoint)

	bottomQuadrant := (pRect.GetHitBoxY() > horizontalMidpoint)

	if (pRect.GetHitBoxX() < verticalMidpoint) && (pRect.GetHitBoxX()+pRect.Width < verticalMidpoint) {

		if topQuadrant {
			index = 1
		} else if bottomQuadrant {
			index = 2
		}

	} else if pRect.GetHitBoxX() > verticalMidpoint {

		if topQuadrant {
			index = 0
		} else if bottomQuadrant {
			index = 3
		}

	}

	return index
}

func (cs *CollisionSystem) Insert(pRect *Bounds) {
	cs.Total++

	i := 0
	var index int

	if len(cs.Nodes) > 0 == true {

		index = cs.getIndex(pRect)

		if index != -1 {
			cs.Nodes[index].Insert(pRect)
			return
		}
	}

	cs.Objects = append(cs.Objects, pRect)

	if (len(cs.Objects) > cs.MaxObjects) && (cs.Level < cs.MaxLevels) {

		if len(cs.Nodes) > 0 == false {
			cs.split()
		}

		for i < len(cs.Objects) {

			index = cs.getIndex(cs.Objects[i])

			if index != -1 {

				splice := cs.Objects[i]
				cs.Objects = append(cs.Objects[:i], cs.Objects[i+1:]...)

				cs.Nodes[index].Insert(splice)

			} else {

				i++

			}

		}

	}
}

func (cs *CollisionSystem) Retrieve(pRect *Bounds) []*Bounds {
	index := cs.getIndex(pRect)

	returnObjects := cs.Objects

	if len(cs.Nodes) > 0 {

		if index != -1 {

			returnObjects = append(returnObjects, cs.Nodes[index].Retrieve(pRect)...)

		} else {

			for i := 0; i < len(cs.Nodes); i++ {
				returnObjects = append(returnObjects, cs.Nodes[i].Retrieve(pRect)...)
			}

		}
	}

	return returnObjects
}

func (cs *CollisionSystem) RetrievePoints(find *Bounds) []*Bounds {
	var foundPoints []*Bounds
	potentials := cs.Retrieve(find)
	for o := 0; o < len(potentials); o++ {

		xyMatch := potentials[o].Pos.X == float64(find.GetHitBoxX()) && potentials[o].Pos.Y == float64(find.GetHitBoxY())
		if xyMatch && potentials[o].IsPoint() {
			foundPoints = append(foundPoints, find)
		}
	}

	return foundPoints
}

func (cs *CollisionSystem) RetrieveIntersections(find *Bounds) []*Bounds {
	var foundIntersections []*Bounds

	potentials := cs.Retrieve(find)
	for o := 0; o < len(potentials); o++ {
		if potentials[o].Intersects(find) {
			foundIntersections = append(foundIntersections, potentials[o])
		}
	}

	return foundIntersections
}

func (cs *CollisionSystem) Clear() {
	cs.Objects = []*Bounds{}

	if len(cs.Nodes)-1 > 0 {
		for i := 0; i < len(cs.Nodes); i++ {
			cs.Nodes[i].Clear()
		}
	}

	cs.Nodes = []CollisionSystem{}
	cs.Total = 0
}

func (cs *CollisionSystem) GetAllBounds() []*Bounds {
	if len(cs.Nodes) == 0 {
		return cs.Objects
	} else {
		return append(cs.Nodes[0].GetAllBounds(), append(cs.Nodes[1].GetAllBounds(), append(cs.Nodes[2].GetAllBounds(), cs.Nodes[3].GetAllBounds()...)...)...)
	}
}
