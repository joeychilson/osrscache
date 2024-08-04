package openrs2

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"

	"github.com/joeychilson/osrscache"
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

func (s *OpenRS2Store) ArchiveList() ([]int, error) {
	var archives []int
	err := filepath.Walk(s.path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() && archiveNameRegex.MatchString(info.Name()) {
			if id, err := strconv.Atoi(info.Name()); err == nil {
				archives = append(archives, id)
			}
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to walk directory: %w", err)
	}
	sort.Ints(archives)
	return archives, nil
}

func (s *OpenRS2Store) ArchiveExists(archiveID int) bool {
	if archiveID < 0 || archiveID > osrscache.MaxArchive {
		return false
	}

	path := filepath.Join(s.path, strconv.Itoa(archiveID))
	fi, err := os.Stat(path)
	return err == nil && fi.IsDir()
}

func (s *OpenRS2Store) GroupList(archiveID int) ([]int, error) {
	path := filepath.Join(s.path, strconv.Itoa(archiveID))

	entries, err := os.ReadDir(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("archive does not exist")
		}
		return nil, fmt.Errorf("failed to read archive directory: %w", err)
	}

	groups := make([]int, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if len(name) > 4 && name[len(name)-4:] == groupExtension {
			if id, err := strconv.Atoi(name[:len(name)-4]); err == nil {
				groups = append(groups, id)
			}
		}
	}
	sort.Ints(groups)
	return groups, nil
}

func (s *OpenRS2Store) GroupExists(archiveID, groupID int) bool {
	if groupID < 0 {
		return false
	}

	path := filepath.Join(s.path, strconv.Itoa(archiveID), strconv.Itoa(groupID)+groupExtension)
	_, err := os.Stat(path)
	return err == nil
}

func (s *OpenRS2Store) Read(archiveID, groupID int) ([]byte, error) {
	path := filepath.Join(s.path, strconv.Itoa(archiveID), strconv.Itoa(groupID)+groupExtension)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("group does not exist: %w", err)
		}
		return nil, fmt.Errorf("failed to read group file: %w", err)
	}
	return data, nil
}
