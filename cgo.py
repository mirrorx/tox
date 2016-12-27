#!/usr/bin/env python

import os

from waflib import Build, Logs
from waflib.Configure import conf

class TestContext(Build.BuildContext):
    cmd = fun = "test"

@conf
def check_var(self, var, **kw):
    self.start_msg("Checking for environment variable '%s'" % var)
    val = os.environ.get(var)
    if not val:
        Logs.warn("not found")
        self.fatal("The configuration failed")
    self.end_msg(val)
    self.env[var] = val
    return val

@conf
def init(self, **kw):
    self.check_var("LD_LIBRARY_PATH")
    gopath = self.check_var("GOPATH")
    goos = self.check_var("GOOS")
    goarch = self.check_var("GOARCH")
    prefix = os.path.join(gopath, "pkg/%s_%s" % (goos, goarch))
    self.env.PREFIX = prefix

@conf
def check_c_lib(self, lib, **kw):
    self.start_msg("Checking for library '%s'" % lib)
    archive = os.path.join(self.env.LD_LIBRARY_PATH, "lib%s.so" % lib)
    if not os.path.exists(archive):
        Logs.warn("not found")
        self.fatal("The configuration failed")
    self.end_msg("yes")

@conf
def check_g_lib(self, lib, **kw):
    self.start_msg("Checking for library '%s'" % lib)
    archive = os.path.join(self.env.PREFIX, "%s.a" % lib)
    if not os.path.exists(archive):
        Logs.warn("not found")
        self.fatal("The configuration failed")
    self.end_msg("yes")

def configure(ctx):
    ctx.init()
    ctx.find_program("gcc", var = "GCC")
    ctx.find_program("go", var = "GO")
