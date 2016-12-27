/**
 * File        : errors.go
 * Copyright   : Copyright (c) 2015-2017 Mirror Labs, Inc. All rights reserved.
 * License     : GPLv3
 * Maintainer  : Enzo Haussecker <enzo@mirror.co>, Dominic Williams <dominic@string.technology>
 * Stability   : Experimental
 * Portability : Non-portable (requires Tox core at commit dcf2aaa)
 */

package tox

import "errors"

////////////////////////////////////////////////////////////////////////////////
//////////////////////////////////// ERRORS ////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

// A collection of errors to indicate that a specific C-side error was received.
var (

    ToxErrOptionsNewMalloc                     = errors.New("The function failed to allocate enough memory for the options struct.")
    ToxErrNewNull                              = errors.New("One of the arguments to the function was NULL when it was not expected.")
    ToxErrNewMalloc                            = errors.New("The function was unable to allocate enough memory to store the internal structures for the Tox object.")
    ToxErrNewPortAlloc                         = errors.New("The function was unable to bind to a port. This may mean that all ports have already been bound, e.g. by other Tox instances, or it may mean a permission error. You may be able to gather more information from errno.")
    ToxErrNewProxyBadType                      = errors.New("proxy_type was invalid.")
    ToxErrNewProxyBadHost                      = errors.New("proxy_type was valid, but the proxy_host passed had an invalid format or was NULL.")
    ToxErrNewProxyBadPort                      = errors.New("proxy_type was valid, but the proxy_port was invalid.")
    ToxErrNewProxyNotFound                     = errors.New("The proxy address passed could not be resolved.")
    ToxErrNewLoadEncrypted                     = errors.New("The byte array to be loaded contained an encrypted save.")
    ToxErrNewLoadBadFormat                     = errors.New("The data format was invalid. This can happen when loading data that was saved by an older version of Tox, or when the data has been corrupted. When loading from badly formatted data, some data may have been loaded, and the rest is discarded. Passing an invalid length parameter also causes this error.")
    ToxErrBootstrapNull                        = errors.New("One of the arguments to the function was NULL when it was not expected.")
    ToxErrBootstrapBadHost                     = errors.New("The address could not be resolved to an IP address, or the IP address passed was invalid.")
    ToxErrBootstrapBadPort                     = errors.New("The port passed was invalid. The valid port range is (1, 65535).")
    ToxErrSetInfoNull                          = errors.New("One of the arguments to the function was NULL when it was not expected.")
    ToxErrSetInfoTooLong                       = errors.New("Information length exceeded maximum permissible size.")
    ToxErrFriendAddNull                        = errors.New("One of the arguments to the function was NULL when it was not expected.")
    ToxErrFriendAddTooLong                     = errors.New("The length of the friend request message exceeded TOX_MAX_FRIEND_REQUEST_LENGTH.")
    ToxErrFriendAddNoMessage                   = errors.New("The friend request message was empty. This, and the TOO_LONG code will never be returned from tox_friend_add_norequest.")
    ToxErrFriendAddOwnKey                      = errors.New("The friend address belongs to the sending client.")
    ToxErrFriendAddAlreadySent                 = errors.New("A friend request has already been sent, or the address belongs to a friend that is already on the friend list.")
    ToxErrFriendAddBadChecksum                 = errors.New("The friend address checksum failed.")
    ToxErrFriendAddSetNewNoSpam                = errors.New("The friend was already there, but the nospam value was different.")
    ToxErrFriendAddMalloc                      = errors.New("A memory allocation failed when trying to increase the friend list size.")
    ToxErrFriendDeleteFriendNotFound           = errors.New("There was no friend with the given friend number. No friends were deleted.")
    ToxErrFriendByPublicKeyNull                = errors.New("One of the arguments to the function was NULL when it was not expected.")
    ToxErrFriendByPublicKeyNotFound            = errors.New("No friend with the given public key exists on the friend list.")
    ToxErrFriendGetPublicKeyFriendNotFound     = errors.New("No friend with the given number exists on the friend list.")
    ToxErrFriendGetLastOnlineFriendNotFound    = errors.New("No friend with the given number exists on the friend list.")
    ToxErrFriendQueryNull                      = errors.New("The pointer parameter for storing the query result (name, message) was NULL. Unlike the _self_ variants of these functions, which have no effect when a parameter is NULL, these functions return an error in that case.")
    ToxErrFriendQueryFriendNotFound            = errors.New("The friend number did not designate a valid friend.")
    ToxErrSetTypingFriendNotFound              = errors.New("The friend number did not designate a valid friend.")
    ToxErrFriendSendMessageNull                = errors.New("One of the arguments to the function was NULL when it was not expected.")
    ToxErrFriendSendMessageFriendNotFound      = errors.New("The friend number did not designate a valid friend.")
    ToxErrFriendSendMessageFriendNotConnected  = errors.New("This client is currently not connected to the friend.")
    ToxErrFriendSendMessageSendQ               = errors.New("An allocation error occurred while increasing the send queue size.")
    ToxErrFriendSendMessageTooLong             = errors.New("Message length exceeded TOX_MAX_MESSAGE_LENGTH.")
    ToxErrFriendSendMessageEmpty               = errors.New("Attempted to send a zero-length message.")
    ToxErrFriendCustomPacketNull               = errors.New("One of the arguments to the function was NULL when it was not expected.")
    ToxErrFriendCustomPacketFriendNotFound     = errors.New("The friend number did not designate a valid friend.")
    ToxErrFriendCustomPacketFriendNotConnected = errors.New("This client is currently not connected to the friend.")
    ToxErrFriendCustomPacketInvalid            = errors.New("The first byte of data was not in the specified range for the packet type. This range is 200-254 for lossy, and 160-191 for lossless packets.")
    ToxErrFriendCustomPacketEmpty              = errors.New("Attempted to send an empty packet.")
    ToxErrFriendCustomPacketTooLong            = errors.New("Packet data length exceeded TOX_MAX_CUSTOM_PACKET_SIZE.")
    ToxErrFriendCustomPacketSendQ              = errors.New("Packet queue is full.")

)

// An error to indicate that an unrecognized C-side error was received. This is
// usually a result of a version mismatch. Recall that this wrapper is pegged to
// commit dcf2aaa.
var (

    ToxErrUnknown                              = errors.New("Unknown error returned by Tox core")

)
