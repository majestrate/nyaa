package common

import "strconv"

type Status uint8

const (
	ShowAll Status = iota
	FilterRemakes
	Trusted
	APlus
)

type SortMode uint8

const (
	ID SortMode = iota
	Name
	Date
	Downloads
	Size
	Seeders
	Leechers
	Completed
)

type Category struct {
	Main, Sub uint8
}

func (c Category) String() (s string) {
	if c.Main != 0 {
		s += strconv.Itoa(int(c.Main))
	}
	s += "_"
	if c.Sub != 0 {
		s += strconv.Itoa(int(c.Sub))
	}
	return
}

func (c Category) IsSet() bool {
	return c.Main != 0 && c.Sub != 0
}

type SearchParam struct {
	Order    bool // True means acsending
	Status   Status
	Sort     SortMode
	Category Category
	Offset   uint32
	UserID   uint32
	Max      uint32
	NotNull  string // csv
	Null     string // csv
	NameLike string // csv
}
