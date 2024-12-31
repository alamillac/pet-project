package entities

import "github.com/alamillac/dummy-cat/components"

type Enemy struct {
	*Sprite
	FollowsPlayer bool
	CombatComp *components.EnemyCombat
}
