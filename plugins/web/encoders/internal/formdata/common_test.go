package formdata

import (
	"bytes"
	"io"
	"testing"
	"time"

	"go.osspkg.com/bb"
	"go.osspkg.com/casecheck"
	"go.osspkg.com/cast"
)

type testStruct struct {
	Name    string        `form:"name"`
	Age     int           `form:"age"`
	Ignore  string        // no tag
	Private string        `form:"private"` // unexported
	Empty   *string       `form:"empty,omitempty"`
	Time    time.Time     `form:"time"`
	File    *bytes.Buffer `form:"file" filename:"hello.txt"`
}

func TestUnit_EncodeDecode(t *testing.T) {

	in := testStruct{
		Name:  "ddd",
		File:  bytes.NewBufferString("hello"),
		Empty: cast.Ptr("dddddddd"),
	}
	out := testStruct{File: &bytes.Buffer{}}

	w := bb.New(1024)
	ct, err := NewEncoder().Encode(w, in)
	t.Logf("%q", ct)
	casecheck.NoError(t, err, "Encode")

	w.Seek(0, io.SeekStart)
	t.Log(w.String())

	err = NewDecoder().Decode(w, defaultMaxMemory, &out)
	casecheck.NoError(t, err, "Decode")

	casecheck.Equal(t, in.Name, out.Name)
	casecheck.Equal(t, in.Age, out.Age)
	casecheck.Equal(t, in.Ignore, out.Ignore)
	casecheck.Equal(t, in.Private, out.Private)
	casecheck.Equal(t, in.Time, out.Time)
	casecheck.Equal(t, out.File.Bytes(), out.File.Bytes())
}
