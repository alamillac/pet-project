package entities

import (
	"github.com/alamillac/dummy-cat/animations"
	"github.com/alamillac/dummy-cat/components"
)

type PlayerState uint8

const (
	Rest PlayerState = iota
	Left
	Right
)

type Player struct {
	*Sprite
	Health     uint
	Animations map[PlayerState]animations.Animation
	CombatComp *components.BasicCombat
}

func (p *Player) ActiveAnimation(dx, dy int) animations.Animation {
	if dx > 0 {
		return p.Animations[Right]
	}
	if dx < 0 {
		return p.Animations[Left]
	}
	return p.Animations[Rest]
}
