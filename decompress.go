package osrscache

import (
	"bytes"
	"compress/bzip2"
	"compress/gzip"
	"encoding/binary"
	"fmt"
	"io"
)

const (
	CompressionNone  = 0
	CompressionBZIP2 = 1
	CompressionGZIP  = 2
)

func DecompressArchiveData(data []byte) ([]byte, error) {
	if len(data) < 5 {
		return nil, fmt.Errorf("archive data too short: %d bytes", len(data))
	}

	compressionType := data[0]
	compressedLength := int(binary.BigEndian.Uint32(data[1:5]))

	if compressionType == CompressionNone {
		return data[5:], nil
	}

	if len(data) < compressedLength+9 {
		return nil, fmt.Errorf("archive data shorter than expected: %d < %d", len(data), compressedLength+9)
	}

	uncompressedLength := int(binary.BigEndian.Uint32(data[5:9]))
	compressedData := data[9 : compressedLength+9]

	var reader io.Reader

	switch compressionType {
	case CompressionGZIP:
		gzipReader, err := gzip.NewReader(bytes.NewReader(compressedData))
		if err != nil {
			return nil, fmt.Errorf("failed to create gzip reader: %w", err)
		}
		defer gzipReader.Close()
		reader = gzipReader
	case CompressionBZIP2:
		bzip2Header := []byte{'B', 'Z', 'h', '1'}
		customCompressedData := append(bzip2Header, compressedData...)
		reader = bzip2.NewReader(bytes.NewReader(customCompressedData))
	default:
		return nil, fmt.Errorf("unknown compression type: %d", compressionType)
	}

	uncompressedData := make([]byte, uncompressedLength)
	n, err := io.ReadFull(reader, uncompressedData)
	if err != nil {
		return nil, fmt.Errorf("failed to decompress data (read %d bytes): %w", n, err)
	}

	return uncompressedData, nil
}
