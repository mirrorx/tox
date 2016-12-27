/**
 * File        : callbacks.h
 * Copyright   : Copyright (c) 2015-2017 Mirror Labs, Inc. All rights reserved.
 * License     : GPLv3
 * Maintainer  : Enzo Haussecker <enzo@mirror.co>, Dominic Williams <dominic@string.technology>
 * Stability   : Experimental
 * Portability : Non-portable (requires Tox core at commit dcf2aaa)
 */

#include <stdlib.h>
#include <tox/tox.h>

void callback_self_connection_status(struct Tox *, TOX_CONNECTION, void *);
void callback_friend_name(struct Tox *, uint32_t, const uint8_t *, size_t, void *);
void callback_friend_request(struct Tox *, const uint8_t *, const uint8_t *, size_t, void *);
void callback_friend_status_message(struct Tox *, uint32_t, const uint8_t *, size_t, void *);
void callback_friend_status(struct Tox *, uint32_t, TOX_USER_STATUS, void *);
void callback_friend_connection_status(struct Tox *, uint32_t, TOX_CONNECTION, void *);
void callback_friend_message(struct Tox *, uint32_t, TOX_MESSAGE_TYPE, const uint8_t *, size_t, void *);
void callback_friend_lossless_packet(struct Tox *, uint32_t, const uint8_t *, size_t, void *);

// We cannot register our callbacks directly from Go. This macro creates a C
// function that registers a pointer to our callback function defined in Go.
#define GEN_CALLBACK_API(x) \
static void register_##x(Tox *tox, void *t) { \
    tox_callback_##x(tox, callback_##x, t); \
}

GEN_CALLBACK_API(self_connection_status)
GEN_CALLBACK_API(friend_name)
GEN_CALLBACK_API(friend_request)
GEN_CALLBACK_API(friend_status_message)
GEN_CALLBACK_API(friend_status)
GEN_CALLBACK_API(friend_connection_status)
GEN_CALLBACK_API(friend_message)
GEN_CALLBACK_API(friend_lossless_packet)
