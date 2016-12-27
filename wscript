#!/usr/bin/env python

top = "."
out = "build"

def configure(ctx):
    ctx.load("cgo")
    ctx.check_c_lib("toxcore")
    ctx.check_g_lib("golang.org/x/crypto/curve25519")

def build(ctx):
    modules = ctx.path.ant_glob("*.go", excl = ["*_test.go"])
    ctx(name = "build",
        rule = "${GO} build -o ${TGT} -i ${SRC}",
        source = modules,
        target = "tox.a")
    if ctx.is_install:
        headers = ctx.path.ant_glob("*.h")
        ctx.install_files("${GOPATH}/src/mirrorx/tox", modules + headers)
        ctx.install_files("${PREFIX}/mirrorx", "tox.a")

def test(ctx):
    ctx(name = "tests",
        rule = "go test ..")
