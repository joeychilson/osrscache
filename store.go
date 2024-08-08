package osrscache

const (
	MaxArchive = 255
)

type Store interface {
	ArchiveList() ([]uint8, error)
	ArchiveExists(archiveID uint8) bool
	GroupList(archiveID uint8) ([]uint32, error)
	GroupExists(archiveID uint8, groupID uint32) bool
	Read(archiveID uint8, groupID uint32) ([]byte, error)
}
