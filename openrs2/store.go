package openrs2

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
)

const (
	groupExtension = ".dat"
)

var (
	archiveNameRegex = regexp.MustCompile(`^[0-9]+$`)
	groupNameRegex   = regexp.MustCompile(`^(\d+)\.dat$`)
)

type OpenRS2Store struct {
	path string
}

func Open(path string) (*OpenRS2Store, error) {
	return &OpenRS2Store{path: path}, nil
}

func (s *OpenRS2Store) ArchiveList() ([]uint8, error) {
	var archives []uint8
	err := filepath.Walk(s.path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() && archiveNameRegex.MatchString(info.Name()) {
			if id, err := strconv.Atoi(info.Name()); err == nil {
				archives = append(archives, uint8(id))
			}
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to walk directory: %w", err)
	}
	slices.Sort(archives)
	return archives, nil
}

func (s *OpenRS2Store) ArchiveExists(archiveID uint8) bool {
	path := filepath.Join(s.path, strconv.Itoa(int(archiveID)))
	fi, err := os.Stat(path)
	return err == nil && fi.IsDir()
}

func (s *OpenRS2Store) GroupList(archiveID uint8) ([]uint32, error) {
	path := filepath.Join(s.path, strconv.Itoa(int(archiveID)))

	entries, err := os.ReadDir(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("archive does not exist")
		}
		return nil, fmt.Errorf("failed to read archive directory: %w", err)
	}

	groups := make([]uint32, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if len(name) > 4 && name[len(name)-4:] == groupExtension {
			if id, err := strconv.Atoi(name[:len(name)-4]); err == nil {
				groups = append(groups, uint32(id))
			}
		}
	}
	slices.Sort(groups)
	return groups, nil
}

func (s *OpenRS2Store) GroupExists(archiveID uint8, groupID uint32) bool {
	path := filepath.Join(s.path, strconv.Itoa(int(archiveID)), strconv.Itoa(int(groupID))+groupExtension)
	_, err := os.Stat(path)
	return err == nil
}

func (s *OpenRS2Store) Read(archiveID uint8, groupID uint32) ([]byte, error) {
	path := filepath.Join(s.path, strconv.Itoa(int(archiveID)), strconv.Itoa(int(groupID))+groupExtension)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("group does not exist: %w", err)
		}
		return nil, fmt.Errorf("failed to read group file: %w", err)
	}
	return data, nil
}
