package ossucs

// Course is one course in the OSSU Computer Science curriculum.
type Course struct {
	Rank    int    `json:"rank"    csv:"rank"    tsv:"rank"`
	Section string `json:"section" csv:"section" tsv:"section"`
	Title   string `json:"title"   csv:"title"   tsv:"title"`
	URL     string `json:"url"     csv:"url"     tsv:"url"`
}
