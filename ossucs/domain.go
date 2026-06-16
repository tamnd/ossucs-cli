package ossucs

import (
	"context"
	"fmt"
	"strconv"

	"github.com/tamnd/any-cli/kit"
	"github.com/tamnd/any-cli/kit/errs"
)

// domain.go exposes ossucs as a kit Domain: a driver that a multi-domain
// host (ant) enables with a single blank import,
//
//	import _ "github.com/tamnd/ossucs-cli/ossucs"
//
// exactly as a database/sql program enables a driver with `import _
// "github.com/lib/pq"`. The init below registers it; the host then dereferences
// ossucs:// URIs by routing to the operations Register installs. The same
// Domain also builds the standalone ossucs binary (see cli.NewApp), so the
// binary and a host share one source of truth.
func init() { kit.Register(Domain{}) }

// Domain is the ossucs driver. It carries no state; the per-run client is
// built by the factory Register hands kit.
type Domain struct{}

// Info describes the scheme, the hostnames a pasted link is matched against, and
// the identity reused for the binary's help and version.
func (Domain) Info() kit.DomainInfo {
	return kit.DomainInfo{
		Scheme: "ossucs",
		Hosts:  []string{"github.com/ossu/computer-science"},
		Identity: kit.Identity{
			Binary: "ossucs",
			Short:  "Browse the OSSU Computer Science curriculum from the command line",
			Long: `Browse the Open Source Society University Computer Science curriculum
from the command line.

ossucs reads the OSSU CS curriculum over plain HTTPS, shapes it into
clean records, and prints output that pipes into the rest of your tools. No API
key, nothing to run alongside it.`,
			Site: "github.com/ossu/computer-science",
			Repo: "https://github.com/tamnd/ossucs-cli",
		},
	}
}

// Register installs the client factory and every operation onto app.
func (Domain) Register(app *kit.App) {
	app.SetClient(newClient)

	// list: enumerate all courses (alias for courses).
	kit.Handle(app, kit.OpMeta{Name: "list", Group: "read", List: true,
		Summary: "List courses (optionally filtered by section)",
		URIType: "course"}, listCoursesFiltered)

	// courses: enumerate all courses.
	kit.Handle(app, kit.OpMeta{Name: "courses", Group: "read", List: true,
		Summary: "List all courses in the OSSU Computer Science curriculum",
		URIType: "course"}, listCourses)

	// course: one course by 1-based index.
	kit.Handle(app, kit.OpMeta{Name: "course", Group: "read", Single: true,
		Summary: "Show one course by index",
		Args:    []kit.Arg{{Name: "index", Help: "1-based course index"}}}, getCourse)

	// info: site-level stats.
	kit.Handle(app, kit.OpMeta{Name: "info", Group: "read", Single: true,
		Summary: "Print site stats (total courses, sections)"}, getInfo)
}

// newClient builds the client from the host-resolved config, so a host and the
// standalone binary pace and identify themselves the same way.
func newClient(_ context.Context, cfg kit.Config) (any, error) {
	dc := DefaultConfig()
	if cfg.UserAgent != "" {
		dc.UserAgent = cfg.UserAgent
	}
	if cfg.Rate > 0 {
		dc.Rate = cfg.Rate
	}
	if cfg.Retries > 0 {
		dc.Retries = cfg.Retries
	}
	if cfg.Timeout > 0 {
		dc.Timeout = cfg.Timeout
	}
	return NewClient(dc), nil
}

// --- inputs ---

type coursesInput struct {
	Limit  int     `kit:"flag,inherit" help:"max results"`
	Client *Client `kit:"inject"`
}

type coursesFilteredInput struct {
	Section string  `kit:"flag" help:"filter by section name"`
	Limit   int     `kit:"flag,inherit" help:"max results"`
	Client  *Client `kit:"inject"`
}

type courseInput struct {
	Index  string  `kit:"arg" help:"1-based course index"`
	Client *Client `kit:"inject"`
}

type infoInput struct {
	Client *Client `kit:"inject"`
}

// --- handlers ---

func listCourses(ctx context.Context, in coursesInput, emit func(*Course) error) error {
	courses, err := in.Client.Courses(ctx)
	if err != nil {
		return mapErr(err)
	}
	for i, c := range courses {
		if in.Limit > 0 && i >= in.Limit {
			break
		}
		if err := emit(c); err != nil {
			return err
		}
	}
	return nil
}

func listCoursesFiltered(ctx context.Context, in coursesFilteredInput, emit func(*Course) error) error {
	var courses []*Course
	var err error
	if in.Section != "" {
		courses, err = in.Client.CoursesBySection(ctx, in.Section)
	} else {
		courses, err = in.Client.Courses(ctx)
	}
	if err != nil {
		return mapErr(err)
	}
	for i, c := range courses {
		if in.Limit > 0 && i >= in.Limit {
			break
		}
		if err := emit(c); err != nil {
			return err
		}
	}
	return nil
}

func getCourse(ctx context.Context, in courseInput, emit func(*Course) error) error {
	idx, err := strconv.Atoi(in.Index)
	if err != nil {
		return errs.Usage("index must be an integer: %s", err)
	}
	c, err := in.Client.CourseByIndex(ctx, idx)
	if err != nil {
		return fmt.Errorf("%w", err)
	}
	return emit(c)
}

func getInfo(ctx context.Context, in infoInput, emit func(*Info) error) error {
	info, err := in.Client.Info(ctx)
	if err != nil {
		return mapErr(err)
	}
	return emit(info)
}

// --- Resolver ---

// Classify turns a reference into (type, id). Courses are addressed by rank.
func (Domain) Classify(input string) (uriType, id string, err error) {
	if input == "" {
		return "", "", errs.Usage("unrecognized ossucs reference: %q", input)
	}
	return "course", input, nil
}

// Locate is the inverse: the live https URL for a (type, id).
func (Domain) Locate(uriType, id string) (string, error) {
	if uriType != "course" {
		return "", errs.Usage("ossucs has no resource type %q", uriType)
	}
	return "https://github.com/ossu/computer-science", nil
}

// mapErr converts a library error into the kit error kind that carries the right exit code.
func mapErr(err error) error {
	return err
}
