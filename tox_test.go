/**
 * File        : tox_test.go
 * Copyright   : Copyright (c) 2015-2017 Mirror Labs, Inc. All rights reserved.
 * License     : GPLv3
 * Maintainer  : Enzo Haussecker <enzo@mirror.co>, Dominic Williams <dominic@string.technology>
 * Stability   : Experimental
 * Portability : Non-portable (requires Tox core at commit dcf2aaa)
 *
 * This module provides a test suite for the high-level API that allows clients
 * to communicate using the Tox protocol.
 */

package tox

import "bytes"
import "golang.org/x/crypto/curve25519"
import "math/rand"
import "testing"
import "time"

func RandomOptions(noise *rand.Rand) (options *ToxOptions, err error) {
    options = &ToxOptions{}
    options.IPv6Enabled = noise.Intn(2) % 2 == 0
    options.UDPEnabled  = noise.Intn(2) % 2 == 0
    switch noise.Intn(3) {
        case 0: options.ProxyType = ToxProxyTypeNone
        case 1: options.ProxyType = ToxProxyTypeHttp
        case 2: options.ProxyType = ToxProxyTypeSocks5
    }
    var buffer bytes.Buffer
    capacity := noise.Intn(256)
    hexchars := []byte("0123456789ABCDEF")
    for i := 0; i <= capacity; i++ {
        buffer.WriteByte(hexchars[noise.Intn(len(hexchars))])
    }
    options.ProxyHost = buffer.String()
    options.ProxyPort = uint16(noise.Intn(65535) + 1)
    options.StartPort = uint16(noise.Intn(65535) + 1)
    options.EndPort   = uint16(noise.Intn(65535) + 1)
    options.TCPPort   = uint16(noise.Intn(65535) + 1)
    client, err := New(nil)
    options.SaveData = client.Serialize()
    return
}

func TestConvertOptions(test *testing.T) {
    noise := rand.New(rand.NewSource(time.Now().UnixNano()))
    options, err := RandomOptions(noise)
    if err != nil {
        test.Fatal(err)
    }
    c_options, err := COptions(options)
    defer c_options.FreeOptions()
    result, err := GoOptions(c_options)
    if (options.IPv6Enabled != result.IPv6Enabled) {
        test.Fatalf("Failed to convert Tox startup options. IPv6 enabled option does not match.")
    }
    if (options.UDPEnabled != result.UDPEnabled) {
        test.Fatalf("Failed to convert Tox startup options. UDP enabled option does not match.")
    }
    if (options.ProxyType != result.ProxyType) {
        test.Fatalf("Failed to convert Tox startup options. Proxy type option does not match.")
    }
    if (options.ProxyHost != result.ProxyHost) {
        test.Fatalf("Failed to convert Tox startup options. Proxy host option does not match.")
    }
    if (options.ProxyPort != result.ProxyPort) {
        test.Fatalf("Failed to convert Tox startup options. Proxy port option does not match.")
    }
    if (options.StartPort != result.StartPort) {
        test.Fatalf("Failed to convert Tox startup options. Start port option does not match.")
    }
    if (options.EndPort != result.EndPort) {
        test.Fatalf("Failed to convert Tox startup options. End port option does not match.")
    }
    if (options.TCPPort != result.TCPPort) {
        test.Fatalf("Failed to convert Tox startup options. TCP port option does not match.")
    }
    if (!equal(options.SaveData, result.SaveData)) {
        test.Fatalf("Failed to convert Tox startup options. Save data option does not match.")
    }
}

////////////////////////////////////////////////////////////////////////////////
///////////////////////////////// MEMORY TESTS /////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

func TestSetGetNoSpam(test *testing.T) {
    tox := initialise(test)
    defer tox.Destroy()
    noise := rand.New(rand.NewSource(time.Now().UnixNano()))
    input := noise.Uint32()
    tox.SetNoSpam(input)
    output := tox.GetNoSpam()
    if input != output {
        test.Fatalf("Failed memory test for Tox no-spam value.")
    }
}

func TestSetGetName(test *testing.T) {
    tox := initialise(test)
    defer tox.Destroy()
    noise := rand.New(rand.NewSource(time.Now().UnixNano()))
    input := make([]byte, noise.Intn(ToxMaxNameLength+1))
    for i := range input {
        input[i] = byte(noise.Intn(256))
    }
    err := tox.SetName(input)
    if err != nil {
        test.Fatal(err)
    }
    output := tox.GetName()
    if !equal(input, output) {
        test.Fatalf("Failed memory test for Tox name.")
    }
}

func TestSetGetStatus(test *testing.T) {
    tox := initialise(test)
    defer tox.Destroy()
    noise := rand.New(rand.NewSource(time.Now().UnixNano()))
    var input ToxUserStatus
    switch noise.Intn(3) {
        case 0: input = ToxUserStatusNone
        case 1: input = ToxUserStatusAway
        case 2: input = ToxUserStatusBusy
    }
    tox.SetStatus(input)
    output := tox.GetStatus()
    if input != output {
        test.Fatalf("Failed memory test for Tox status.")
    }
}

func TestSetGetStatusMessage(test *testing.T) {
    tox := initialise(test)
    defer tox.Destroy()
    noise := rand.New(rand.NewSource(time.Now().UnixNano()))
    input := make([]byte, noise.Intn(ToxMaxStatusMessageLength+1))
    for i := range input {
        input[i] = byte(noise.Intn(256))
    }
    err := tox.SetStatusMessage(input)
    if err != nil {
        test.Fatal(err)
    }
    output := tox.GetStatusMessage()
    if !equal(input, output) {
        test.Fatalf("Failed memory test for Tox status message.")
    }
}

////////////////////////////////////////////////////////////////////////////////
///////////////////////////////// CRYPTO TESTS /////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

func TestPublicPrivateKeys(test *testing.T) {
    tox := initialise(test)
    defer tox.Destroy()
    publicKey := tox.GetPublicKey()
    secretKey := tox.GetSecretKey()
    var product [32]byte
    curve25519.ScalarBaseMult(&product, (*[32]byte)(&secretKey))
    if product != publicKey {
        test.Fatalf("Failed to generate valid Tox public/private key pair.")
    }
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////// UTILITIES ///////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

func initialise(test *testing.T) (*Tox) {
    tox, err := New(nil)
    if err != nil {
        test.Fatal(err)
    }
    return tox 
}

func equal(a, b []byte) bool {
    if len(a) != len(b) {
        return false
    }
    for i := range a {
        if a[i] != b[i] {
            return false
        }
    }
    return true
}
