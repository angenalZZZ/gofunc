package f

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestDeathPkgPath(t *testing.T) {
	Convey("Give pkgPath a ptr", t, func() {
		c := &Closer{}
		name, pkgPath := getPkgPath(c)
		So(name, ShouldEqual, "Closer")
		So(pkgPath, ShouldEqual, "github.com/vrecan/death")

	})

	Convey("Give pkgPath a interface", t, func() {
		var closable Closable
		closable = Closer{}
		name, pkgPath := getPkgPath(closable)
		So(name, ShouldEqual, "Closer")
		So(pkgPath, ShouldEqual, "github.com/vrecan/death")
	})

	Convey("Give pkgPath a copy", t, func() {
		c := Closer{}
		name, pkgPath := getPkgPath(c)
		So(name, ShouldEqual, "Closer")
		So(pkgPath, ShouldEqual, "github.com/vrecan/death")
	})
}
