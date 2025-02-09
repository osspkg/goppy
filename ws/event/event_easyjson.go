// Code generated by easyjson for marshaling/unmarshaling. DO NOT EDIT.

package event

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

func easyjsonF642ad3eDecodeGoOsspkgComGoppyV2WsEvent(in *jlexer.Lexer, out *event) {
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
		case "e":
			out.Id = Id(in.Uint16())
		case "d":
			if data := in.Raw(); in.Ok() {
				in.AddError((out.Data).UnmarshalJSON(data))
			}
		case "err":
			if in.IsNull() {
				in.Skip()
				out.Err = nil
			} else {
				if out.Err == nil {
					out.Err = new(string)
				}
				*out.Err = string(in.String())
			}
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
func easyjsonF642ad3eEncodeGoOsspkgComGoppyV2WsEvent(out *jwriter.Writer, in event) {
	out.RawByte('{')
	first := true
	_ = first
	{
		const prefix string = ",\"e\":"
		out.RawString(prefix[1:])
		out.Uint16(uint16(in.Id))
	}
	if len(in.Data) != 0 {
		const prefix string = ",\"d\":"
		out.RawString(prefix)
		out.Raw((in.Data).MarshalJSON())
	}
	if in.Err != nil {
		const prefix string = ",\"err\":"
		out.RawString(prefix)
		out.String(string(*in.Err))
	}
	out.RawByte('}')
}

// MarshalJSON supports json.Marshaler interface
func (v event) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	easyjsonF642ad3eEncodeGoOsspkgComGoppyV2WsEvent(&w, v)
	return w.Buffer.BuildBytes(), w.Error
}

// MarshalEasyJSON supports easyjson.Marshaler interface
func (v event) MarshalEasyJSON(w *jwriter.Writer) {
	easyjsonF642ad3eEncodeGoOsspkgComGoppyV2WsEvent(w, v)
}

// UnmarshalJSON supports json.Unmarshaler interface
func (v *event) UnmarshalJSON(data []byte) error {
	r := jlexer.Lexer{Data: data}
	easyjsonF642ad3eDecodeGoOsspkgComGoppyV2WsEvent(&r, v)
	return r.Error()
}

// UnmarshalEasyJSON supports easyjson.Unmarshaler interface
func (v *event) UnmarshalEasyJSON(l *jlexer.Lexer) {
	easyjsonF642ad3eDecodeGoOsspkgComGoppyV2WsEvent(l, v)
}
