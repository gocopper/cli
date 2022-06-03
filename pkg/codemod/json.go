package codemod

import (
	"encoding/json"
	"os"

	"github.com/gocopper/copper/cerrors"
)

func AddJSONSection(fp, section string, data interface{}) error {
	var fileData map[string]interface{}

	f, err := os.OpenFile(fp, os.O_RDWR, 0644)
	if err != nil {
		return cerrors.New(err, "failed to open file", map[string]interface{}{
			"fp": fp,
		})
	}
	defer f.Close()

	err = json.NewDecoder(f).Decode(&fileData)
	if err != nil {
		return cerrors.New(err, "failed to decode json", map[string]interface{}{
			"fp": fp,
		})
	}

	err = f.Truncate(0)
	if err != nil {
		return cerrors.New(err, "failed to truncate file", map[string]interface{}{
			"fp": fp,
		})
	}

	_, err = f.Seek(0, 0)
	if err != nil {
		return cerrors.New(err, "failed to seek to 0", map[string]interface{}{
			"fp": fp,
		})
	}

	fileData[section] = data

	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	err = enc.Encode(&fileData)
	if err != nil {
		return cerrors.New(err, "failed to encode & write json file", map[string]interface{}{
			"fp": fp,
		})
	}

	return nil
}
