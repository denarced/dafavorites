#!/bin/sh

ctags -R .
vim `find . -type f -name \*.go`
