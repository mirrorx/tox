/**
 * File        : types.go
 * Copyright   : Copyright (c) 2015-2017 Mirror Labs, Inc. All rights reserved.
 * License     : GPLv3
 * Maintainer  : Enzo Haussecker <enzo@mirror.co>, Dominic Williams <dominic@string.technology>
 * Stability   : Experimental
 * Portability : Non-portable (requires Tox core at commit dcf2aaa)
 */

package tox

//#include <tox/tox.h>
import "C"
import "sync"
import "unsafe"

////////////////////////////////////////////////////////////////////////////////
///////////////////////////////// STRUCT TYPES /////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

// This type represents a Tox instance. All the state associated with a
// connection is held within the instance. Multiple instances can exist and
// operate concurrently. The maximum number of Tox instances that can exist on
// a single network device is limited. Note that this is not just a per-process
// limit, since the limiting factor is the number of usable ports on a device.
type Tox struct {

    handle                   *C.Tox
    lock                     sync.Mutex
    onSelfConnectionStatus   OnSelfConnectionStatus
    onFriendRequest          OnFriendRequest
    onFriendName             OnFriendName
    onFriendStatus           OnFriendStatus
    onFriendStatusMessage    OnFriendStatusMessage
    onFriendConnectionStatus OnFriendConnectionStatus
    onFriendMessage          OnFriendMessage
    onFriendLosslessPacket   OnFriendLosslessPacket
    userData                 unsafe.Pointer

}

// This type represents the options associated with creating a new Tox instance.
type ToxOptions struct {

    // The type of socket to create. If this is set to false, an IPv4 socket is
    // created, which subsequently only allows IPv4 communication. If it is set
    // to true, an IPv6 socket is created, allowing both IPv4 and IPv6
    // communication.
    IPv6Enabled bool

    // Enable the use of UDP communication when available. Setting this to false
    // will force Tox to use TCP only. Communications will need to be relayed
    // through a TCP relay node, potentially slowing them down. Disabling UDP
    // support is necessary when using anonymous proxies or Tor.
    UDPEnabled bool

    // Pass communications through a proxy of this type.
    ProxyType ToxProxyType

    // The IP address or DNS name of the proxy to be used. If used, this must be
    // non-nil and be a valid DNS name. The name must not exceed 255 characters.
    // The value is ignored if ProxyType is ToxProxyTypeNone.
    ProxyHost string

    // The port to use to connect to the proxy server. Ports must be in the
    // range (1, 65535). The value is ignored if ProxyType is ToxProxyTypeNone.
    ProxyPort uint16

    // The start port of the inclusive port range to attempt to use. If both
    // StartPort and EndPort are 0, the default port range will be used: [33445,
    // 33545]. If either StartPort or EndPort is 0 while the other is non-zero,
    // the non-zero port will be the only port in the range. Having StartPort >
    // EndPort will yield the same behavior as if StartPort and EndPort were
    // swapped.
    StartPort uint16

    // The end port of the inclusive port range to attempt to use.
    EndPort uint16

    // The port to use for the TCP server (relay). If 0, the TCP server is
    // disabled. Enabling it is not required for Tox to function properly. When
    // enabled, your Tox instance can act as a TCP relay for other Tox
    // instances. This leads to increased traffic, thus when writing a client it
    // is recommended to enable TCP server only if the user has an option to
    // disable it.
    TCPPort uint16

    // The save data. This data is produced by serializing a Tox instance and
    // supplied as a startup option to restore the instance to its active state.
    SaveData []byte

}

// This type represents a seed node. In order to facilitate quick connections
// with other peers on the network, Tox employs seed nodes that each client
// connects to in order to retrieve a list of current clients connected to the
// pool.
type SeedNode struct {

    Host string
    Port uint16
    PublicKey string

}

////////////////////////////////////////////////////////////////////////////////
//////////////////////////////// CALLBACK TYPES ////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

// This type represents a function that executes when the connection status of
// the client changes. The function can be registered as a callback using
// SetOnSelfConnectionStatus.
type OnSelfConnectionStatus func(

    tox *Tox, connectionStatus ToxConnectionStatus,

)

// This type represents a function that executes when receiving a friend
// request. The function can be registered as a callback using
// SetOnFriendRequest.
type OnFriendRequest func(

    tox *Tox, publicKey ToxPublicKey, message []byte,

)

// This type represents a function that executes when a friend changes their
// name. The function can be registered as a callback using SetOnFriendName.
type OnFriendName func(

    tox *Tox, friendNumber uint32, name []byte,

)

// This type represents a function that executes when a friend changes their
// status. The function can be registered as a callback using SetOnFriendStatus.
type OnFriendStatus func(

    tox *Tox, friendNumber uint32, status ToxUserStatus,

)

// This type represents a function that executes when a friend changes their
// status message. The function can be registered as a callback using
// SetOnFriendStatusMessage.
type OnFriendStatusMessage func(

    tox *Tox, friendNumber uint32, message []byte,

)

// This type represents a function that executes when the connection status of
// a friend changes. The function can be registered as a callback using
// SetOnFriendConnectionStatus.
type OnFriendConnectionStatus func(

    tox *Tox, friendNumber uint32, connectionStatus ToxConnectionStatus,

)

// This type represents a function that executes when receiving a chat message
// from a friend. The function can be registered as a callback using
// SetOnFriendMessage.
type OnFriendMessage func(

    tox *Tox, friendNumber uint32, messageType ToxMessageType, message []byte,

)

// This type represents a function that executes when receiving a custom
// loss-less packet from a friend. The function can be registered as a callback
// using SetOnFriendLosslessPacket.
type OnFriendLosslessPacket func(

    tox *Tox, friendNumber uint32, data []byte,

)

////////////////////////////////////////////////////////////////////////////////
/////////////////////////////// ENUMERATED TYPES ///////////////////////////////
////////////////////////////////////////////////////////////////////////////////

// This type represents a Tox connection status.
type ToxConnectionStatus int

// The set of possible statuses that a Tox connection can have. These could
// relate to the client itself or the friend of a client.
const (

    // No connection has been established.
    ToxConnectionNone ToxConnectionStatus = iota

    // A TCP connection has been established. For the client, this means it is
    // only connected through a TCP relay. For the friend, this means the
    // connection to that particular friend goes through a TCP relay.
    ToxConnectionTCP

    // A UDP connection has been established. For the client, this means it is
    // able to send UDP packets to DHT nodes, but may still be connected to a
    // TCP relay. For the friend, this means the connection to that particular
    // friend was built using direct UDP packets.
    ToxConnectionUDP

)

// This type represents a Tox user status.
type ToxUserStatus int

// The set of possible statuses that a Tox user can have. The user can either be
// available, unavailable after a defined period of inactivity, or unavailable
// after signaling to others that the user does not want to communicate.
const (

    ToxUserStatusNone ToxUserStatus = iota
    ToxUserStatusAway
    ToxUserStatusBusy

)

// This type represents a Tox message type.
type ToxMessageType int

// The set of possible message types. These could relate to friend messages or
// group chat messages. Messages can be normal text messages or describe a user
// action.
const (

    ToxMessageTypeNormal = iota
    ToxMessageTypeAction

)

// This type represents a Tox proxy configuration.
type ToxProxyType int

// The set of possible proxy configurations that a Tox client can have. This
// includes no proxy configuration, an HTTP proxy configuration, or a SOCKS5
// proxy configuration.
const (

    ToxProxyTypeNone ToxProxyType = iota
    ToxProxyTypeHttp
    ToxProxyTypeSocks5

)

////////////////////////////////////////////////////////////////////////////////
///////////////////////////////// ARRAY TYPES //////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

// This type represents a Tox public key.
type ToxPublicKey [ToxPublicKeySize]byte

// This type represents a Tox secret key.
type ToxSecretKey [ToxSecretKeySize]byte

// This type represents a Tox address.
type ToxAddress [ToxAddressSize]byte

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////// CONSTANTS ///////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

// A set of numeric constants synonymous with their C-side counterparts.
const (

    ToxPublicKeySize          = C.TOX_PUBLIC_KEY_SIZE
    ToxSecretKeySize          = C.TOX_SECRET_KEY_SIZE
    ToxAddressSize            = C.TOX_ADDRESS_SIZE
    ToxMaxNameLength          = C.TOX_MAX_NAME_LENGTH
    ToxMaxStatusMessageLength = C.TOX_MAX_STATUS_MESSAGE_LENGTH
    ToxMaxFriendRequestLength = C.TOX_MAX_FRIEND_REQUEST_LENGTH
    ToxMaxMessageLength       = C.TOX_MAX_MESSAGE_LENGTH
    ToxMaxCustomPacketSize    = C.TOX_MAX_CUSTOM_PACKET_SIZE
    ToxHashLength             = C.TOX_HASH_LENGTH
    ToxFileIdLength           = C.TOX_FILE_ID_LENGTH
    ToxMaxFilenameLength      = C.TOX_MAX_FILENAME_LENGTH

)
