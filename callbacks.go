/**
 * File        : callbacks.go
 * Copyright   : Copyright (c) 2015-2017 Mirror Labs, Inc. All rights reserved.
 * License     : GPLv3
 * Maintainer  : Enzo Haussecker <enzo@mirror.co>, Dominic Williams <dominic@string.technology>
 * Stability   : Experimental
 * Portability : Non-portable (requires Tox core at commit dcf2aaa)
 *
 * Tox instances handle events using callback functions. Only one callback can
 * be registered per event, so if a client needs multiple event listeners, then
 * it needs to implement the dispatch functionality itself. This module only
 * provides the hooks for registering them.
 */

package tox

//#include <memory.h>
//#include <tox/tox.h>
import "C"
import "unsafe"

////////////////////////////////////////////////////////////////////////////////
//////////////////////////////// CALLBACK HOOKS ////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

//export callback_self_connection_status
func callback_self_connection_status(
    c_tox_ptr *C.Tox,
    c_connection_status C.TOX_CONNECTION,
    c_user_data unsafe.Pointer,
) {
    tox := (*Tox)(c_user_data)
    var connectionStatus ToxConnectionStatus
    switch c_connection_status {
        case C.TOX_CONNECTION_NONE:
            connectionStatus = ToxConnectionNone
        case C.TOX_CONNECTION_TCP:
            connectionStatus = ToxConnectionTCP
        case C.TOX_CONNECTION_UDP:
            connectionStatus = ToxConnectionUDP
        default:
            panic("unknown connection status")
    }
    tox.onSelfConnectionStatus(tox, connectionStatus)
}

//export callback_friend_name
func callback_friend_name(
    c_tox *C.Tox,
    c_friend_number C.uint32_t,
    c_name *C.uint8_t,
    c_length C.size_t,
    c_user_data unsafe.Pointer,
) {
    tox := (*Tox)(c_user_data)
    friendNumber := uint32(c_friend_number)
    name := make([]byte, c_length)
    if (c_length > 0) {
        C.memcpy(
            unsafe.Pointer(&name[0]),
            unsafe.Pointer(c_name),
            c_length,
        )
    }
    tox.onFriendName(tox, friendNumber, name)
}

//export callback_friend_request
func callback_friend_request(
    c_tox *C.Tox,
    c_public_key *C.uint8_t,
    c_message *C.uint8_t,
    c_length C.size_t,
    c_user_data unsafe.Pointer,
) {
    tox := (*Tox)(c_user_data)
    var publicKey ToxPublicKey
    message := make([]byte, c_length)
    C.memcpy(
        unsafe.Pointer(&publicKey[0]),
        unsafe.Pointer(c_public_key),
        ToxPublicKeySize,
    )
    if (c_length > 0) {
        C.memcpy(
            unsafe.Pointer(&message[0]),
            unsafe.Pointer(c_message),
            c_length,
        )
    }
    tox.onFriendRequest(tox, publicKey, message)
}

//export callback_friend_status_message
func callback_friend_status_message(
    c_tox *C.Tox,
    c_friend_number C.uint32_t,
    c_message *C.uint8_t,
    c_length C.size_t,
    c_user_data unsafe.Pointer,
) {
    tox := (*Tox)(c_user_data)
    friendNumber := uint32(c_friend_number)
    message := make([]byte, c_length)
    if (c_length > 0) {
        C.memcpy(
            unsafe.Pointer(&message[0]),
            unsafe.Pointer(c_message),
            c_length,
        )
    }
    tox.onFriendStatusMessage(tox, friendNumber, message)
}

//export callback_friend_status
func callback_friend_status(
    c_tox *C.Tox,
    c_friend_number C.uint32_t,
    c_user_status C.TOX_USER_STATUS,
    c_user_data unsafe.Pointer,
) {
    tox := (*Tox)(c_user_data)
    friendNumber := uint32(c_friend_number)
    var userStatus ToxUserStatus
    switch c_user_status {
        case C.TOX_USER_STATUS_NONE:
            userStatus = ToxUserStatusNone
        case C.TOX_USER_STATUS_AWAY:
            userStatus = ToxUserStatusAway
        case C.TOX_USER_STATUS_BUSY:
            userStatus = ToxUserStatusBusy
        default:
            panic("unknown user status")
    }
    tox.onFriendStatus(tox, friendNumber, userStatus)
}

//export callback_friend_connection_status
func callback_friend_connection_status(
    c_tox *C.Tox,
    c_friend_number C.uint32_t,
    c_connection_status C.TOX_CONNECTION,
    c_user_data unsafe.Pointer,
) {
    tox := (*Tox)(c_user_data)
    friendNumber := uint32(c_friend_number)
    var connectionStatus ToxConnectionStatus
    switch c_connection_status {
        case C.TOX_CONNECTION_NONE:
            connectionStatus = ToxConnectionNone
        case C.TOX_CONNECTION_TCP:
            connectionStatus = ToxConnectionTCP
        case C.TOX_CONNECTION_UDP:
            connectionStatus = ToxConnectionUDP
        default:
            panic("unknown connection status")
    }
    tox.onFriendConnectionStatus(tox, friendNumber, connectionStatus)
}

//export callback_friend_message
func callback_friend_message(
    c_tox *C.Tox,
    c_friend_number C.uint32_t,
    c_message_type C.TOX_MESSAGE_TYPE,
    c_message *C.uint8_t,
    c_length C.size_t,
    c_user_data unsafe.Pointer,
) {
    tox := (*Tox)(c_user_data)
    friendNumber := uint32(c_friend_number)
    var messageType ToxMessageType
    switch c_message_type {
        case C.TOX_MESSAGE_TYPE_NORMAL:
            messageType = ToxMessageTypeNormal
        case C.TOX_MESSAGE_TYPE_ACTION:
            messageType = ToxMessageTypeAction
        default:
            panic("unknown message type")
    }
    message := make([]byte, c_length)
    if (c_length > 0) {
        C.memcpy(
            unsafe.Pointer(&message[0]),
            unsafe.Pointer(c_message),
            c_length,
        )
    }
    tox.onFriendMessage(tox, friendNumber, messageType, message)
}

//export callback_friend_lossless_packet
func callback_friend_lossless_packet(
    c_tox *C.Tox,
    c_friend_number C.uint32_t,
    c_data *C.uint8_t,
    c_length C.size_t,
    c_user_data unsafe.Pointer,
) {
    tox := (*Tox)(c_user_data)
    friendNumber := uint32(c_friend_number)
    data := make([]byte, c_length)
    if (c_length > 0) {
        C.memcpy(
            unsafe.Pointer(&data[0]),
            unsafe.Pointer(c_data),
            c_length,
        )
    }
    tox.onFriendLosslessPacket(tox, friendNumber, data)
}
