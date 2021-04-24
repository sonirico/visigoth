package engines

type lsmtSegmentOpts struct {
	SizeLimit   int
	SegmentPath string
}

func (o lsmtSegmentOpts) Path() string {
	return o.SegmentPath
}

func (o lsmtSegmentOpts) Size() int {
	return o.SizeLimit
}
