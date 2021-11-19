package query

import (
	"github.com/pkg/errors"
	"testing"
	"text/template"

	. "github.com/smartystreets/goconvey/convey"
)

func MockNewQueryBuilder() (*Builder, error) {
	searchTemplates, err := setupSearchTestTemplates("dummy{{.Moo}}")
	if err != nil {
		return nil, errors.Wrap(err, "failed to load search templates")
	}
	return &Builder{
		searchTemplates: searchTemplates,
	}, nil
}
func setupSearchTestTemplates(rawTemplate string) (*template.Template, error) {
	templates, err := template.ParseFiles(rawTemplate)
	return templates, err
}

func TestMockNewQueryBuilder(t *testing.T) {
	Convey("Should return InternalError for unable to load template", t, func() {

		query, err := MockNewQueryBuilder()

		So(err, ShouldNotBeNil)
		So(query, ShouldBeNil)
		So(err.Error(), ShouldContainSubstring, "failed to load search templates")
	})
}