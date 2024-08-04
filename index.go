package osrscache

import (
	"fmt"
)

const (
	DigestBits  = 512
	DigestBytes = DigestBits >> 3
)

type Protocol uint8

const (
	ProtocolOriginal Protocol = iota + 5
	ProtocolVersioned
	ProtocolSmart
)

func ProtocolFromID(id uint8) (Protocol, error) {
	if Protocol(id) >= ProtocolOriginal && Protocol(id) <= ProtocolSmart {
		return Protocol(id), nil
	}
	return 0, fmt.Errorf("unknown protocol id: %d", id)
}

type Index struct {
	Protocol                 Protocol
	Version                  uint32
	HasNames                 bool
	HasDigests               bool
	HasLengths               bool
	HasUncompressedChecksums bool
	Groups                   []*Group
}

const (
	FlagNames                 = 0x01
	FlagDigests               = 0x02
	FlagLengths               = 0x04
	FlagUncompressedChecksums = 0x08
)

func NewIndex(data []byte) (*Index, error) {
	reader := NewReader(data)

	protocolID, err := reader.ReadUint8()
	if err != nil {
		return nil, fmt.Errorf("reading protocol id: %w", err)
	}

	protocol, err := ProtocolFromID(protocolID)
	if err != nil {
		return nil, fmt.Errorf("getting protocol: %w", err)
	}

	var version uint32
	if protocol >= ProtocolVersioned {
		version, err = reader.ReadUint32()
		if err != nil {
			return nil, fmt.Errorf("reading version: %w", err)
		}
	}

	flags, err := reader.ReadUint8()
	if err != nil {
		return nil, fmt.Errorf("reading flags: %w", err)
	}

	size, err := readSize(reader, protocol)
	if err != nil {
		return nil, fmt.Errorf("reading size: %w", err)
	}

	index := &Index{
		Protocol:                 protocol,
		Version:                  version,
		HasNames:                 flags&FlagNames != 0,
		HasDigests:               flags&FlagDigests != 0,
		HasLengths:               flags&FlagLengths != 0,
		HasUncompressedChecksums: flags&FlagUncompressedChecksums != 0,
		Groups:                   make([]*Group, size),
	}

	prevGroupID := 0
	for i := 0; i < size; i++ {
		delta, err := readSize(reader, protocol)
		if err != nil {
			return nil, fmt.Errorf("reading delta: %w", err)
		}
		groupID := prevGroupID + delta
		index.Groups[i] = &Group{ID: groupID, Files: make([]*File, 0)}
		prevGroupID = groupID
	}

	if index.HasNames {
		for i := range index.Groups {
			nameHash, err := reader.ReadInt32()
			if err != nil {
				return nil, fmt.Errorf("reading name hash: %w", err)
			}
			index.Groups[i].NameHash = nameHash
		}
	}

	for i := range index.Groups {
		checksum, err := reader.ReadInt32()
		if err != nil {
			return nil, fmt.Errorf("reading checksum: %w", err)
		}
		index.Groups[i].Checksum = checksum
	}

	if index.HasUncompressedChecksums {
		for i := range index.Groups {
			uncompressedChecksum, err := reader.ReadInt32()
			if err != nil {
				return nil, fmt.Errorf("reading uncompressed checksum: %w", err)
			}
			index.Groups[i].UncompressedChecksum = uncompressedChecksum
		}
	}

	if index.HasDigests {
		for i := range index.Groups {
			digest, err := reader.ReadBytes(DigestBytes)
			if err != nil {
				return nil, fmt.Errorf("reading digest: %w", err)
			}
			index.Groups[i].Digest = digest
		}
	}

	if index.HasLengths {
		for i := range index.Groups {
			length, err := reader.ReadInt32()
			if err != nil {
				return nil, fmt.Errorf("reading length: %w", err)
			}
			index.Groups[i].Length = length

			uncompressedLength, err := reader.ReadInt32()
			if err != nil {
				return nil, fmt.Errorf("reading uncompressed length: %w", err)
			}
			index.Groups[i].UncompressedLength = uncompressedLength
		}
	}

	for i := range index.Groups {
		version, err := reader.ReadInt32()
		if err != nil {
			return nil, fmt.Errorf("reading version: %w", err)
		}
		index.Groups[i].Version = version
	}

	groupSizes := make([]int, size)
	for i := range groupSizes {
		groupSize, err := readSize(reader, protocol)
		if err != nil {
			return nil, fmt.Errorf("reading group size: %w", err)
		}
		groupSizes[i] = groupSize
	}

	for i, group := range index.Groups {
		groupSize := groupSizes[i]

		prevFileID := 0
		for j := 0; j < groupSize; j++ {
			delta, err := readSize(reader, protocol)
			if err != nil {
				return nil, fmt.Errorf("reading file id delta: %w", err)
			}
			prevFileID += delta
			group.Files = append(group.Files, &File{ID: prevFileID})
		}
	}

	if index.HasNames {
		for _, group := range index.Groups {
			for _, file := range group.Files {
				nameHash, err := reader.ReadInt32()
				if err != nil {
					return nil, fmt.Errorf("reading file name hash: %w", err)
				}
				file.NameHash = nameHash
			}
		}
	}

	return index, nil
}

func (i *Index) Group(id int) (*Group, error) {
	for _, group := range i.Groups {
		if group.ID == id {
			return group, nil
		}
	}
	return nil, fmt.Errorf("group %d not found", id)
}

func readSize(reader *Reader, protocol Protocol) (int, error) {
	if protocol >= ProtocolSmart {
		size, err := reader.ReadSmartUint()
		if err != nil {
			return 0, fmt.Errorf("reading size: %w", err)
		}
		return int(size), nil
	} else {
		size, err := reader.ReadUint16()
		if err != nil {
			return 0, fmt.Errorf("reading size: %w", err)
		}
		return int(size), nil
	}
}

type Group struct {
	ID                   int
	NameHash             int32
	Version              int32
	Checksum             int32
	UncompressedChecksum int32
	Length               int32
	UncompressedLength   int32
	Digest               []byte
	Files                []*File
}

type File struct {
	ID       int
	NameHash int32
}

func (g *Group) Unpack(data []byte) (map[int][]byte, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("data must be readable")
	}

	if len(g.Files) < 1 {
		return nil, fmt.Errorf("group must have at least one file")
	}

	if len(g.Files) == 1 {
		return map[int][]byte{g.Files[0].ID: data}, nil
	}

	stripes := int(data[len(data)-1])
	trailerIndex := len(data) - (stripes * len(g.Files) * 4) - 1

	if trailerIndex < 0 {
		return nil, fmt.Errorf("invalid trailer index")
	}

	lens := make([]int, len(g.Files))
	reader := NewReader(data[trailerIndex:])
	for i := 0; i < stripes; i++ {
		prevLen := 0
		for j := range lens {
			delta, err := reader.ReadInt32()
			if err != nil {
				return nil, fmt.Errorf("reading data delta: %w", err)
			}
			prevLen += int(delta)
			lens[j] += prevLen
		}
	}

	files := make(map[int][]byte, len(g.Files))
	for i, file := range g.Files {
		files[file.ID] = make([]byte, 0, lens[i])
	}

	dataIndex := 0
	reader.Reset(data[trailerIndex:])
	for i := 0; i < stripes; i++ {
		prevLen := 0
		for _, file := range g.Files {
			delta, err := reader.ReadInt32()
			if err != nil {
				return nil, fmt.Errorf("reading data delta: %w", err)
			}
			prevLen += int(delta)
			end := dataIndex + prevLen
			if end > trailerIndex {
				return nil, fmt.Errorf("data overflow")
			}
			files[file.ID] = append(files[file.ID], data[dataIndex:end]...)
			dataIndex = end
		}
	}

	if dataIndex != trailerIndex {
		return nil, fmt.Errorf("data index does not match trailer index")
	}

	return files, nil
}
