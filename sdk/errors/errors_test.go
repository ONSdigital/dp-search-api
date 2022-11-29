package errors

import (
	"errors"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func Test(t *testing.T) {
	t.Parallel()

	Convey("given a status error", t, func() {
		sErr := StatusError{
			Code: 500,
			Err:  errors.New("test error"),
		}

		Convey("when calling the Error method on status error", func() {
			message := sErr.Error()

			Convey("then the error message is returned", func() {
				So(message, ShouldEqual, "test error")
			})
		})

		Convey("when calling the Status method on status error", func() {
			statusCode := sErr.Status()

			Convey("then the status is returned", func() {
				So(statusCode, ShouldEqual, 500)
			})
		})

		Convey("when passing status error into ErrorStatus func", func() {
			statusCode := ErrorStatus(sErr)

			Convey("then the status is returned", func() {
				So(statusCode, ShouldEqual, 500)
			})
		})

		Convey("when passing status error into ErrorMessage func", func() {
			message := ErrorMessage(sErr)

			Convey("then the error message is returned", func() {
				So(message, ShouldEqual, "test error")
			})
		})
	})
}
