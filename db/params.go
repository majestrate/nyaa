package db

type WhereParams struct {
	NotNull        []string
	Category       int
	SubCategory    int
	UploaderID     int
	StatusGreater  int
	StatusNotEqual int
	NameLike       []string
}

type SortParams struct {
	OrderBy   string
	Ascending bool
}
