/*
 *  Copyright (c) 2022-2023 Mikhail Knyazhev <markus621@yandex.ru>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package headers

type Header string

const (
	RequestURI  Header = "Request-URI"
	To          Header = "To"
	From        Header = "From"
	CallID      Header = "Call-ID"
	CSeq        Header = "CSeq"
	MaxForwards Header = "Max-Forwards"
	Via         Header = "Via"
	Contact     Header = "Contact"
)
