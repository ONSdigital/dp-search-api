package errors

import (
	"errors"
	"testing"

	c "github.com/smartystreets/goconvey/convey"
)

func Test(t *testing.T) {
	t.Parallel()

	c.Convey("given a status error", t, func() {
		sErr := StatusError{
			Code: 500,
			Err:  errors.New("test error"),
		}

		c.Convey("when calling the Error method on status error", func() {
			message := sErr.Error()

			c.Convey("then the error message is returned", func() {
				c.So(message, c.ShouldEqual, "test error")
			})
		})

		c.Convey("when calling the Status method on status error", func() {
			statusCode := sErr.Status()

			c.Convey("then the status is returned", func() {
				c.So(statusCode, c.ShouldEqual, 500)
			})
		})

		c.Convey("when passing status error into ErrorStatus func", func() {
			statusCode := ErrorStatus(sErr)

			c.Convey("then the status is returned", func() {
				c.So(statusCode, c.ShouldEqual, 500)
			})
		})

		c.Convey("when passing status error into ErrorMessage func", func() {
			message := ErrorMessage(sErr)

			c.Convey("then the error message is returned", func() {
				c.So(message, c.ShouldEqual, "test error")
			})
		})
	})
}
