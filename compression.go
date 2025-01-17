package osrscache

import (
	"bytes"
	"compress/bzip2"
	"compress/gzip"
	"fmt"
	"io"
)

const (
	CompressionNone  = 0
	CompressionBZIP2 = 1
	CompressionGZIP  = 2
)

func DecompressData(data []byte) ([]byte, error) {
	reader := NewReader(data)

	compressionType, err := reader.ReadByte()
	if err != nil {
		return nil, fmt.Errorf("failed to read compression type: %w", err)
	}

	compressedLength, err := reader.ReadUint32()
	if err != nil {
		return nil, fmt.Errorf("failed to read compressed length: %w", err)
	}

	if compressionType == CompressionNone {
		uncompressedData := make([]byte, compressedLength)
		_, err := io.ReadFull(reader, uncompressedData)
		if err != nil {
			return nil, fmt.Errorf("failed to read uncompressed data: %w", err)
		}
		return uncompressedData, nil
	}

	uncompressedLength, err := reader.ReadUint32()
	if err != nil {
		return nil, fmt.Errorf("failed to read uncompressed length: %w", err)
	}

	if uint32(reader.Len()) < compressedLength {
		return nil, fmt.Errorf("archive data shorter than expected: %d < %d", reader.Len(), compressedLength)
	}

	var decompressor io.Reader

	switch compressionType {
	case CompressionGZIP:
		gzipReader, err := gzip.NewReader(reader)
		if err != nil {
			return nil, fmt.Errorf("failed to create gzip reader: %w", err)
		}
		defer gzipReader.Close()
		decompressor = gzipReader
	case CompressionBZIP2:
		bzip2Header := []byte{'B', 'Z', 'h', '1'}
		decompressor = bzip2.NewReader(io.MultiReader(bytes.NewReader(bzip2Header), reader))
	default:
		return nil, fmt.Errorf("unknown compression type: %d", compressionType)
	}

	uncompressedData := make([]byte, uncompressedLength)
	n, err := io.ReadFull(decompressor, uncompressedData)
	if err != nil {
		return nil, fmt.Errorf("failed to decompress data (read %d bytes): %w", n, err)
	}
	return uncompressedData, nil
}
