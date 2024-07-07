package osrscache

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Exporter[K comparable, V any] struct {
	definitions map[K]V
	outputDir   string
}

func NewExporter[K comparable, V any](definitions map[K]V, outputDir string) *Exporter[K, V] {
	return &Exporter[K, V]{
		definitions: definitions,
		outputDir:   outputDir,
	}
}

func (e *Exporter[K, V]) ExportAll(filename string) error {
	data, err := json.MarshalIndent(e.definitions, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling definitions: %w", err)
	}

	fullPath := filepath.Join(e.outputDir, filename)
	err = os.WriteFile(fullPath, data, 0644)
	if err != nil {
		return fmt.Errorf("writing file: %w", err)
	}
	return nil
}

func (e *Exporter[K, V]) ExportIndividual(prefix string) error {
	for id, def := range e.definitions {
		data, err := json.MarshalIndent(def, "", "  ")
		if err != nil {
			return fmt.Errorf("marshaling definition %v: %w", id, err)
		}

		filename := fmt.Sprintf("%s_%v.json", prefix, id)
		fullPath := filepath.Join(e.outputDir, filename)
		err = os.WriteFile(fullPath, data, 0644)
		if err != nil {
			return fmt.Errorf("writing file %s: %w", filename, err)
		}
	}
	return nil
}

func (e *Exporter[K, V]) ExportToJSON(mode string, filename string) error {
	if err := os.MkdirAll(e.outputDir, os.ModePerm); err != nil {
		return fmt.Errorf("creating output directory: %w", err)
	}

	switch mode {
	case "single":
		return e.ExportAll(filename)
	case "individual":
		return e.ExportIndividual(filename)
	default:
		return fmt.Errorf("invalid export mode: %s", mode)
	}
}
