// Code generated by ffjson <https://github.com/pquerna/ffjson>. DO NOT EDIT.
// source: bitmex_orderbook.go

package orderbook

import (
	"bytes"
	"encoding/json"
	"fmt"
	fflib "github.com/pquerna/ffjson/fflib/v1"
)

// MarshalJSON marshal bytes to json - template
func (j *IBitMexTick) MarshalJSON() ([]byte, error) {
	var buf fflib.Buffer
	if j == nil {
		buf.WriteString("null")
		return buf.Bytes(), nil
	}
	err := j.MarshalJSONBuf(&buf)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// MarshalJSONBuf marshal buff to json - template
func (j *IBitMexTick) MarshalJSONBuf(buf fflib.EncodingBuffer) error {
	if j == nil {
		buf.WriteString("null")
		return nil
	}
	var err error
	var obj []byte
	_ = obj
	_ = err
	buf.WriteString(`{"table":`)
	fflib.WriteJsonString(buf, string(j.Table))
	buf.WriteString(`,"action":`)
	fflib.WriteJsonString(buf, string(j.Action))
	buf.WriteString(`,"data":`)
	if j.Data != nil {
		buf.WriteString(`[`)
		for i, v := range j.Data {
			if i != 0 {
				buf.WriteString(`,`)
			}
			/* Inline struct. type=struct { Symbol string "json:\"symbol\""; ID int64 "json:\"id\""; Side string "json:\"side\""; Size int "json:\"size\""; Price int "json:\"price\"" } kind=struct */
			buf.WriteString(`{ "symbol":`)
			fflib.WriteJsonString(buf, string(v.Symbol))
			buf.WriteString(`,"id":`)
			fflib.FormatBits2(buf, uint64(v.ID), 10, v.ID < 0)
			buf.WriteString(`,"side":`)
			fflib.WriteJsonString(buf, string(v.Side))
			buf.WriteString(`,"size":`)
			fflib.FormatBits2(buf, uint64(v.Size), 10, v.Size < 0)
			buf.WriteString(`,"price":`)
			fflib.FormatBits2(buf, uint64(v.Price), 10, v.Price < 0)
			buf.WriteByte('}')
		}
		buf.WriteString(`]`)
	} else {
		buf.WriteString(`null`)
	}
	buf.WriteByte('}')
	return nil
}

const (
	ffjtIBitMexTickbase = iota
	ffjtIBitMexTicknosuchkey

	ffjtIBitMexTickTable

	ffjtIBitMexTickAction

	ffjtIBitMexTickData
)

var ffjKeyIBitMexTickTable = []byte("table")

var ffjKeyIBitMexTickAction = []byte("action")

var ffjKeyIBitMexTickData = []byte("data")

// UnmarshalJSON umarshall json - template of ffjson
func (j *IBitMexTick) UnmarshalJSON(input []byte) error {
	fs := fflib.NewFFLexer(input)
	return j.UnmarshalJSONFFLexer(fs, fflib.FFParse_map_start)
}

// UnmarshalJSONFFLexer fast json unmarshall - template ffjson
func (j *IBitMexTick) UnmarshalJSONFFLexer(fs *fflib.FFLexer, state fflib.FFParseState) error {
	var err error
	currentKey := ffjtIBitMexTickbase
	_ = currentKey
	tok := fflib.FFTok_init
	wantedTok := fflib.FFTok_init

mainparse:
	for {
		tok = fs.Scan()
		//	println(fmt.Sprintf("debug: tok: %v  state: %v", tok, state))
		if tok == fflib.FFTok_error {
			goto tokerror
		}

		switch state {

		case fflib.FFParse_map_start:
			if tok != fflib.FFTok_left_bracket {
				wantedTok = fflib.FFTok_left_bracket
				goto wrongtokenerror
			}
			state = fflib.FFParse_want_key
			continue

		case fflib.FFParse_after_value:
			if tok == fflib.FFTok_comma {
				state = fflib.FFParse_want_key
			} else if tok == fflib.FFTok_right_bracket {
				goto done
			} else {
				wantedTok = fflib.FFTok_comma
				goto wrongtokenerror
			}

		case fflib.FFParse_want_key:
			// json {} ended. goto exit. woo.
			if tok == fflib.FFTok_right_bracket {
				goto done
			}
			if tok != fflib.FFTok_string {
				wantedTok = fflib.FFTok_string
				goto wrongtokenerror
			}

			kn := fs.Output.Bytes()
			if len(kn) <= 0 {
				// "" case. hrm.
				currentKey = ffjtIBitMexTicknosuchkey
				state = fflib.FFParse_want_colon
				goto mainparse
			} else {
				switch kn[0] {

				case 'a':

					if bytes.Equal(ffjKeyIBitMexTickAction, kn) {
						currentKey = ffjtIBitMexTickAction
						state = fflib.FFParse_want_colon
						goto mainparse
					}

				case 'd':

					if bytes.Equal(ffjKeyIBitMexTickData, kn) {
						currentKey = ffjtIBitMexTickData
						state = fflib.FFParse_want_colon
						goto mainparse
					}

				case 't':

					if bytes.Equal(ffjKeyIBitMexTickTable, kn) {
						currentKey = ffjtIBitMexTickTable
						state = fflib.FFParse_want_colon
						goto mainparse
					}

				}

				if fflib.SimpleLetterEqualFold(ffjKeyIBitMexTickData, kn) {
					currentKey = ffjtIBitMexTickData
					state = fflib.FFParse_want_colon
					goto mainparse
				}

				if fflib.SimpleLetterEqualFold(ffjKeyIBitMexTickAction, kn) {
					currentKey = ffjtIBitMexTickAction
					state = fflib.FFParse_want_colon
					goto mainparse
				}

				if fflib.SimpleLetterEqualFold(ffjKeyIBitMexTickTable, kn) {
					currentKey = ffjtIBitMexTickTable
					state = fflib.FFParse_want_colon
					goto mainparse
				}

				currentKey = ffjtIBitMexTicknosuchkey
				state = fflib.FFParse_want_colon
				goto mainparse
			}

		case fflib.FFParse_want_colon:
			if tok != fflib.FFTok_colon {
				wantedTok = fflib.FFTok_colon
				goto wrongtokenerror
			}
			state = fflib.FFParse_want_value
			continue
		case fflib.FFParse_want_value:

			if tok == fflib.FFTok_left_brace || tok == fflib.FFTok_left_bracket || tok == fflib.FFTok_integer || tok == fflib.FFTok_double || tok == fflib.FFTok_string || tok == fflib.FFTok_bool || tok == fflib.FFTok_null {
				switch currentKey {

				case ffjtIBitMexTickTable:
					goto handle_Table

				case ffjtIBitMexTickAction:
					goto handle_Action

				case ffjtIBitMexTickData:
					goto handle_Data

				case ffjtIBitMexTicknosuchkey:
					err = fs.SkipField(tok)
					if err != nil {
						return fs.WrapErr(err)
					}
					state = fflib.FFParse_after_value
					goto mainparse
				}
			} else {
				goto wantedvalue
			}
		}
	}

handle_Table:

	/* handler: j.Table type=string kind=string quoted=false*/

	{

		{
			if tok != fflib.FFTok_string && tok != fflib.FFTok_null {
				return fs.WrapErr(fmt.Errorf("cannot unmarshal %s into Go value for string", tok))
			}
		}

		if tok == fflib.FFTok_null {

		} else {

			outBuf := fs.Output.Bytes()

			j.Table = string(string(outBuf))

		}
	}

	state = fflib.FFParse_after_value
	goto mainparse

handle_Action:

	/* handler: j.Action type=string kind=string quoted=false*/

	{

		{
			if tok != fflib.FFTok_string && tok != fflib.FFTok_null {
				return fs.WrapErr(fmt.Errorf("cannot unmarshal %s into Go value for string", tok))
			}
		}

		if tok == fflib.FFTok_null {

		} else {

			outBuf := fs.Output.Bytes()

			j.Action = string(string(outBuf))

		}
	}

	state = fflib.FFParse_after_value
	goto mainparse

handle_Data:

	/* handler: j.Data type=[]struct { Symbol string "json:\"symbol\""; ID int64 "json:\"id\""; Side string "json:\"side\""; Size int "json:\"size\""; Price int "json:\"price\"" } kind=slice quoted=false*/

	{
		/* Falling back. type=[]struct { Symbol string "json:\"symbol\""; ID int64 "json:\"id\""; Side string "json:\"side\""; Size int "json:\"size\""; Price int "json:\"price\"" } kind=slice */
		tbuf, err := fs.CaptureField(tok)
		if err != nil {
			return fs.WrapErr(err)
		}

		err = json.Unmarshal(tbuf, &j.Data)
		if err != nil {
			return fs.WrapErr(err)
		}
	}

	state = fflib.FFParse_after_value
	goto mainparse

wantedvalue:
	return fs.WrapErr(fmt.Errorf("wanted value token, but got token: %v", tok))
wrongtokenerror:
	return fs.WrapErr(fmt.Errorf("ffjson: wanted token: %v, but got token: %v output=%s", wantedTok, tok, fs.Output.String()))
tokerror:
	if fs.BigError != nil {
		return fs.WrapErr(fs.BigError)
	}
	err = fs.Error.ToError()
	if err != nil {
		return fs.WrapErr(err)
	}
	panic("ffjson-generated: unreachable, please report bug.")
done:

	return nil
}