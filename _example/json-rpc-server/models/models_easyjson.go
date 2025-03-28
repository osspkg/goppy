// Code generated by easyjson for marshaling/unmarshaling. DO NOT EDIT.

package models

import (
	json "encoding/json"

	easyjson "github.com/mailru/easyjson"
	jlexer "github.com/mailru/easyjson/jlexer"
	jwriter "github.com/mailru/easyjson/jwriter"
)

// suppress unused package warning
var (
	_ *json.RawMessage
	_ *jlexer.Lexer
	_ *jwriter.Writer
	_ easyjson.Marshaler
)

func easyjsonD2b7633eDecodeGoOsspkgComGoppyV2ExampleJsonRpcServerModels(in *jlexer.Lexer, out *Users) {
	isTopLevel := in.IsStart()
	if in.IsNull() {
		in.Skip()
		*out = nil
	} else {
		in.Delim('[')
		if *out == nil {
			if !in.IsDelim(']') {
				*out = make(Users, 0, 2)
			} else {
				*out = Users{}
			}
		} else {
			*out = (*out)[:0]
		}
		for !in.IsDelim(']') {
			var v1 User
			(v1).UnmarshalEasyJSON(in)
			*out = append(*out, v1)
			in.WantComma()
		}
		in.Delim(']')
	}
	if isTopLevel {
		in.Consumed()
	}
}
func easyjsonD2b7633eEncodeGoOsspkgComGoppyV2ExampleJsonRpcServerModels(out *jwriter.Writer, in Users) {
	if in == nil && (out.Flags&jwriter.NilSliceAsEmpty) == 0 {
		out.RawString("null")
	} else {
		out.RawByte('[')
		for v2, v3 := range in {
			if v2 > 0 {
				out.RawByte(',')
			}
			(v3).MarshalEasyJSON(out)
		}
		out.RawByte(']')
	}
}

// MarshalJSON supports json.Marshaler interface
func (v Users) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	easyjsonD2b7633eEncodeGoOsspkgComGoppyV2ExampleJsonRpcServerModels(&w, v)
	return w.Buffer.BuildBytes(), w.Error
}

// MarshalEasyJSON supports easyjson.Marshaler interface
func (v Users) MarshalEasyJSON(w *jwriter.Writer) {
	easyjsonD2b7633eEncodeGoOsspkgComGoppyV2ExampleJsonRpcServerModels(w, v)
}

// UnmarshalJSON supports json.Unmarshaler interface
func (v *Users) UnmarshalJSON(data []byte) error {
	r := jlexer.Lexer{Data: data}
	easyjsonD2b7633eDecodeGoOsspkgComGoppyV2ExampleJsonRpcServerModels(&r, v)
	return r.Error()
}

// UnmarshalEasyJSON supports easyjson.Unmarshaler interface
func (v *Users) UnmarshalEasyJSON(l *jlexer.Lexer) {
	easyjsonD2b7633eDecodeGoOsspkgComGoppyV2ExampleJsonRpcServerModels(l, v)
}
func easyjsonD2b7633eDecodeGoOsspkgComGoppyV2ExampleJsonRpcServerModels1(in *jlexer.Lexer, out *User) {
	isTopLevel := in.IsStart()
	if in.IsNull() {
		if isTopLevel {
			in.Consumed()
		}
		in.Skip()
		return
	}
	in.Delim('{')
	for !in.IsDelim('}') {
		key := in.UnsafeFieldName(false)
		in.WantColon()
		if in.IsNull() {
			in.Skip()
			in.WantComma()
			continue
		}
		switch key {
		case "id":
			out.Id = int(in.Int())
		case "name":
			out.Name = string(in.String())
		default:
			in.SkipRecursive()
		}
		in.WantComma()
	}
	in.Delim('}')
	if isTopLevel {
		in.Consumed()
	}
}
func easyjsonD2b7633eEncodeGoOsspkgComGoppyV2ExampleJsonRpcServerModels1(out *jwriter.Writer, in User) {
	out.RawByte('{')
	first := true
	_ = first
	{
		const prefix string = ",\"id\":"
		out.RawString(prefix[1:])
		out.Int(int(in.Id))
	}
	{
		const prefix string = ",\"name\":"
		out.RawString(prefix)
		out.String(string(in.Name))
	}
	out.RawByte('}')
}

// MarshalJSON supports json.Marshaler interface
func (v User) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	easyjsonD2b7633eEncodeGoOsspkgComGoppyV2ExampleJsonRpcServerModels1(&w, v)
	return w.Buffer.BuildBytes(), w.Error
}

// MarshalEasyJSON supports easyjson.Marshaler interface
func (v User) MarshalEasyJSON(w *jwriter.Writer) {
	easyjsonD2b7633eEncodeGoOsspkgComGoppyV2ExampleJsonRpcServerModels1(w, v)
}

// UnmarshalJSON supports json.Unmarshaler interface
func (v *User) UnmarshalJSON(data []byte) error {
	r := jlexer.Lexer{Data: data}
	easyjsonD2b7633eDecodeGoOsspkgComGoppyV2ExampleJsonRpcServerModels1(&r, v)
	return r.Error()
}

// UnmarshalEasyJSON supports easyjson.Unmarshaler interface
func (v *User) UnmarshalEasyJSON(l *jlexer.Lexer) {
	easyjsonD2b7633eDecodeGoOsspkgComGoppyV2ExampleJsonRpcServerModels1(l, v)
}
func easyjsonD2b7633eDecodeGoOsspkgComGoppyV2ExampleJsonRpcServerModels2(in *jlexer.Lexer, out *IntArray) {
	isTopLevel := in.IsStart()
	if in.IsNull() {
		in.Skip()
		*out = nil
	} else {
		in.Delim('[')
		if *out == nil {
			if !in.IsDelim(']') {
				*out = make(IntArray, 0, 8)
			} else {
				*out = IntArray{}
			}
		} else {
			*out = (*out)[:0]
		}
		for !in.IsDelim(']') {
			var v4 int
			v4 = int(in.Int())
			*out = append(*out, v4)
			in.WantComma()
		}
		in.Delim(']')
	}
	if isTopLevel {
		in.Consumed()
	}
}
func easyjsonD2b7633eEncodeGoOsspkgComGoppyV2ExampleJsonRpcServerModels2(out *jwriter.Writer, in IntArray) {
	if in == nil && (out.Flags&jwriter.NilSliceAsEmpty) == 0 {
		out.RawString("null")
	} else {
		out.RawByte('[')
		for v5, v6 := range in {
			if v5 > 0 {
				out.RawByte(',')
			}
			out.Int(int(v6))
		}
		out.RawByte(']')
	}
}

// MarshalJSON supports json.Marshaler interface
func (v IntArray) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	easyjsonD2b7633eEncodeGoOsspkgComGoppyV2ExampleJsonRpcServerModels2(&w, v)
	return w.Buffer.BuildBytes(), w.Error
}

// MarshalEasyJSON supports easyjson.Marshaler interface
func (v IntArray) MarshalEasyJSON(w *jwriter.Writer) {
	easyjsonD2b7633eEncodeGoOsspkgComGoppyV2ExampleJsonRpcServerModels2(w, v)
}

// UnmarshalJSON supports json.Unmarshaler interface
func (v *IntArray) UnmarshalJSON(data []byte) error {
	r := jlexer.Lexer{Data: data}
	easyjsonD2b7633eDecodeGoOsspkgComGoppyV2ExampleJsonRpcServerModels2(&r, v)
	return r.Error()
}

// UnmarshalEasyJSON supports easyjson.Unmarshaler interface
func (v *IntArray) UnmarshalEasyJSON(l *jlexer.Lexer) {
	easyjsonD2b7633eDecodeGoOsspkgComGoppyV2ExampleJsonRpcServerModels2(l, v)
}
