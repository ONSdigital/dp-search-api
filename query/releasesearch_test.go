package query

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"text/template"
	"time"

	c "github.com/smartystreets/goconvey/convey"
)

var validators = NewReleaseQueryParamValidator()

func TestLimit(t *testing.T) {
	t.Parallel()
	c.Convey("given a limit validator, and a set of limits as strings", t, func() {
		validator := validators["limit"]
		limits := []struct {
			given   string
			exValue int
			exError error
		}{
			{given: "XXX", exValue: 0, exError: errors.New("limit search parameter provided with non numeric characters")},
			{given: "-1", exValue: 0, exError: errors.New("limit search parameter provided with negative value")},
			{given: "1001", exValue: 0, exError: fmt.Errorf("limit search parameter provided with a value that is too high")},
			{given: "0", exValue: 0, exError: nil},
			{given: "1000", exValue: 1000, exError: nil},
		}

		for _, ls := range limits {
			v, e := validator(ls.given)

			c.So(v, c.ShouldEqual, ls.exValue)
			c.So(e, c.ShouldResemble, ls.exError)
		}
	})
}

func TestOffset(t *testing.T) {
	t.Parallel()
	c.Convey("given an offset validator, and a set of offsets as strings", t, func() {
		validator := validators["offset"]
		offsets := []struct {
			given   string
			exValue int
			exError error
		}{
			{given: "XXX", exValue: 0, exError: errors.New("offset search parameter provided with non numeric characters")},
			{given: "-1", exValue: 0, exError: errors.New("offset search parameter provided with negative value")},
			{given: "0", exValue: 0, exError: nil},
			{given: "1", exValue: 1, exError: nil},
			{given: "15000", exValue: 15000, exError: nil},
		}

		for _, ls := range offsets {
			v, e := validator(ls.given)

			c.So(v, c.ShouldEqual, ls.exValue)
			c.So(e, c.ShouldResemble, ls.exError)
		}
	})
}

func TestDates(t *testing.T) {
	t.Parallel()
	c.Convey("given a date validator, and a set of erroneous date strings", t, func() {
		validator := validators["date"]
		badDates := []string{"XXX", "Jan 21", "30/12/2021", "2021-13-31", "2021-12-32", "2021-02-29", "2300-12-31"}

		c.Convey("errors are generated, and zero values returned on validation", func() {
			for _, ds := range badDates {
				v, e := validator(ds)

				c.So(v, c.ShouldBeNil)
				c.So(e, c.ShouldNotBeNil)
			}
		})

		c.Convey("but a good date string is validated without error, and the appropriate Date returned", func() {
			date := "2022-12-31"
			v, e := validator(date)

			c.So(v, c.ShouldResemble, MustParseDate(date))
			c.So(e, c.ShouldBeNil)
		})
	})
}

func TestSort(t *testing.T) {
	t.Parallel()
	c.Convey("given a sort validator, and a set of erroneous sort string options", t, func() {
		validator := validators["sort"]
		badSortOptions := []string{"dont sort", "sort-by-date", "date-ascending"}

		c.Convey("errors are generated, and zero values returned on validation", func() {
			for _, ds := range badSortOptions {
				v, e := validator(ds)

				c.So(v, c.ShouldBeNil)
				c.So(e, c.ShouldNotBeNil)
			}
		})

		c.Convey("but a good sort option string is validated without error, and the appropriate Sort option returned", func() {
			goodSortOptions := []struct {
				given   string
				exValue Sort
			}{
				{given: "release_date_asc", exValue: RelDateAsc},
				{given: "release_date_desc", exValue: RelDateDesc},
				{given: "title_asc", exValue: TitleAsc},
				{given: "title_desc", exValue: TitleDesc},
			}

			for _, gso := range goodSortOptions {
				v, e := validator(gso.given)

				c.So(v, c.ShouldEqual, gso.exValue)
				c.So(e, c.ShouldBeNil)
			}
		})
	})
}

func TestReleaseType(t *testing.T) {
	t.Parallel()
	c.Convey("given a release-type validator, and a set of erroneous release-type option strings", t, func() {
		validator := validators["release-type"]
		badReleaseTypes := []string{"coming-up", "finished", "done"}

		c.Convey("errors are generated, and zero values returned on validation", func() {
			for _, rt := range badReleaseTypes {
				v, e := validator(rt)

				c.So(v, c.ShouldBeNil)
				c.So(e, c.ShouldNotBeNil)
			}
		})

		c.Convey("but a good release-type option string is validated without error, and the appropriate ReleaseType returned", func() {
			goodReleaseTypes := []struct {
				given   string
				exValue ReleaseType
			}{
				{given: "type-upcoming", exValue: Upcoming},
				{given: "type-published", exValue: Published},
				{given: "type-cancelled", exValue: Cancelled},
			}

			for _, grt := range goodReleaseTypes {
				v, e := validator(grt.given)

				c.So(v, c.ShouldEqual, grt.exValue)
				c.So(e, c.ShouldBeNil)
			}
		})
	})
}

func TestParseQuery(t *testing.T) {
	t.Parallel()
	c.Convey("Given a query string with no template prefix", t, func() {
		query := `A "standard query"`

		c.Convey("the function return the query string exactly with the standard template name", func() {
			q, tmpl := ParseQuery(query)
			c.So(q, c.ShouldEqual, `A \"standard query\"`)
			c.So(tmpl, c.ShouldEqual, templateNames[standard])
		})
	})

	c.Convey("Given a query string with the prefix for the simple template", t, func() {
		query := `!!s:A "simple query"`

		c.Convey("the function returns the query string without the template prefix, and the simple template name", func() {
			q, tmpl := ParseQuery(query)
			c.So(q, c.ShouldEqual, `A \"simple query\"`)
			c.So(tmpl, c.ShouldEqual, templateNames[simple])
		})
	})

	c.Convey("Given a query string with the prefix for the sitewide template", t, func() {
		query := `!!sw:A simple "sitewide query"`

		c.Convey("the function returns the query string without the template prefix, and the sitewide template name", func() {
			q, tmpl := ParseQuery(query)
			c.So(q, c.ShouldEqual, `A simple \"sitewide query\"`)
			c.So(tmpl, c.ShouldEqual, templateNames[sitewide])
		})
	})
}

func TestBuildSearchReleaseQuery(t *testing.T) {
	t.Parallel()
	c.Convey("Should return InternalError for invalid template", t, func() {
		qb := createReleaseQueryBuilderForTemplate("dummy{{.Moo}}")
		query, err := qb.BuildSearchQuery(context.Background(), ReleaseSearchRequest{})

		c.So(err, c.ShouldNotBeNil)
		c.So(query, c.ShouldBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, "creation of search from template failed")
	})

	c.Convey("Should include all search parameters in elastic search query", t, func() {
		qb := createReleaseQueryBuilderForTemplate("test-index\n" +
			Term +
			From +
			Size +
			"SortBy={{.SortBy.ESString}};" +
			"ReleasedAfter={{.ReleasedAfter.ESString}};" +
			"ReleasedBefore={{.ReleasedBefore.ESString}};" +
			"Type={{.Type.String}};" +
			Highlight +
			Now)

		query, err := qb.BuildSearchQuery(context.Background(), ReleaseSearchRequest{
			Term:           "query+term",
			From:           0,
			Size:           25,
			SortBy:         TitleAsc,
			ReleasedAfter:  Date{},
			ReleasedBefore: MustParseDate("2020-12-31"),
			Type:           Published,
			Highlight:      true,
		})

		c.So(err, c.ShouldBeNil)
		c.So(len(query), c.ShouldEqual, 1)

		queryString := string(query[0].Query)
		c.So(queryString, c.ShouldContainSubstring, "Term=query+term")
		c.So(queryString, c.ShouldContainSubstring, "From=0")
		c.So(queryString, c.ShouldContainSubstring, "Size=25")
		c.So(queryString, c.ShouldContainSubstring, `SortBy={"title.title_raw":"asc"}`)
		c.So(queryString, c.ShouldContainSubstring, "ReleasedAfter=null")
		c.So(queryString, c.ShouldContainSubstring, `ReleasedBefore="2020-12-31"`)
		c.So(queryString, c.ShouldContainSubstring, "Type=type-published")
		c.So(queryString, c.ShouldContainSubstring, "Highlight=true")
		c.So(queryString, c.ShouldContainSubstring, fmt.Sprintf(`Now=%q`, time.Now().Format(dateFormat)))
	})
}

func createReleaseQueryBuilderForTemplate(rawTemplate string) *ReleaseBuilder {
	temp, err := template.New("search.tmpl").Parse(rawTemplate)
	c.So(err, c.ShouldBeNil)
	return &ReleaseBuilder{
		searchTemplates: temp,
	}
}
