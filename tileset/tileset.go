package tileset

import (
	"encoding/json"
	"image"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/alamillac/dummy-cat/constants"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

type Tileset interface {
	Img(id int) *ebiten.Image
	GetTilecount() int
}

type UniformTilesetJSON struct {
	Path      string `json:"image"`
	TileCount int    `json:"tilecount"`
	Columns   int    `json:"columns"`
}

type UniformTileset struct {
	img       *ebiten.Image
	gid       int
	tilecount int
	columns   int
}

func (u *UniformTileset) Img(id int) *ebiten.Image {
	id -= u.gid

	srcX := id % u.columns
	srcY := id / u.columns

	srcX *= constants.Tilesize
	srcY *= constants.Tilesize

	return u.img.SubImage(
		image.Rect(
			srcX, srcY, srcX+constants.Tilesize, srcY+constants.Tilesize,
		),
	).(*ebiten.Image)
}

func (u *UniformTileset) GetTilecount() int {
	return u.tilecount
}

type TileJSON struct {
	Id     int    `json:"id"`
	Path   string `json:"image"`
	Width  int    `json:"imagewidth"`
	Height int    `json:"imageheight"`
}

type DynTilesetJSON struct {
	Tiles     []*TileJSON `json:"tiles"`
	TileCount int         `json:"tilecount"`
}

type DynTileset struct {
	imgs      []*ebiten.Image
	gid       int
	tilecount int
}

func (d *DynTileset) Img(id int) *ebiten.Image {
	id -= d.gid

	return d.imgs[id]
}

func (d *DynTileset) GetTilecount() int {
	return d.tilecount
}

func cleanPath(path string) string {
	path = filepath.Clean(path)
	path = strings.ReplaceAll(path, "\\", "/")
	path = strings.TrimPrefix(path, "../")
	path = strings.TrimPrefix(path, "../")
	return filepath.Join("assets/", path)
}

func NewTileset(path string, gid int) (Tileset, error) {
	contents, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	if strings.Contains(path, "buildings") { // TODO: Find a better way to load this
		// Dynamic tileset
		var dynTilesetJSON DynTilesetJSON
		err = json.Unmarshal(contents, &dynTilesetJSON)
		if err != nil {
			return nil, err
		}
		dynTileset := DynTileset{}
		dynTileset.gid = gid
		dynTileset.imgs = make([]*ebiten.Image, 0)
		dynTileset.tilecount = dynTilesetJSON.TileCount

		for _, tileJSON := range dynTilesetJSON.Tiles {
			tileJSONPath := cleanPath(tileJSON.Path)

			img, _, err := ebitenutil.NewImageFromFile(tileJSONPath)
			if err != nil {
				return nil, err
			}

			dynTileset.imgs = append(dynTileset.imgs, img)
		}

		return &dynTileset, nil
	}

	// Uniform tileset
	var uniformTilesetJSON UniformTilesetJSON
	err = json.Unmarshal(contents, &uniformTilesetJSON)
	if err != nil {
		return nil, err
	}

	uniformTileset := UniformTileset{}
	tileJSONPath := cleanPath(uniformTilesetJSON.Path)
	log.Printf("Loading image %s\n", tileJSONPath)
	img, _, err := ebitenutil.NewImageFromFile(tileJSONPath)
	if err != nil {
		return nil, err
	}
	uniformTileset.img = img
	uniformTileset.gid = gid
	uniformTileset.tilecount = uniformTilesetJSON.TileCount
	uniformTileset.columns = uniformTilesetJSON.Columns

	return &uniformTileset, nil
}
