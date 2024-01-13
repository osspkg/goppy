/*
 *  Copyright (c) 2022-2024 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package status

type Code uint16

const (
	Trying                       Code = 100
	Ringing                      Code = 180
	CallIsForwarded              Code = 181
	Queued                       Code = 182
	SessionInProgress            Code = 183
	OK                           Code = 200
	MovedPermanently             Code = 301
	MovedTemporarily             Code = 302
	UseProxy                     Code = 305
	BadRequest                   Code = 400
	Unauthorized                 Code = 401
	PaymentRequired              Code = 402
	Forbidden                    Code = 403
	NotFound                     Code = 404
	MethodNotAllowed             Code = 405
	NotAcceptable                Code = 406
	ProxyAuthRequired            Code = 407
	RequestTimeout               Code = 408
	Conflict                     Code = 409
	Gone                         Code = 410
	RequestEntityTooLarge        Code = 413
	RequestURITooLong            Code = 414
	UnsupportedMediaType         Code = 415
	RequestedRangeNotSatisfiable Code = 416
	BadExtension                 Code = 420
	ExtensionRequired            Code = 421
	IntervalToBrief              Code = 423
	TemporarilyUnavailable       Code = 480
	CallTransactionDoesNotExists Code = 481
	LoopDetected                 Code = 482
	TooManyHops                  Code = 483
	AddressIncomplete            Code = 484
	Ambiguous                    Code = 485
	BusyHere                     Code = 486
	RequestTerminated            Code = 487
	NotAcceptableHere            Code = 488
	InternalServerError          Code = 500
	NotImplemented               Code = 501
	BadGateway                   Code = 502
	ServiceUnavailable           Code = 503
	GatewayTimeout               Code = 504
	VersionNotSupported          Code = 505
	MessageTooLarge              Code = 513
	GlobalBusyEverywhere         Code = 600
	GlobalDecline                Code = 603
	GlobalDoesNotExistAnywhere   Code = 604
	GlobalNotAcceptable          Code = 606
)
