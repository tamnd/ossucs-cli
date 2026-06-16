package ossucs

// Course is one course in the OSSU Computer Science curriculum.
type Course struct {
	Rank    int    `json:"rank"    csv:"rank"    tsv:"rank"`
	Section string `json:"section" csv:"section" tsv:"section"`
	Title   string `json:"title"   csv:"title"   tsv:"title"`
	URL     string `json:"url"     csv:"url"     tsv:"url"`
}

// Info is site-level stats.
type Info struct {
	Site     string `json:"site"     csv:"site"     tsv:"site"`
	Courses  int    `json:"courses"  csv:"courses"  tsv:"courses"`
	Sections int    `json:"sections" csv:"sections" tsv:"sections"`
	Source   string `json:"source"   csv:"source"   tsv:"source"`
}
