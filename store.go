package osrscache

const (
	MaxArchive = 255
)

type Store interface {
	ArchiveList() ([]int, error)
	ArchiveExists(archiveID int) bool
	GroupList(archiveID int) ([]int, error)
	GroupExists(archiveID int, groupID int) bool
	Read(archiveID int, groupID int) ([]byte, error)
}
