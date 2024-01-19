package partupload

const (
	StatusProgress Status = "progress"
	StatusComplete Status = "done"
	StatusCanceled Status = "canceled"
)

// Status is an upload status name.
type Status string

func (s Status) String() string {
	return string(s)
}
