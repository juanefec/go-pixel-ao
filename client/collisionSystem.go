package main

import "github.com/faiface/pixel"

type CollisionSystem struct {
	Bounds     Bounds
	MaxObjects int
	MaxLevels  int
	Level      int
	Objects    []Bounds
	Nodes      []CollisionSystem
	Total      int
}

type Bounds struct {
	pos    pixel.Vec
	Width  float64
	Height float64
}

func (b *Bounds) IsPoint() bool {

	if b.Width == 0 && b.Height == 0 {
		return true
	}

	return false

}

func (b *Bounds) Intersects(a Bounds) bool {

	aMaxX := a.pos.X + a.Width
	aMaxY := a.pos.Y + a.Height
	bMaxX := b.pos.X + b.Width
	bMaxY := b.pos.Y + b.Height

	if aMaxX < b.pos.X {
		return false
	}

	if a.pos.X > bMaxX {
		return false
	}

	if aMaxY < b.pos.Y {
		return false
	}

	if a.pos.Y > bMaxY {
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
	x := cs.Bounds.pos.X
	y := cs.Bounds.pos.Y

	cs.Nodes = append(cs.Nodes, CollisionSystem{
		Bounds: Bounds{
			pos:    pixel.V(x+subWidth, y),
			Width:  subWidth,
			Height: subHeight,
		},
		MaxObjects: cs.MaxObjects,
		MaxLevels:  cs.MaxLevels,
		Level:      nextLevel,
		Objects:    make([]Bounds, 0),
		Nodes:      make([]CollisionSystem, 0, 4),
	})

	cs.Nodes = append(cs.Nodes, CollisionSystem{
		Bounds: Bounds{
			pos:    pixel.V(x, y),
			Width:  subWidth,
			Height: subHeight,
		},
		MaxObjects: cs.MaxObjects,
		MaxLevels:  cs.MaxLevels,
		Level:      nextLevel,
		Objects:    make([]Bounds, 0),
		Nodes:      make([]CollisionSystem, 0, 4),
	})

	cs.Nodes = append(cs.Nodes, CollisionSystem{
		Bounds: Bounds{
			pos:    pixel.V(x, y+subHeight),
			Width:  subWidth,
			Height: subHeight,
		},
		MaxObjects: cs.MaxObjects,
		MaxLevels:  cs.MaxLevels,
		Level:      nextLevel,
		Objects:    make([]Bounds, 0),
		Nodes:      make([]CollisionSystem, 0, 4),
	})

	cs.Nodes = append(cs.Nodes, CollisionSystem{
		Bounds: Bounds{
			pos:    pixel.V(x+subWidth, y+subHeight),
			Width:  subWidth,
			Height: subHeight,
		},
		MaxObjects: cs.MaxObjects,
		MaxLevels:  cs.MaxLevels,
		Level:      nextLevel,
		Objects:    make([]Bounds, 0),
		Nodes:      make([]CollisionSystem, 0, 4),
	})

}

func (cs *CollisionSystem) getIndex(pRect Bounds) int {

	index := -1

	verticalMidpoint := cs.Bounds.pos.X + (cs.Bounds.Width / 2)
	horizontalMidpoint := cs.Bounds.pos.Y + (cs.Bounds.Height / 2)

	topQuadrant := (pRect.pos.Y < horizontalMidpoint) && (pRect.pos.Y+pRect.Height < horizontalMidpoint)

	bottomQuadrant := (pRect.pos.Y > horizontalMidpoint)

	if (pRect.pos.X < verticalMidpoint) && (pRect.pos.X+pRect.Width < verticalMidpoint) {

		if topQuadrant {
			index = 1
		} else if bottomQuadrant {
			index = 2
		}

	} else if pRect.pos.X > verticalMidpoint {

		if topQuadrant {
			index = 0
		} else if bottomQuadrant {
			index = 3
		}

	}

	return index

}

func (cs *CollisionSystem) Insert(pRect Bounds) {

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

func (cs *CollisionSystem) Retrieve(pRect Bounds) []Bounds {

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

func (cs *CollisionSystem) RetrievePoints(find Bounds) []Bounds {

	var foundPoints []Bounds
	potentials := cs.Retrieve(find)
	for o := 0; o < len(potentials); o++ {

		xyMatch := potentials[o].pos.X == float64(find.pos.X) && potentials[o].pos.Y == float64(find.pos.Y)
		if xyMatch && potentials[o].IsPoint() {
			foundPoints = append(foundPoints, find)
		}
	}

	return foundPoints

}

func (cs *CollisionSystem) RetrieveIntersections(find Bounds) []Bounds {

	var foundIntersections []Bounds

	potentials := cs.Retrieve(find)
	for o := 0; o < len(potentials); o++ {
		if potentials[o].Intersects(find) {
			foundIntersections = append(foundIntersections, potentials[o])
		}
	}

	return foundIntersections

}

func (cs *CollisionSystem) Clear() {

	cs.Objects = []Bounds{}

	if len(cs.Nodes)-1 > 0 {
		for i := 0; i < len(cs.Nodes); i++ {
			cs.Nodes[i].Clear()
		}
	}

	cs.Nodes = []CollisionSystem{}
	cs.Total = 0

}
