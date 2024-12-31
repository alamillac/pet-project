package tilemap

import (
	"encoding/json"
	"errors"
	"log"
	"os"
	"path"

	"github.com/alamillac/dummy-cat/tileset"
)

type TilemapLayerJSON struct {
	Type   string `json:"type"`
	Data   []int  `json:"data"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
	Name   string `json:"name"`
}

type TilesetRange struct {
	Min int
	Max int
}

type TilemapJSON struct {
	Layers        []TilemapLayerJSON `json:"layers"`
	Tilesets      []map[string]any   `json:"tilesets"`
	tilesetsRange []TilesetRange
}

func (t *TilemapJSON) GetTilesetIndex(id int) (int, error) {
	for idx, tr := range t.tilesetsRange {
		if id >= tr.Min && id < tr.Max {
			return idx, nil
		}
	}
	return 0, errors.New("Not found")
}

func (t *TilemapJSON) GenTilesets() ([]tileset.Tileset, error) {
	tilesets := make([]tileset.Tileset, 0)
	tilesetsRange := make([]TilesetRange, 0)

	for _, tilesetData := range t.Tilesets {
		gid := int(tilesetData["firstgid"].(float64))
		tilesetPath := path.Join("assets/maps/", tilesetData["source"].(string))
		log.Printf("Loading tileset %s\n", tilesetPath)
		tileset, err := tileset.NewTileset(tilesetPath, gid)
		if err != nil {
			return nil, err
		}

		tilesets = append(tilesets, tileset)
		tilesetsRange = append(tilesetsRange, TilesetRange{
			Min: gid,
			Max: gid + tileset.GetTilecount(),
		})
	}
	t.tilesetsRange = tilesetsRange
	return tilesets, nil
}

func NewTilemapJSON(filepath string) (*TilemapJSON, error) {
	contents, err := os.ReadFile(filepath)
	if err != nil {
		return nil, err
	}

	var tilemapJSON TilemapJSON
	err = json.Unmarshal(contents, &tilemapJSON)
	if err != nil {
		return nil, err
	}

	return &tilemapJSON, nil
}
