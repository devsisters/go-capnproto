package capn_test

import (
	"bytes"
	"fmt"
	"testing"

	capn "github.com/devsisters/go-capnproto"
	air "github.com/devsisters/go-capnproto/aircraftlib"
	cv "github.com/glycerine/goconvey/convey"
)

func TestTextAndListTextContaintingEmptyStruct(t *testing.T) {

	emptyZjobBytes := CapnpEncode("()", "Zjob")

	cv.Convey("Given a simple struct message Zjob containing a string and a list of string (all empty)", t, func() {
		cv.Convey("then the go-capnproto serialization should match the capnp c++ serialization", func() {
			ShowBytes(emptyZjobBytes, 10)

			seg := capn.NewBuffer(nil)
			air.NewRootZjob(seg)

			buf := bytes.Buffer{}
			seg.WriteTo(&buf)

			cv.So(buf.Bytes(), cv.ShouldResemble, emptyZjobBytes)
		})
	})
}

func TestTextContaintingStruct(t *testing.T) {

	zjobBytes := CapnpEncode(`(cmd = "abc")`, "Zjob")

	cv.Convey("Given a simple struct message Zjob containing a string 'abc' and a list of string (empty)", t, func() {
		cv.Convey("then the go-capnproto serialization should match the capnp c++ serialization", func() {

			seg := capn.NewBuffer(nil)
			zjob := air.NewRootZjob(seg)
			zjob.SetCmd("abc")

			buf := bytes.Buffer{}
			seg.WriteTo(&buf)

			act := buf.Bytes()
			fmt.Printf("          actual:\n")
			ShowBytes(act, 10)

			fmt.Printf("\n\n          expected:\n")
			ShowBytes(zjobBytes, 10)

			cv.So(act, cv.ShouldResemble, zjobBytes)
		})
	})
}

func TestTextListContaintingStruct(t *testing.T) {

	zjobBytes := CapnpEncode(`(args = ["xyz"])`, "Zjob")

	cv.Convey("Given a simple struct message Zjob containing an unset string and a list of string ('xyz' as the only element)", t, func() {
		cv.Convey("then the go-capnproto serialization should match the capnp c++ serialization", func() {

			seg := capn.NewBuffer(nil)
			zjob := air.NewRootZjob(seg)
			tl := seg.NewTextList(1)
			tl.Set(0, "xyz")
			zjob.SetArgs(tl)

			buf := bytes.Buffer{}
			seg.WriteTo(&buf)

			act := buf.Bytes()
			fmt.Printf("          actual:\n")
			ShowBytes(act, 10)

			fmt.Printf("expected:\n")
			ShowBytes(zjobBytes, 10)

			cv.So(act, cv.ShouldResemble, zjobBytes)
		})
	})
}

func TestTextAndTextListContaintingStruct(t *testing.T) {

	zjobBytes := CapnpEncode(`(cmd = "abc", args = ["xyz"])`, "Zjob")

	cv.Convey("Given a simple struct message Zjob containing a string (cmd='abc') and a list of string (args=['xyz'])", t, func() {
		cv.Convey("then the go-capnproto serialization should match the capnp c++ serialization", func() {

			seg := capn.NewBuffer(nil)
			zjob := air.NewRootZjob(seg)
			zjob.SetCmd("abc")
			tl := seg.NewTextList(1)
			tl.Set(0, "xyz")
			zjob.SetArgs(tl)

			buf := bytes.Buffer{}
			seg.WriteTo(&buf)

			act := buf.Bytes()
			fmt.Printf("          actual:\n")
			ShowBytes(act, 10)

			fmt.Printf("expected:\n")
			ShowBytes(zjobBytes, 10)

			cv.So(act, cv.ShouldResemble, zjobBytes)
		})
	})
}

func TestZserverWithOneFullJob(t *testing.T) {

	exp := CapnpEncode(`(waitingjobs = [(cmd = "abc", args = ["xyz"])])`, "Zserver")

	cv.Convey("Given an Zserver with one empty job", t, func() {
		cv.Convey("then the go-capnproto serialization should match the capnp c++ serialization", func() {

			seg := capn.NewBuffer(nil)
			scratch := capn.NewBuffer(nil)

			server := air.NewRootZserver(seg)

			joblist := air.NewZjobList(seg, 1)
			plist := capn.PointerList(joblist)

			zjob := air.NewZjob(scratch)
			zjob.SetCmd("abc")
			tl := scratch.NewTextList(1)
			tl.Set(0, "xyz")
			zjob.SetArgs(tl)

			plist.Set(0, capn.Object(zjob))

			server.SetWaitingjobs(joblist)

			buf := bytes.Buffer{}
			seg.WriteTo(&buf)

			act := buf.Bytes()
			fmt.Printf("          actual:\n")
			ShowBytes(act, 10)
			fmt.Printf("act decoded by capnp: '%s'\n", string(CapnpDecode(act, "Zserver")))
			save(act, "myact")

			fmt.Printf("expected:\n")
			ShowBytes(exp, 10)
			fmt.Printf("exp decoded by capnp: '%s'\n", string(CapnpDecode(exp, "Zserver")))
			save(exp, "myexp")

			cv.So(act, cv.ShouldResemble, exp)
		})
	})
}

func TestZserverWithAccessors(t *testing.T) {

	exp := CapnpEncode(`(waitingjobs = [(cmd = "abc"), (cmd = "xyz")])`, "Zserver")

	cv.Convey("Given an Zserver with a custom list", t, func() {
		cv.Convey("then all the accessors should work as expected", func() {

			seg := capn.NewBuffer(nil)
			scratch := capn.NewBuffer(nil)

			server := air.NewRootZserver(seg)

			joblist := air.NewZjobList(seg, 2)

			// .Set(int, item)
			zjob := air.NewZjob(scratch)
			zjob.SetCmd("abc")
			joblist.Set(0, zjob)

			zjob = air.NewZjob(scratch)
			zjob.SetCmd("xyz")
			joblist.Set(1, zjob)

			// .At(int)
			cv.So(joblist.At(0).Cmd(), cv.ShouldEqual, "abc")
			cv.So(joblist.At(1).Cmd(), cv.ShouldEqual, "xyz")

			// .Len()
			cv.So(joblist.Len(), cv.ShouldEqual, 2)

			// .ToArray()
			cv.So(len(joblist.ToArray()), cv.ShouldEqual, 2)
			cv.So(joblist.ToArray()[0].Cmd(), cv.ShouldEqual, "abc")
			cv.So(joblist.ToArray()[1].Cmd(), cv.ShouldEqual, "xyz")

			server.SetWaitingjobs(joblist)

			buf := bytes.Buffer{}
			seg.WriteTo(&buf)

			act := buf.Bytes()
			fmt.Printf("          actual:\n")
			ShowBytes(act, 10)
			fmt.Printf("act decoded by capnp: '%s'\n", string(CapnpDecode(act, "Zserver")))
			save(act, "myact")

			fmt.Printf("expected:\n")
			ShowBytes(exp, 10)
			fmt.Printf("exp decoded by capnp: '%s'\n", string(CapnpDecode(exp, "Zserver")))
			save(exp, "myexp")

			cv.So(act, cv.ShouldResemble, exp)
		})
	})
}

func TestEnumFromString(t *testing.T) {
	cv.Convey("Given an enum tag string matching a constant", t, func() {
		cv.Convey("FromString should return the corresponding matching constant value", func() {
			cv.So(air.AirportFromString("jfk"), cv.ShouldEqual, air.AIRPORT_JFK)
		})
	})
	cv.Convey("Given an enum tag string that does not match a constant", t, func() {
		cv.Convey("FromString should return 0", func() {
			cv.So(air.AirportFromString("notEverMatching"), cv.ShouldEqual, 0)
		})
	})
}

func TestSetObjectBetweenSegments(t *testing.T) {

	exp := CapnpEncode(`(counter = (size = 9))`, "Bag")

	cv.Convey("Given an Counter in one segment and a Bag in another", t, func() {
		cv.Convey("we should be able to copy from one segment to the other with SetCounter() on a Bag", func() {

			seg := capn.NewBuffer(nil)
			scratch := capn.NewBuffer(nil)

			// in seg
			segbag := air.NewRootBag(seg)

			// in scratch
			xc := air.NewRootCounter(scratch)
			xc.SetSize(9)

			// copy from scratch to seg
			segbag.SetCounter(xc)

			buf := bytes.Buffer{}
			seg.WriteTo(&buf)

			act := buf.Bytes()
			fmt.Printf("          actual:\n")
			ShowBytes(act, 10)
			fmt.Printf("act decoded by capnp: '%s'\n", string(CapnpDecode(act, "Bag")))
			save(act, "myact")

			fmt.Printf("expected:\n")
			ShowBytes(exp, 10)
			fmt.Printf("exp decoded by capnp: '%s'\n", string(CapnpDecode(exp, "Bag")))
			save(exp, "myexp")

			cv.So(act, cv.ShouldResemble, exp)
		})
	})
}

func TestObjectWithTextBetweenSegments(t *testing.T) {

	exp := CapnpEncode(`(counter = (size = 9, words = "hello"))`, "Bag")

	cv.Convey("Given an Counter in one segment and a Bag with text in another", t, func() {
		cv.Convey("we should be able to copy from one segment to the other with SetCounter() on a Bag", func() {

			seg := capn.NewBuffer(nil)
			scratch := capn.NewBuffer(nil)

			// in seg
			segbag := air.NewRootBag(seg)

			// in scratch
			xc := air.NewRootCounter(scratch)
			xc.SetSize(9)
			xc.SetWords("hello")

			// copy from scratch to seg
			segbag.SetCounter(xc)

			buf := bytes.Buffer{}
			seg.WriteTo(&buf)

			act := buf.Bytes()
			fmt.Printf("          actual:\n")
			ShowBytes(act, 10)
			fmt.Printf("act decoded by capnp: '%s'\n", string(CapnpDecode(act, "Bag")))
			save(act, "myact")

			fmt.Printf("expected:\n")
			ShowBytes(exp, 10)
			fmt.Printf("exp decoded by capnp: '%s'\n", string(CapnpDecode(exp, "Bag")))
			save(exp, "myexp")

			cv.So(act, cv.ShouldResemble, exp)
		})
	})
}

func TestObjectWithListOfTextBetweenSegments(t *testing.T) {

	exp := CapnpEncode(`(counter = (size = 9, wordlist = ["hello","bye"]))`, "Bag")

	cv.Convey("Given an Counter in one segment and a Bag with text in another", t, func() {
		cv.Convey("we should be able to copy from one segment to the other with SetCounter() on a Bag", func() {

			seg := capn.NewBuffer(nil)
			scratch := capn.NewBuffer(nil)

			// in seg
			segbag := air.NewRootBag(seg)

			// in scratch
			xc := air.NewRootCounter(scratch)
			xc.SetSize(9)
			tl := scratch.NewTextList(2)
			tl.Set(0, "hello")
			tl.Set(1, "bye")
			xc.SetWordlist(tl)

			xbuf := bytes.Buffer{}
			scratch.WriteTo(&xbuf)

			x := xbuf.Bytes()
			save(x, "myscratch")
			fmt.Printf("scratch segment (%p):\n", scratch)
			ShowBytes(x, 10)
			fmt.Printf("scratch segment (%p) with Counter decoded by capnp: '%s'\n", scratch, string(CapnpDecode(x, "Counter")))

			prebuf := bytes.Buffer{}
			seg.WriteTo(&prebuf)
			fmt.Printf("Bag only segment seg (%p), pre-transfer:\n", seg)
			ShowBytes(prebuf.Bytes(), 10)

			// now for the actual test:
			// copy from scratch to seg
			segbag.SetCounter(xc)

			buf := bytes.Buffer{}
			seg.WriteTo(&buf)

			act := buf.Bytes()
			save(act, "myact")
			save(exp, "myexp")

			fmt.Printf("expected:\n")
			ShowBytes(exp, 10)
			fmt.Printf("exp decoded by capnp: '%s'\n", string(CapnpDecode(exp, "Bag")))

			fmt.Printf("          actual:\n")
			ShowBytes(act, 10)
			fmt.Printf("act decoded by capnp: '%s'\n", string(CapnpDecode(act, "Bag")))

			cv.So(act, cv.ShouldResemble, exp)
		})
	})
}

func TestSetBetweenSegments(t *testing.T) {

	exp := CapnpEncode(`(counter = (size = 9, words = "abc", wordlist = ["hello","byenow"]))`, "Bag")

	cv.Convey("Given an struct with Text and List(Text) in one segment", t, func() {
		cv.Convey("assigning it to a struct in a different segment should recursively import", func() {

			seg := capn.NewBuffer(nil)
			scratch := capn.NewBuffer(nil)

			// in seg
			segbag := air.NewRootBag(seg)

			// in scratch
			xc := air.NewRootCounter(scratch)
			xc.SetSize(9)
			tl := scratch.NewTextList(2)
			tl.Set(0, "hello")
			tl.Set(1, "byenow")
			xc.SetWordlist(tl)
			xc.SetWords("abc")

			fmt.Printf("\n\n starting copy from scratch to seg \n\n")

			// copy from scratch to seg
			segbag.SetCounter(xc)

			buf := bytes.Buffer{}
			seg.WriteTo(&buf)

			act := buf.Bytes()
			fmt.Printf("          actual:\n")
			ShowBytes(act, 10)
			//fmt.Printf("act decoded by capnp: '%s'\n", string(CapnpDecode(act, "Bag")))
			save(act, "myact")

			fmt.Printf("expected:\n")
			ShowBytes(exp, 10)
			//fmt.Printf("exp decoded by capnp: '%s'\n", string(CapnpDecode(exp, "Bag")))
			save(exp, "myexp")

			cv.So(act, cv.ShouldResemble, exp)
		})
	})
}

func ShowSeg(msg string, seg *capn.Segment) []byte {
	pre := bytes.Buffer{}
	seg.WriteTo(&pre)

	fmt.Printf("%s\n", msg)
	by := pre.Bytes()
	ShowBytes(by, 10)
	return by
}

func TestZserverWithOneEmptyJob(t *testing.T) {

	exp := CapnpEncode(`(waitingjobs = [()])`, "Zserver")

	cv.Convey("Given an Zserver with one empty job", t, func() {
		cv.Convey("then the go-capnproto serialization should match the capnp c++ serialization", func() {

			seg := capn.NewBuffer(nil)
			scratch := capn.NewBuffer(nil)
			server := air.NewRootZserver(seg)

			joblist := air.NewZjobList(seg, 1)
			plist := capn.PointerList(joblist)

			ShowSeg("          pre NewZjob, segment seg is:", seg)

			zjob := air.NewZjob(scratch)
			plist.Set(0, capn.Object(zjob))

			ShowSeg("          pre SetWaitingjobs, segment seg is:", seg)

			fmt.Printf("Then we do the SetWaitingjobs:\n")
			server.SetWaitingjobs(joblist)

			// save
			buf := bytes.Buffer{}
			seg.WriteTo(&buf)
			act := buf.Bytes()
			save(act, "my.act.zserver")

			// show
			ShowSeg("          actual:\n", seg)

			fmt.Printf("act decoded by capnp: '%s'\n", string(CapnpDecode(act, "Zserver")))

			fmt.Printf("expected:\n")
			ShowBytes(exp, 10)
			fmt.Printf("exp decoded by capnp: '%s'\n", string(CapnpDecode(exp, "Zserver")))

			cv.So(act, cv.ShouldResemble, exp)
		})
	})
}

func TestDefaultStructField(t *testing.T) {
	cv.Convey("Given a new root StackingRoot", t, func() {
		cv.Convey("then the aWithDefault field should have a default", func() {
			seg := capn.NewBuffer(nil)
			root := air.NewRootStackingRoot(seg)

			cv.So(root.AWithDefault().Num(), cv.ShouldEqual, 42)
		})
	})
}

func TestDataTextCopyOptimization(t *testing.T) {
	cv.Convey("Given a text list from a different segment", t, func() {
		cv.Convey("Adding it to a different segment shouldn't panic", func() {
			seg := capn.NewBuffer(nil)
			seg2 := capn.NewBuffer(nil)
			root := air.NewRootNester1Capn(seg)

			strsl := seg2.NewTextList(256)
			for i := 0; i < strsl.Len(); i++ {
				strsl.Set(i, "testess")
			}

			root.SetStrs(strsl)
		})
	})
}
