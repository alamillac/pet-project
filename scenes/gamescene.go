package scenes

import (
	"fmt"
	"image"
	"image/color"
	"log"
	"math"

	"github.com/alamillac/dummy-cat/animations"
	"github.com/alamillac/dummy-cat/camera"
	"github.com/alamillac/dummy-cat/components"
	"github.com/alamillac/dummy-cat/constants"
	"github.com/alamillac/dummy-cat/entities"
	"github.com/alamillac/dummy-cat/spritesheet"
	"github.com/alamillac/dummy-cat/tilemap"
	"github.com/alamillac/dummy-cat/tileset"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

func CheckCollisionHorizontal(sprite *entities.Sprite, colliders []image.Rectangle) {
	for _, collider := range colliders {
		if collider.Overlaps(
			image.Rect(
				int(sprite.X),
				int(sprite.Y),
				int(sprite.X)+constants.Tilesize,
				int(sprite.Y)+constants.Tilesize,
			)) {
			if sprite.Dx > 0.0 {
				sprite.X = float64(collider.Min.X) - constants.Tilesize
			} else if sprite.Dx < 0.0 {
				sprite.X = float64(collider.Max.X)
			}
		}
	}
}

func CheckCollisionVertical(sprite *entities.Sprite, colliders []image.Rectangle) {
	for _, collider := range colliders {
		if collider.Overlaps(
			image.Rect(
				int(sprite.X),
				int(sprite.Y),
				int(sprite.X)+constants.Tilesize,
				int(sprite.Y)+constants.Tilesize,
			)) {
			if sprite.Dy > 0.0 {
				sprite.Y = float64(collider.Min.Y) - constants.Tilesize
			} else if sprite.Dy < 0.0 {
				sprite.Y = float64(collider.Max.Y)
			}
		}
	}
}

type GameScene struct {
	loaded            bool
	player            *entities.Player
	playerSpritesheet *spritesheet.SpriteSheet
	enemies           []*entities.Enemy
	potions           []*entities.Potion
	tilemapJSON       *tilemap.TilemapJSON
	tilesets          []tileset.Tileset
	cam               *camera.Camera
	colliders         []image.Rectangle
}

func NewGameScene() *GameScene {
	return &GameScene{
		player:            nil,
		playerSpritesheet: nil,
		enemies:           make([]*entities.Enemy, 0),
		potions:           make([]*entities.Potion, 0),
		tilemapJSON:       nil,
		tilesets:          nil,
		cam:               nil,
		colliders:         make([]image.Rectangle, 0),
		loaded:            false,
	}
}

func (g *GameScene) IsLoaded() bool {
	return g.loaded
}

func (g *GameScene) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{120, 180, 255, 255})
	// ebitenutil.DebugPrint(screen, "Hello, World!")

	opts := ebiten.DrawImageOptions{}

	for _, layer := range g.tilemapJSON.Layers {
		if layer.Type != "tilelayer" {
			//log.Printf("Invalid layer %s\n", layer.Name)
			continue
		}
		for index, id := range layer.Data {
			if id == 0 {
				continue
			}

			x := index % layer.Width
			y := index / layer.Width

			x *= constants.Tilesize
			y *= constants.Tilesize

			layerIndex, err := g.tilemapJSON.GetTilesetIndex(id)
			if err != nil {
				log.Fatal(fmt.Printf("Invalid id %d\n", id))
			}
			img := g.tilesets[layerIndex].Img(id)

			opts.GeoM.Translate(float64(x), float64(y))
			opts.GeoM.Translate(0.0, -(float64(img.Bounds().Dy() + constants.Tilesize)))
			opts.GeoM.Translate(g.cam.X, g.cam.Y)

			screen.DrawImage(img, &opts)

			opts.GeoM.Reset()
		}
	}

	opts.GeoM.Translate(g.player.X, g.player.Y)
	opts.GeoM.Translate(g.cam.X, g.cam.Y)

	playerFrame := 0

	playerActiveAnim := g.player.ActiveAnimation(int(g.player.Dx), int(g.player.Dy))
	if playerActiveAnim != nil {
		playerFrame = playerActiveAnim.Frame()
	}

	// Draw player
	screen.DrawImage(
		g.player.Img.SubImage(
			g.playerSpritesheet.Rect(
				playerFrame,
			),
		).(*ebiten.Image),
		&opts,
	)
	opts.GeoM.Reset()

	for _, sprite := range g.enemies {
		opts.GeoM.Translate(sprite.X, sprite.Y)
		opts.GeoM.Translate(g.cam.X, g.cam.Y)

		screen.DrawImage(
			sprite.Img.SubImage(
				image.Rect(0, 0, constants.Tilesize, constants.Tilesize),
			).(*ebiten.Image),
			&opts,
		)

		opts.GeoM.Reset()
	}

	for _, sprite := range g.potions {
		opts.GeoM.Translate(sprite.X, sprite.Y)
		opts.GeoM.Translate(g.cam.X, g.cam.Y)

		screen.DrawImage(
			sprite.Img.SubImage(
				image.Rect(0, 0, constants.Tilesize, constants.Tilesize),
			).(*ebiten.Image),
			&opts,
		)

		opts.GeoM.Reset()
	}

	for _, collider := range g.colliders {
		vector.StrokeRect(
			screen,
			float32(collider.Min.X)+float32(g.cam.X),
			float32(collider.Min.Y)+float32(g.cam.Y),
			float32(collider.Dx()),
			float32(collider.Dy()),
			1.0,
			color.RGBA{255, 0, 0, 255},
			true,
		)
	}
}

func (g *GameScene) FirstLoad() error {
	playerImg, _, err := ebitenutil.NewImageFromFile("assets/images/cat1.png")
	if err != nil {
		return err
	}

	skeletonImg, _, err := ebitenutil.NewImageFromFile("assets/images/skeleton.png")
	if err != nil {
		return err
	}

	potionImg, _, err := ebitenutil.NewImageFromFile("assets/images/potion.png")
	if err != nil {
		return err
	}

	tilemapJSON, err := tilemap.NewTilemapJSON("assets/maps/spawn.json")
	if err != nil {
		return err
	}

	tilesets, err := tilemapJSON.GenTilesets()
	if err != nil {
		return err
	}

	restAnimations := []animations.AnimationStep{}

	baseAnimation := *animations.NewAnimation(0, 42, 14, 7.0)
	restAnimations = append(restAnimations,
		animations.AnimationStep{
			Animation: baseAnimation,
			Delay:     300,
		},
		animations.AnimationStep{
			Animation: *animations.NewAnimation(1, 43, 14, 7.0),
			Delay:     300,
		},
		animations.AnimationStep{
			Animation: baseAnimation,
			Delay:     300,
		},
		animations.AnimationStep{
			Animation: *animations.NewAnimation(2, 44, 14, 7.0),
			Delay:     300,
		},
		animations.AnimationStep{
			Animation: baseAnimation,
			Delay:     300,
		},
		animations.AnimationStep{
			Animation: *animations.NewAnimation(3, 45, 14, 7.0),
			Delay:     300,
		},
		animations.AnimationStep{
			Animation: baseAnimation,
			Delay:     300,
		},
		animations.AnimationStep{
			Animation: *animations.NewAnimation(4, 46, 14, 7.0),
			Delay:     800,
		},
	)

	player := &entities.Player{
		Sprite: &entities.Sprite{
			Img: playerImg,
			X:   16,
			Y:   112,
		},
		Health: 3,
		Animations: map[entities.PlayerState]animations.Animation{
			entities.Rest:  animations.NewComposeAnimation(restAnimations),
			entities.Left:  animations.NewAnimation(6, 104, 14, 5.0),
			entities.Right: animations.NewAnimation(7, 105, 14, 5.0),
		},
		CombatComp: components.NewBasicCombat(3, 1),
	}
	enemies := []*entities.Enemy{
		{
			Sprite: &entities.Sprite{
				Img: skeletonImg,
				X:   47,
				Y:   29,
			},
			FollowsPlayer: true,
			CombatComp:    components.NewEnemyCombat(3, 1, 30),
		},
		{
			Sprite: &entities.Sprite{
				Img: skeletonImg,
				X:   90,
				Y:   20,
			},
			FollowsPlayer: false,
			CombatComp:    components.NewEnemyCombat(3, 1, 30),
		},
		{
			Sprite: &entities.Sprite{
				Img: skeletonImg,
				X:   30,
				Y:   30,
			},
			FollowsPlayer: true,
			CombatComp:    components.NewEnemyCombat(3, 1, 30),
		},
	}
	potions := []*entities.Potion{
		{
			Sprite: &entities.Sprite{
				Img: potionImg,
				X:   47,
				Y:   29,
			},
			AmountHeal: 1.0,
		},
	}

	playerSpriteSheet := spritesheet.NewSpriteSheet(14, 8, constants.Tilesize)

	g.player = player
	g.playerSpritesheet = playerSpriteSheet
	g.enemies = enemies
	g.potions = potions
	g.tilemapJSON = tilemapJSON
	g.tilesets = tilesets
	g.cam = camera.NewCamera(0.0, 0.0)

	g.colliders = []image.Rectangle{
		image.Rect(20, 20, 36, 36),
	}
	g.loaded = true

	return nil
}

func (g *GameScene) OnEnter() {
}

func (g *GameScene) OnExit() {
}

func (g *GameScene) Update() SceneId {
	if inpututil.IsKeyJustPressed(ebiten.KeyQ) {
		return ExitSceneId
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
		return PauseSceneId
	}

	g.player.Dx = 0.0
	g.player.Dy = 0.0

	if ebiten.IsKeyPressed(ebiten.KeyRight) {
		g.player.Dx = 2
	}
	if ebiten.IsKeyPressed(ebiten.KeyLeft) {
		g.player.Dx = -2
	}
	// if ebiten.IsKeyPressed(ebiten.KeyUp) {
	// 	g.player.Dy = -2
	// }
	// if ebiten.IsKeyPressed(ebiten.KeyDown) {
	// 	g.player.Dy = 2
	// }

	g.player.X += g.player.Dx
	CheckCollisionHorizontal(g.player.Sprite, g.colliders)

	g.player.Y += g.player.Dy
	CheckCollisionVertical(g.player.Sprite, g.colliders)

	playerActiveAnim := g.player.ActiveAnimation(int(g.player.Dx), int(g.player.Dy))
	if playerActiveAnim != nil {
		playerActiveAnim.Update()
	}

	for _, sprite := range g.enemies {

		sprite.Dx = 0.0
		sprite.Dy = 0.0

		if sprite.FollowsPlayer {
			if sprite.X < g.player.X {
				sprite.Dx = 1
			} else if sprite.X > g.player.X {
				sprite.Dx = -1
			}
			if sprite.Y < g.player.Y {
				sprite.Dy = 1
			} else if sprite.Y > g.player.Y {
				sprite.Dy = -1
			}
		}

		sprite.X += sprite.Dx
		CheckCollisionHorizontal(sprite.Sprite, g.colliders)

		sprite.Y += sprite.Dy
		CheckCollisionVertical(sprite.Sprite, g.colliders)
	}

	clicked := inpututil.IsMouseButtonJustPressed(ebiten.MouseButton0)
	cX, cY := ebiten.CursorPosition()
	cX += int(g.cam.X)
	cY -= int(g.cam.Y)
	g.player.CombatComp.Update()

	pRect := image.Rect(
		int(g.player.X),
		int(g.player.Y),
		int(g.player.X)+constants.Tilesize,
		int(g.player.Y)+constants.Tilesize,
	)

	// Combat action
	deadEnemies := make(map[int]struct{})
	for index, enemy := range g.enemies {
		enemy.CombatComp.Update()

		rect := image.Rect(
			int(enemy.X),
			int(enemy.Y),
			int(enemy.X)+constants.Tilesize,
			int(enemy.Y)+constants.Tilesize,
		)

		if rect.Overlaps(pRect) {
			if enemy.CombatComp.Attack() {
				g.player.CombatComp.Damage(enemy.CombatComp.AttackPower())
				fmt.Printf("Player damaged. Health: %d\n", g.player.CombatComp.Health())
				if g.player.CombatComp.Health() <= 0 {
					fmt.Println("Player has died!!")
				}
			}
		}

		// Is cursor in rect?
		if cX > rect.Min.X && cX < rect.Max.X && cY > rect.Min.Y && cY < rect.Max.Y {
			if clicked &&
				math.Sqrt(
					math.Pow(
						float64(cX)-g.player.X+(constants.Tilesize/2),
						2,
					)+math.Pow(
						float64(cY)-g.player.Y+(constants.Tilesize/2),
						2,
					),
				) < constants.Tilesize*5 {
				fmt.Println("Damaged enemy")
				enemy.CombatComp.Damage(g.player.CombatComp.AttackPower())

				if enemy.CombatComp.Health() <= 0 {
					deadEnemies[index] = struct{}{}
					fmt.Println("Enemy has been eliminated")
				}
			}
		}
	}
	if len(deadEnemies) > 0 {
		newEnemies := make([]*entities.Enemy, 0)
		for index, enemy := range g.enemies {
			if _, exists := deadEnemies[index]; !exists {
				newEnemies = append(newEnemies, enemy)
			}
		}
		g.enemies = newEnemies
	}

	// for _, potion := range g.potions {
	// 	if g.player.X > potion.X {
	// 		g.player.Health += potion.AmountHeal
	// 		fmt.Printf("Picked up potion! Health: %d\n", g.player.Health)
	// 	}
	// }

	// width := g.tilemapJSON.Layers[0].Width
	// height := g.tilemapJSON.Layers[0].Height
	width := 100 // TODO: get from layers or map
	height := 20 // TODO: get from layers or map
	g.cam.FollowTarget(g.player.X+8, g.player.Y+8, 320, 240)
	g.cam.Constrain(
		float64(width)*constants.Tilesize,
		float64(height)*constants.Tilesize,
		320,
		240,
	)

	return GameSceneId
}
