package osrscache

import (
	"encoding/json"
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"
)

type JSONExporter[K comparable, V any] struct {
	definitions map[K]V
	outputDir   string
}

func NewJSONExporter[K comparable, V any](definitions map[K]V, outputDir string) *JSONExporter[K, V] {
	return &JSONExporter[K, V]{
		definitions: definitions,
		outputDir:   outputDir,
	}
}

func (e *JSONExporter[K, V]) ExportAll(filename string) error {
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

func (e *JSONExporter[K, V]) ExportIndividual(prefix string) error {
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

func (e *JSONExporter[K, V]) ExportToJSON(mode string, filename string) error {
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

type ImageExportable interface {
	Image() *image.RGBA
}

type ImageExporter[K comparable, V ImageExportable] struct {
	definitions map[K]V
	outputDir   string
}

func NewImageExporter[K comparable, V ImageExportable](definitions map[K]V, outputDir string) *ImageExporter[K, V] {
	return &ImageExporter[K, V]{
		definitions: definitions,
		outputDir:   outputDir,
	}
}

func (e *ImageExporter[K, V]) ExportToImage(prefix string) error {
	if err := os.MkdirAll(e.outputDir, os.ModePerm); err != nil {
		return fmt.Errorf("creating output directory: %w", err)
	}

	for id, def := range e.definitions {
		img := def.Image()
		filename := fmt.Sprintf("%s_%v.png", prefix, id)
		fullPath := filepath.Join(e.outputDir, filename)

		f, err := os.Create(fullPath)
		if err != nil {
			return fmt.Errorf("creating file %s: %w", filename, err)
		}
		defer f.Close()

		if err := png.Encode(f, img); err != nil {
			return fmt.Errorf("encoding image %s: %w", filename, err)
		}
	}
	return nil
}
