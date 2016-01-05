package tomlstructs

import (
	"github.com/BurntSushi/toml"
	"github.com/yangchenxing/foochow/structs"
)

func LoadTomlMap(path, include string) (map[string]interface{}, error) {
	return structs.LoadMap(path,
		func(path string) (map[string]interface{}, error) {
			data := make(map[string]interface{})
			_, err := toml.DecodeFile(path, &data)
			return data, err
		}, include)
}

func LoadTomlStruct(dest interface{}, path, include string) error {
	content, err := LoadTomlMap(path, include)
	if err != nil {
		return err
	}
	return structs.UnmarshalMap(dest, content)
}
