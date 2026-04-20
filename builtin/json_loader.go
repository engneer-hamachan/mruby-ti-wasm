package builtin

import (
	"encoding/json"
	"fmt"
	"io/fs"
)

func (ts *TypeSpec) UnmarshalJSON(data []byte) error {
	var single string
	if err := json.Unmarshal(data, &single); err == nil {
		*ts = []string{single}
		return nil
	}

	var array []string
	if err := json.Unmarshal(data, &array); err != nil {
		return err
	}
	*ts = array
	return nil
}

func LoadFromFS(fsys fs.FS) error {
	matches, err := fs.Glob(fsys, ".ti-config/*.json")
	if err != nil {
		return fmt.Errorf("failed to find JSON config files: %w", err)
	}

	for _, match := range matches {
		jsonData, err := fs.ReadFile(fsys, match)
		if err != nil {
			return fmt.Errorf("failed to read %s: %w", match, err)
		}

		if len(jsonData) == 0 {
			continue
		}

		var classDef ClassDefinition
		if err := json.Unmarshal(jsonData, &classDef); err != nil {
			return fmt.Errorf("failed to parse %s: %w", match, err)
		}

		loadClassDef(classDef)
	}

	return nil
}
