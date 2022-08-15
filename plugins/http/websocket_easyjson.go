// Code generated by easyjson for marshaling/unmarshaling. DO NOT EDIT.

package http

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

func easyjsonC8566e17DecodeGithubComDewepOnlineGoppyPluginsHttp(in *jlexer.Lexer, out *event) {
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
			out.ID = uint(in.Uint())
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
		case "u":
			if data := in.Raw(); in.Ok() {
				in.AddError((out.UID).UnmarshalJSON(data))
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
func easyjsonC8566e17EncodeGithubComDewepOnlineGoppyPluginsHttp(out *jwriter.Writer, in event) {
	out.RawByte('{')
	first := true
	_ = first
	{
		const prefix string = ",\"e\":"
		out.RawString(prefix[1:])
		out.Uint(uint(in.ID))
	}
	{
		const prefix string = ",\"d\":"
		out.RawString(prefix)
		out.Raw((in.Data).MarshalJSON())
	}
	if in.Err != nil {
		const prefix string = ",\"err\":"
		out.RawString(prefix)
		out.String(string(*in.Err))
	}
	if len(in.UID) != 0 {
		const prefix string = ",\"u\":"
		out.RawString(prefix)
		out.Raw((in.UID).MarshalJSON())
	}
	out.RawByte('}')
}

// MarshalJSON supports json.Marshaler interface
func (v event) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	easyjsonC8566e17EncodeGithubComDewepOnlineGoppyPluginsHttp(&w, v)
	return w.Buffer.BuildBytes(), w.Error
}

// MarshalEasyJSON supports easyjson.Marshaler interface
func (v event) MarshalEasyJSON(w *jwriter.Writer) {
	easyjsonC8566e17EncodeGithubComDewepOnlineGoppyPluginsHttp(w, v)
}

// UnmarshalJSON supports json.Unmarshaler interface
func (v *event) UnmarshalJSON(data []byte) error {
	r := jlexer.Lexer{Data: data}
	easyjsonC8566e17DecodeGithubComDewepOnlineGoppyPluginsHttp(&r, v)
	return r.Error()
}

// UnmarshalEasyJSON supports easyjson.Unmarshaler interface
func (v *event) UnmarshalEasyJSON(l *jlexer.Lexer) {
	easyjsonC8566e17DecodeGithubComDewepOnlineGoppyPluginsHttp(l, v)
}
