/*
 *  Copyright (c) 2022-2024 Mikhail Knyazhev <markus621@yandex.ru>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package method

type Method string

const (
	Invite    Method = "INVITE"
	Ack       Method = "ACK"
	Cancel    Method = "CANCEL"
	Bye       Method = "BYE"
	Register  Method = "REGISTER"
	Options   Method = "OPTIONS"
	Subscribe Method = "SUBSCRIBE"
	Notify    Method = "NOTIFY"
	Refer     Method = "REFER"
	Info      Method = "INFO"
	Message   Method = "MESSAGE"
	Prack     Method = "PRACK"
	Update    Method = "UPDATE"
	Publish   Method = "PUBLISH"
)
