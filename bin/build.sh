#!/bin/bash

SRC=$(dirname "$0")/../
conan build ${SRC} --profile ${SRC}/conf/profiles/debug
