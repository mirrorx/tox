/**
 * File        : tox.go
 * Copyright   : Copyright (c) 2015-2017 Mirror Labs, Inc. All rights reserved.
 * License     : GPLv3
 * Maintainer  : Enzo Haussecker <enzo@mirror.co>, Dominic Williams <dominic@string.technology>
 * Stability   : Experimental
 * Portability : Non-portable (requires Tox core at commit dcf2aaa)
 *
 * This module establishes a high-level API that allows clients to communicate
 * using the Tox protocol.
 */

package tox

//#cgo LDFLAGS: -l toxcore
//#include "callbacks.h"
//#include <memory.h>
import "C"
import "encoding/hex"
import "errors"
import "sync"
import "time"
import "unsafe"

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////// UTILITIES ///////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

// Convert a byte slice to a C array.
func slice2array(slice []byte) unsafe.Pointer {
    var size = len(slice)
    var array = C.malloc(C.size_t(size))
    var arrayPtr = uintptr(array)
    for i := 0; i < size; i ++ {
        *(*C.uint8_t)(unsafe.Pointer(arrayPtr)) = C.uint8_t(slice[i])
        arrayPtr++
    }
    return array
}

// Convert a C array to a byte slice.
func array2slice(array unsafe.Pointer, size int) []byte {
    var slice = make([]byte, size)
    var arrayPtr = uintptr(array)
    for i := 0; i < size; i ++ {
        slice[i] = byte(*(*C.uint8_t)(unsafe.Pointer(arrayPtr)))
        arrayPtr++
    }
    return slice
}

////////////////////////////////////////////////////////////////////////////////
/////////////////////////////// STARTUP OPTIONS ////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

// Convert startup options from C to Go.
func GoOptions(c_options *C.struct_Tox_Options) (options *ToxOptions, throw error) {
    options = &ToxOptions{}
    options.IPv6Enabled = bool(c_options.ipv6_enabled)
    options.UDPEnabled  = bool(c_options.udp_enabled)
    switch c_options.proxy_type {
        case C.TOX_PROXY_TYPE_NONE:
            options.ProxyType = ToxProxyTypeNone
        case C.TOX_PROXY_TYPE_HTTP:
            options.ProxyType = ToxProxyTypeHttp
        case C.TOX_PROXY_TYPE_SOCKS5:
            options.ProxyType = ToxProxyTypeSocks5
        default:
            return nil, errors.New("unknown proxy type")
    }
    options.ProxyHost = C.GoString(c_options.proxy_host)
    options.ProxyPort = uint16(c_options.proxy_port)
    options.StartPort = uint16(c_options.start_port)
    options.EndPort   = uint16(c_options.end_port)
    options.TCPPort   = uint16(c_options.tcp_port)
    options.SaveData  = array2slice(
        unsafe.Pointer(c_options.savedata_data),
        int(c_options.savedata_length),
    )
    return options, nil
}

// Convert startup options from Go to C. This result will contain C heap items
// that must be freed to prevent memory leaks. It is the caller's responsibility
// to arrange for them to be freed.
func COptions(options *ToxOptions) (c_options *C.struct_Tox_Options, throw error) {
    c_options = &C.struct_Tox_Options{}
    c_options.ipv6_enabled = C.bool(options.IPv6Enabled)
    c_options.udp_enabled  = C.bool(options.UDPEnabled)
    switch options.ProxyType {
        case ToxProxyTypeNone:
            c_options.proxy_type = C.TOX_PROXY_TYPE_NONE
        case ToxProxyTypeHttp:
            c_options.proxy_type = C.TOX_PROXY_TYPE_HTTP
        case ToxProxyTypeSocks5:
            c_options.proxy_type = C.TOX_PROXY_TYPE_SOCKS5
        default:
            return nil, errors.New("unknown proxy type")
    }
    c_options.proxy_host = C.CString(options.ProxyHost)
    c_options.proxy_port = C.uint16_t(options.ProxyPort)
    c_options.start_port = C.uint16_t(options.StartPort)
    c_options.end_port   = C.uint16_t(options.EndPort)
    c_options.tcp_port   = C.uint16_t(options.TCPPort)
    var length = len(options.SaveData)
    if (length == 0) {
        c_options.savedata_type = C.TOX_SAVEDATA_TYPE_NONE
    } else {
        c_options.savedata_type = C.TOX_SAVEDATA_TYPE_TOX_SAVE
    }
    c_options.savedata_data = (*C.uint8_t)(slice2array(options.SaveData))
    c_options.savedata_length = C.size_t(length)
    return c_options, nil
}

// Free all resources associated with a startup options object.
func (c_options *C.struct_Tox_Options) FreeOptions() {
    C.free(unsafe.Pointer(c_options.proxy_host))
    C.free(unsafe.Pointer(c_options.savedata_data))
}

// The default startup options for Tox.
func DefaultOptions() (options *ToxOptions, throw error) {
    var c_options C.struct_Tox_Options
    C.tox_options_default(unsafe.Pointer(&c_options))
    return GoOptions(&c_options)
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////// INSTANCE LIFECYCLE //////////////////////////////
////////////////////////////////////////////////////////////////////////////////

// Create or restore a Tox instance. This will bring the instance into a valid
// state. If the startup options are nil, then the default options are used.
func New(options *ToxOptions) (tox *Tox, throw error) {
    var c_options *C.struct_Tox_Options
    var c_error C.TOX_ERR_NEW
    if (options != nil) {
        c_options, throw = COptions(options)
        if throw != nil {
            return
        }
        defer c_options.FreeOptions()
    }
    var c_tox = C.tox_new(c_options, &c_error)
    if (c_error != C.TOX_ERR_NEW_OK) {
        switch c_error {
            case C.TOX_ERR_NEW_NULL:
                throw = ToxErrNewNull
            case C.TOX_ERR_NEW_MALLOC:
                throw = ToxErrNewMalloc
            case C.TOX_ERR_NEW_PORT_ALLOC:
                throw = ToxErrNewPortAlloc
            case C.TOX_ERR_NEW_PROXY_BAD_TYPE:
                throw = ToxErrNewProxyBadType
            case C.TOX_ERR_NEW_PROXY_BAD_HOST:
                throw = ToxErrNewProxyBadHost
            case C.TOX_ERR_NEW_PROXY_BAD_PORT:
                throw = ToxErrNewProxyBadPort
            case C.TOX_ERR_NEW_PROXY_NOT_FOUND:
                throw = ToxErrNewProxyNotFound
            case C.TOX_ERR_NEW_LOAD_ENCRYPTED:
                throw = ToxErrNewLoadEncrypted
            case C.TOX_ERR_NEW_LOAD_BAD_FORMAT:
                throw = ToxErrNewLoadBadFormat
            default:
                throw = ToxErrUnknown
        }
    } else {
        tox = &Tox {
            handle: c_tox,
            lock: sync.Mutex{},
        }
    }
    return
}

// Serialize a Tox instance.
func (tox *Tox) Serialize() (data []byte) {
    var c_length = C.tox_get_savedata_size(tox.handle)
    data = make([]byte, c_length)
    if (c_length > 0) {
        C.tox_get_savedata(tox.handle, (*C.uint8_t)(&data[0]))
    }
    return
}

// Destroy a Tox instance. This will disconnect the instance from the Tox
// network and release all other resources associated with it. The Tox pointer
// becomes invalid and can no longer be used.
func (tox *Tox) Destroy() {
    C.tox_kill(tox.handle)
}

////////////////////////////////////////////////////////////////////////////////
/////////////////////////////// EVENT PROCESSING ///////////////////////////////
////////////////////////////////////////////////////////////////////////////////

// Run the main event processing loop.
func (tox *Tox) Process() {
    tox.lock.Lock()
    C.tox_iterate(tox.handle)
    tox.lock.Unlock()
}

// Get the iteration interval in milliseconds.
func (tox *Tox) ProcessDelay() time.Duration {
    var c_millis = C.tox_iteration_interval(tox.handle)
    return time.Duration(uint32(c_millis)) * time.Millisecond
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////// CALLBACK FUNCTIONS //////////////////////////////
////////////////////////////////////////////////////////////////////////////////

// This function registers a function that executes when the connection status
// of the client changes.
func (tox *Tox) SetOnSelfConnectionStatus(callback OnSelfConnectionStatus) {
    tox.onSelfConnectionStatus = callback
    C.register_self_connection_status(tox.handle, unsafe.Pointer(tox))
}

// This function registers a function that executes when receiving a friend
// request.
func (tox *Tox) SetOnFriendRequest(callback OnFriendRequest) {
    tox.onFriendRequest = callback
    C.register_friend_request(tox.handle, unsafe.Pointer(tox))
}

// This function registers a function that executes when a friend changes their
// name.
func (tox *Tox) SetOnFriendName(callback OnFriendName) {
    tox.onFriendName = callback
    C.register_friend_name(tox.handle, unsafe.Pointer(tox))
}

// This function registers a function that executes when a friend changes their
// status.
func (tox *Tox) SetOnFriendStatus(callback OnFriendStatus) {
    tox.onFriendStatus = callback
    C.register_friend_status(tox.handle, unsafe.Pointer(tox))
}

// This function registers a function that executes when a friend changes their
// status message.
func (tox *Tox) SetOnFriendStatusMessage(callback OnFriendStatusMessage) {
    tox.onFriendStatusMessage = callback
    C.register_friend_status_message(tox.handle, unsafe.Pointer(tox))
}

// This function registers a function that executes when the connection status
// of a friend changes.
func (tox *Tox) SetOnFriendConnectionStatus(callback OnFriendConnectionStatus) {
    tox.onFriendConnectionStatus = callback
    C.register_friend_connection_status(tox.handle, unsafe.Pointer(tox))
}

// This function registers a function that executes when receiving a chat
// message from a friend.
func (tox *Tox) SetOnFriendMessage(callback OnFriendMessage) {
    tox.onFriendMessage = callback
    C.register_friend_message(tox.handle, unsafe.Pointer(tox))
}

// This function registers a function that executes when receiving a custom
// loss-less packet from a friend.
func (tox *Tox) SetOnFriendLosslessPacket(callback OnFriendLosslessPacket) {
    tox.onFriendLosslessPacket = callback
    C.register_friend_lossless_packet(tox.handle, unsafe.Pointer(tox))
}

////////////////////////////////////////////////////////////////////////////////
///////////////////////////////// CLIENT STATE /////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

// Get the address of the Tox client.
func (tox *Tox) GetAddress() (address ToxAddress) {
    C.tox_self_get_address(tox.handle, (*C.uint8_t)(&address[0]))
    return
}

// Get the no-spam value of the Tox client.
func (tox *Tox) GetNoSpam() uint32 {
    return uint32(C.tox_self_get_nospam(tox.handle))
}

// Set the no-spam value of the Tox client.
func (tox *Tox) SetNoSpam(nospam uint32) {
    C.tox_self_set_nospam(tox.handle, C.uint32_t(nospam))
}

// Get the public key of the Tox client.
func (tox *Tox) GetPublicKey() (publicKey ToxPublicKey) {
    C.tox_self_get_public_key(tox.handle, (*C.uint8_t)(&publicKey[0]))
    return
}

// Get the secret key of the Tox client.
func (tox *Tox) GetSecretKey() (secretKey ToxSecretKey) {
    C.tox_self_get_secret_key(tox.handle, (*C.uint8_t)(&secretKey[0]))
    return
}

// Get the name of the Tox client.
func (tox *Tox) GetName() (name []byte) {
    var c_length = C.tox_self_get_name_size(tox.handle)
    var c_name *C.uint8_t
    name = make([]byte, c_length)
    if (c_length > 0) {
        c_name = (*C.uint8_t)(&name[0])
    }
    C.tox_self_get_name(tox.handle, c_name)
    return
}

// Set the name of the Tox client.
func (tox *Tox) SetName(name []byte) (throw error) {
    var c_length = C.size_t(len(name))
    var c_name *C.uint8_t
    var c_error C.TOX_ERR_SET_INFO
    if (c_length > 0) {
        c_name = (*C.uint8_t)(&name[0])
    }
    C.tox_self_set_name(tox.handle, c_name, c_length, &c_error)
    if (c_error != C.TOX_ERR_SET_INFO_OK) {
        switch c_error {
            case C.TOX_ERR_SET_INFO_NULL:
                throw = ToxErrSetInfoNull
            case C.TOX_ERR_SET_INFO_TOO_LONG:
                throw = ToxErrSetInfoTooLong
            default:
                throw = ToxErrUnknown
        }
    }
    return
}

// Get the status of the Tox client.
func (tox *Tox) GetStatus() (userStatus ToxUserStatus) {
    var c_user_status = C.tox_self_get_status(tox.handle)
    switch c_user_status {
        case C.TOX_USER_STATUS_AWAY:
            userStatus = ToxUserStatusAway
        case C.TOX_USER_STATUS_BUSY:
            userStatus = ToxUserStatusBusy
        default:
            userStatus = ToxUserStatusNone
    }
    return
}

// Set the status of the Tox client.
func (tox *Tox) SetStatus(userStatus ToxUserStatus) {
    var c_user_status C.TOX_USER_STATUS
    switch userStatus {
        case ToxUserStatusAway:
            c_user_status = C.TOX_USER_STATUS_AWAY
        case ToxUserStatusBusy:
            c_user_status = C.TOX_USER_STATUS_BUSY
        default:
            c_user_status = C.TOX_USER_STATUS_NONE
    }
    C.tox_self_set_status(tox.handle, c_user_status)
}

// Get the status message of the Tox client.
func (tox *Tox) GetStatusMessage() (message []byte) {
    var c_length = C.tox_self_get_status_message_size(tox.handle)
    var c_message *C.uint8_t
    message = make([]byte, c_length)
    if (c_length > 0) {
        c_message = (*C.uint8_t)(&message[0])
    }
    C.tox_self_get_status_message(tox.handle, c_message)
    return
}

// Set the status message of the Tox client.
func (tox *Tox) SetStatusMessage(message []byte) (throw error) {
    var c_length = C.size_t(len(message))
    var c_message *C.uint8_t
    var c_error C.TOX_ERR_SET_INFO
    if (c_length > 0) {
        c_message = (*C.uint8_t)(&message[0])
    }
    C.tox_self_set_status_message(tox.handle, c_message, c_length, &c_error)
    if (c_error != C.TOX_ERR_SET_INFO_OK) {
        switch c_error {
            case C.TOX_ERR_SET_INFO_NULL:
                throw = ToxErrSetInfoNull
            case C.TOX_ERR_SET_INFO_TOO_LONG:
                throw = ToxErrSetInfoTooLong
            default:
                throw = ToxErrUnknown
        }
    }
    return
}

// Get the connection status of the Tox client.
func (tox *Tox) GetConnectionStatus() (connectionStatus ToxConnectionStatus) {
    var c_connection_status = C.tox_self_get_connection_status(tox.handle)
    switch c_connection_status {
        case C.TOX_CONNECTION_TCP:
            connectionStatus = ToxConnectionTCP
        case C.TOX_CONNECTION_UDP:
            connectionStatus = ToxConnectionUDP
        default:
            connectionStatus = ToxConnectionNone
    }
    return
}

// Get the friend list of the Tox client.
func (tox *Tox) GetFriendList() (friendList []uint32) {
    var c_length = C.tox_self_get_friend_list_size(tox.handle)
    var c_friend_list *C.uint32_t
    friendList = make([]uint32, c_length)
    if (c_length > 0) {
        c_friend_list = (*C.uint32_t)(&friendList[0])
    }
    C.tox_self_get_friend_list(tox.handle, c_friend_list)
    return
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////// FRIEND MANAGEMENT ///////////////////////////////
////////////////////////////////////////////////////////////////////////////////

// Add a friend.
func (tox *Tox) FriendAdd(address ToxAddress, message []byte) (friendNumber uint32, throw error) {
    var c_address = (*C.uint8_t)(&address[0])
    var c_length = C.size_t(len(message))
    var c_message *C.uint8_t
    var c_error C.TOX_ERR_FRIEND_ADD
    if (c_length > 0) {
        c_message = (*C.uint8_t)(&message[0])
    }
    var c_friend_number = C.tox_friend_add(tox.handle, c_address, c_message, c_length, &c_error)
    if (c_error != C.TOX_ERR_FRIEND_ADD_OK) {
        switch c_error {
            case C.TOX_ERR_FRIEND_ADD_NULL:
                throw = ToxErrFriendAddNull
            case C.TOX_ERR_FRIEND_ADD_TOO_LONG:
                throw = ToxErrFriendAddTooLong
            case C.TOX_ERR_FRIEND_ADD_NO_MESSAGE:
                throw = ToxErrFriendAddNoMessage
            case C.TOX_ERR_FRIEND_ADD_OWN_KEY:
                throw = ToxErrFriendAddOwnKey
            case C.TOX_ERR_FRIEND_ADD_ALREADY_SENT:
                throw = ToxErrFriendAddAlreadySent
            case C.TOX_ERR_FRIEND_ADD_BAD_CHECKSUM:
                throw = ToxErrFriendAddBadChecksum
            case C.TOX_ERR_FRIEND_ADD_SET_NEW_NOSPAM:
                throw = ToxErrFriendAddSetNewNoSpam
            case C.TOX_ERR_FRIEND_ADD_MALLOC:
                throw = ToxErrFriendAddMalloc
            default:
                throw = ToxErrUnknown
        }
    } else {
        friendNumber = uint32(c_friend_number)
    }
    return
}

// Add a friend without sending a friend request.
func (tox *Tox) FriendAddNoRequest(publicKey ToxPublicKey) (friendNumber uint32, throw error) {
    var c_public_key = (*C.uint8_t)(&publicKey[0])
    var c_error C.TOX_ERR_FRIEND_ADD
    var c_friend_number = C.tox_friend_add_norequest(tox.handle, c_public_key, &c_error)
    if (c_error != C.TOX_ERR_FRIEND_ADD_OK) {
        switch c_error {
            case C.TOX_ERR_FRIEND_ADD_NULL:
                throw = ToxErrFriendAddNull
            case C.TOX_ERR_FRIEND_ADD_TOO_LONG:
                throw = ToxErrFriendAddTooLong
            case C.TOX_ERR_FRIEND_ADD_NO_MESSAGE:
                throw = ToxErrFriendAddNoMessage
            case C.TOX_ERR_FRIEND_ADD_OWN_KEY:
                throw = ToxErrFriendAddOwnKey
            case C.TOX_ERR_FRIEND_ADD_ALREADY_SENT:
                throw = ToxErrFriendAddAlreadySent
            case C.TOX_ERR_FRIEND_ADD_BAD_CHECKSUM:
                throw = ToxErrFriendAddBadChecksum
            case C.TOX_ERR_FRIEND_ADD_SET_NEW_NOSPAM:
                throw = ToxErrFriendAddSetNewNoSpam
            case C.TOX_ERR_FRIEND_ADD_MALLOC:
                throw = ToxErrFriendAddMalloc
            default:
                throw = ToxErrUnknown
        }
    } else {
        friendNumber = uint32(c_friend_number)
    }
    return
}

// Delete a friend.
func (tox *Tox) FriendDelete(friendNumber uint32) (throw error) {
    var c_friend_number = C.uint32_t(friendNumber)
    var c_error C.TOX_ERR_FRIEND_DELETE
    C.tox_friend_delete(tox.handle, c_friend_number, &c_error)
    if (c_error != C.TOX_ERR_FRIEND_DELETE_OK) {
        switch c_error {
            case C.TOX_ERR_FRIEND_DELETE_FRIEND_NOT_FOUND:
                throw = ToxErrFriendDeleteFriendNotFound
            default:
                throw = ToxErrUnknown
        }
    }
    return
}

// Check if a friend exists.
func (tox *Tox) FriendExists(friendNumber uint32) bool {
    var c_friend_number = C.uint32_t(friendNumber)
    return bool(C.tox_friend_exists(tox.handle, c_friend_number))
}

////////////////////////////////////////////////////////////////////////////////
///////////////////////////////// FRIEND STATE /////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

// Get the name of a friend.
func (tox *Tox) FriendGetName(friendNumber uint32) (name []byte, throw error) {
    var c_friend_number = C.uint32_t(friendNumber)
    var c_error C.TOX_ERR_FRIEND_QUERY
    var c_length = C.tox_friend_get_name_size(tox.handle, c_friend_number, &c_error)
    if (c_error != C.TOX_ERR_FRIEND_QUERY_OK) {
        switch c_error {
            case C.TOX_ERR_FRIEND_QUERY_NULL:
                throw = ToxErrFriendQueryNull
            case C.TOX_ERR_FRIEND_QUERY_FRIEND_NOT_FOUND:
                throw = ToxErrFriendQueryFriendNotFound
            default:
                throw = ToxErrUnknown
        }
    } else {
        var c_name *C.uint8_t
        name = make([]byte, c_length)
        if (c_length > 0) {
            c_name = (*C.uint8_t)(&name[0])
        }
        C.tox_friend_get_name(tox.handle, c_friend_number, c_name, &c_error)
        if (c_error != C.TOX_ERR_FRIEND_QUERY_OK) {
            name = nil
            switch c_error {
                case C.TOX_ERR_FRIEND_QUERY_NULL:
                    throw = ToxErrFriendQueryNull
                case C.TOX_ERR_FRIEND_QUERY_FRIEND_NOT_FOUND:
                    throw = ToxErrFriendQueryFriendNotFound
                default:
                    throw = ToxErrUnknown
            }
        }
    }
    return
}

// Get the public key of a friend.
func (tox *Tox) FriendGetPublicKey(friendNumber uint32) (publicKey ToxPublicKey, throw error) {
    var c_friend_number = C.uint32_t(friendNumber)
    var c_public_key = (*C.uint8_t)(&publicKey[0])
    var c_error C.TOX_ERR_FRIEND_GET_PUBLIC_KEY
    C.tox_friend_get_public_key(tox.handle, c_friend_number, c_public_key, &c_error)
    if (c_error != C.TOX_ERR_FRIEND_GET_PUBLIC_KEY_OK) {
        switch c_error {
            case C.TOX_ERR_FRIEND_GET_PUBLIC_KEY_FRIEND_NOT_FOUND:
                throw = ToxErrFriendGetPublicKeyFriendNotFound
            default:
                throw = ToxErrUnknown
        }
    }
    return
}

// Get the friend associated with the given public key.
func (tox *Tox) FriendByPublicKey(publicKey ToxPublicKey) (friendNumber uint32, throw error) {
    var c_public_key = (*C.uint8_t)(&publicKey[0])
    var c_error C.TOX_ERR_FRIEND_BY_PUBLIC_KEY
    var c_friend_number = C.tox_friend_by_public_key(tox.handle, c_public_key, &c_error)
    if (c_error != C.TOX_ERR_FRIEND_BY_PUBLIC_KEY_OK) {
        switch c_error {
            case C.TOX_ERR_FRIEND_BY_PUBLIC_KEY_NULL:
                throw = ToxErrFriendByPublicKeyNull
            case C.TOX_ERR_FRIEND_BY_PUBLIC_KEY_NOT_FOUND:
                throw = ToxErrFriendByPublicKeyNotFound
            default:
                throw = ToxErrUnknown
        }
    } else {
        friendNumber = uint32(c_friend_number)
    }
    return
}

// Get the status of a friend.
func (tox *Tox) FriendGetStatus(friendNumber uint32) (userStatus ToxUserStatus, throw error) {
    var c_friend_number = C.uint32_t(friendNumber)
    var c_error C.TOX_ERR_FRIEND_QUERY
    var c_status = C.tox_friend_get_status(tox.handle, c_friend_number, &c_error)
    if (c_error != C.TOX_ERR_FRIEND_QUERY_OK) {
        switch c_error {
            case C.TOX_ERR_FRIEND_QUERY_NULL:
                throw = ToxErrFriendQueryNull
            case C.TOX_ERR_FRIEND_QUERY_FRIEND_NOT_FOUND:
                throw = ToxErrFriendQueryFriendNotFound
            default:
                throw = ToxErrUnknown
        }
    } else {
        switch c_status {
            case C.TOX_USER_STATUS_AWAY:
                userStatus = ToxUserStatusAway
            case C.TOX_USER_STATUS_BUSY:
                userStatus = ToxUserStatusBusy
            default:
                userStatus = ToxUserStatusNone
        }
    }
    return
}

// Get the status message of a friend.
func (tox *Tox) FriendGetStatusMessage(friendNumber uint32) (message []byte, throw error) {
    var c_friend_number = C.uint32_t(friendNumber)
    var c_error C.TOX_ERR_FRIEND_QUERY
    var c_length = C.tox_friend_get_status_message_size(tox.handle, c_friend_number, &c_error)
    if (c_error != C.TOX_ERR_FRIEND_QUERY_OK) {
        switch c_error {
            case C.TOX_ERR_FRIEND_QUERY_NULL:
                throw = ToxErrFriendQueryNull
            case C.TOX_ERR_FRIEND_QUERY_FRIEND_NOT_FOUND:
                throw = ToxErrFriendQueryFriendNotFound
            default:
                throw = ToxErrUnknown
        }
    } else {
        var c_message *C.uint8_t
        message = make([]byte, c_length)
        if (c_length > 0) {
            c_message = (*C.uint8_t)(&message[0])
        }
        C.tox_friend_get_status_message(tox.handle, c_friend_number, c_message, &c_error)
        if (c_error != C.TOX_ERR_FRIEND_QUERY_OK) {
            message = nil
            switch c_error {
                case C.TOX_ERR_FRIEND_QUERY_NULL:
                    throw = ToxErrFriendQueryNull
                case C.TOX_ERR_FRIEND_QUERY_FRIEND_NOT_FOUND:
                    throw = ToxErrFriendQueryFriendNotFound
                default:
                    throw = ToxErrUnknown
            }
        }
    }
    return
}

// Get the connection status of a friend.
func (tox *Tox) FriendGetConnectionStatus(friendNumber uint32) (connectionStatus ToxConnectionStatus, throw error) {
    var c_friend_number = C.uint32_t(friendNumber)
    var c_error C.TOX_ERR_FRIEND_QUERY
    var c_connection_status = C.tox_friend_get_connection_status(tox.handle, c_friend_number, &c_error)
    if (c_error != C.TOX_ERR_FRIEND_QUERY_OK) {
        switch c_error {
            case C.TOX_ERR_FRIEND_QUERY_NULL:
                throw = ToxErrFriendQueryNull
            case C.TOX_ERR_FRIEND_QUERY_FRIEND_NOT_FOUND:
                throw = ToxErrFriendQueryFriendNotFound
            default:
                throw = ToxErrUnknown
        }
    } else {
        switch c_connection_status {
            case C.TOX_CONNECTION_TCP:
                connectionStatus = ToxConnectionTCP
            case C.TOX_CONNECTION_UDP:
                connectionStatus = ToxConnectionUDP
            default:
                connectionStatus = ToxConnectionNone
        }
    }
    return
}

// Get the last time a friend was seen online.
func (tox *Tox) FriendGetLastOnline(friendNumber uint32) (timestamp time.Time, throw error) {
    var c_friend_number = C.uint32_t(friendNumber)
    var c_error C.TOX_ERR_FRIEND_GET_LAST_ONLINE
    var c_timestamp = C.tox_friend_get_last_online(tox.handle, c_friend_number, &c_error)
    if (c_error != C.TOX_ERR_FRIEND_GET_LAST_ONLINE_OK) {
        switch c_error {
            case C.TOX_ERR_FRIEND_GET_LAST_ONLINE_FRIEND_NOT_FOUND:
                throw = ToxErrFriendGetLastOnlineFriendNotFound
            default:
                throw = ToxErrUnknown
        }
    } else {
        timestamp = time.Unix(int64(c_timestamp), 0)
    }
    return
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////// DATA TRANSMISSION ///////////////////////////////
////////////////////////////////////////////////////////////////////////////////

// Send a chat message to an online friend.
func (tox *Tox) FriendSendMessage(friendNumber uint32, messageType ToxMessageType, message []byte) (messageId uint32, throw error) {
    var c_friend_number = C.uint32_t(friendNumber)
    var c_message_type C.TOX_MESSAGE_TYPE
    var c_length = C.size_t(len(message))
    var c_message *C.uint8_t
    var c_error C.TOX_ERR_FRIEND_SEND_MESSAGE
    switch messageType {
        case ToxMessageTypeAction:
            c_message_type = C.TOX_MESSAGE_TYPE_ACTION
        default:
            c_message_type = C.TOX_MESSAGE_TYPE_NORMAL
    }
    if (c_length > 0) {
        c_message = (*C.uint8_t)(&message[0])
    }
    var c_message_id = C.tox_friend_send_message(tox.handle, c_friend_number, c_message_type, c_message, c_length, &c_error)
    if (c_error != C.TOX_ERR_FRIEND_SEND_MESSAGE_OK) {
        switch c_error {
            case C.TOX_ERR_FRIEND_SEND_MESSAGE_NULL:
                throw = ToxErrFriendSendMessageNull
            case C.TOX_ERR_FRIEND_SEND_MESSAGE_FRIEND_NOT_FOUND:
                throw = ToxErrFriendSendMessageFriendNotFound
            case C.TOX_ERR_FRIEND_SEND_MESSAGE_FRIEND_NOT_CONNECTED:
                throw = ToxErrFriendSendMessageFriendNotConnected
            case C.TOX_ERR_FRIEND_SEND_MESSAGE_SENDQ:
                throw = ToxErrFriendSendMessageSendQ
            case C.TOX_ERR_FRIEND_SEND_MESSAGE_TOO_LONG:
                throw = ToxErrFriendSendMessageTooLong
            case C.TOX_ERR_FRIEND_SEND_MESSAGE_EMPTY:
                throw = ToxErrFriendSendMessageEmpty
            default:
                throw = ToxErrUnknown
        }
    } else {
        messageId = uint32(c_message_id)
    }
    return
}

// Send a custom loss-less packet to an online friend.
func (tox *Tox) FriendSendLosslessPacket(friendNumber uint32, data []byte) (throw error) {
    var c_friend_number = C.uint32_t(friendNumber)
    var c_length = C.size_t(len(data))
    var c_data *C.uint8_t
    var c_error C.TOX_ERR_FRIEND_CUSTOM_PACKET
    if (c_length > 0) {
        c_data = (*C.uint8_t)(&data[0])
    }
    C.tox_friend_send_lossless_packet(tox.handle, c_friend_number, c_data, c_length, &c_error)
    if (c_error != C.TOX_ERR_FRIEND_CUSTOM_PACKET_OK) {
        switch c_error {
            case C.TOX_ERR_FRIEND_CUSTOM_PACKET_NULL:
                throw = ToxErrFriendCustomPacketNull
            case C.TOX_ERR_FRIEND_CUSTOM_PACKET_FRIEND_NOT_FOUND:
                throw = ToxErrFriendCustomPacketFriendNotFound
            case C.TOX_ERR_FRIEND_CUSTOM_PACKET_FRIEND_NOT_CONNECTED:
                throw = ToxErrFriendCustomPacketFriendNotConnected
            case C.TOX_ERR_FRIEND_CUSTOM_PACKET_INVALID:
                throw = ToxErrFriendCustomPacketInvalid
            case C.TOX_ERR_FRIEND_CUSTOM_PACKET_EMPTY:
                throw = ToxErrFriendCustomPacketEmpty
            case C.TOX_ERR_FRIEND_CUSTOM_PACKET_TOO_LONG:
                throw = ToxErrFriendCustomPacketTooLong
            case C.TOX_ERR_FRIEND_CUSTOM_PACKET_SENDQ:
                throw = ToxErrFriendCustomPacketSendQ
            default:
                throw = ToxErrUnknown
        }
    }
    return
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////// NETWORKING //////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

// This is the default seed node for this library. It is used for testing
// purposes only. Never use this in production since the host name is subject to
// change.
func DefaultSeedNode() *SeedNode {
    return NewSeedNode(
        "tox.zodiaclabs.org",
        33445,
        "A09162D68618E742FFBCA1C2C70385E6679604B2D80EA6E84AD0996A1AC8A074",
    )
}

// This function creates a new seed node. You can find an active list of nodes
// at https://wiki.tox.im/Nodes
func NewSeedNode(host string, port uint16, publicKey string) *SeedNode {
    return &SeedNode { Host: host, Port: port, PublicKey: publicKey }
}

// This function will establish a connection to the given seed node. It will
// attempt to connect using UDP and TCP at the same time. Tox will use the node
// as a TCP relay if ToxOptions.UDPEnabled was false, and also to connect to
// friends that are in TCP-only mode. Tox will also use the TCP connection when
// NAT hole punching is slow, and later switch to UDP if hole punching succeeds.
func (tox *Tox) Bootstrap(seedNode *SeedNode) (throw error) {
    var c_host = C.CString(seedNode.Host)
    defer C.free(unsafe.Pointer(c_host))
    var c_port = C.uint16_t(seedNode.Port)
    publicKey, err := hex.DecodeString(seedNode.PublicKey)
    if err != nil {
        return err
    }
    if (len(publicKey) != ToxPublicKeySize) {
        return errors.New("invalid public key")
    }
    var c_public_key = (*C.uint8_t)(&publicKey[0])
    var c_error C.TOX_ERR_BOOTSTRAP
    C.tox_bootstrap(tox.handle, c_host, c_port, c_public_key, &c_error)
    if (c_error != C.TOX_ERR_BOOTSTRAP_OK) {
        switch c_error {
            case C.TOX_ERR_BOOTSTRAP_NULL:
                throw = ToxErrBootstrapNull
            case C.TOX_ERR_BOOTSTRAP_BAD_HOST:
                throw = ToxErrBootstrapBadHost
            case C.TOX_ERR_BOOTSTRAP_BAD_PORT:
                throw = ToxErrBootstrapBadPort
            default:
                throw = ToxErrUnknown
        }
    }
    return
}
