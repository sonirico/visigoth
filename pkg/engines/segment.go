package engines

type lsmtSegmentConfig interface {
	Path() string
	Size() int
}

type lsmtSegment interface {
	Name() string
	Open() error
	Close() error
	Insert(key string, val string) error
	Search(key string) (string, error)
	Compare(lsmtSegment) int
}
