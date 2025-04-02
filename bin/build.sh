#!/bin/bash

SRC=$(dirname "$0")/../

conan install ${SRC} --output-folder=${SRC}/build --build=missing --profile ${SRC}/conf/profiles/debug
conan build ${SRC} --profile ${SRC}/conf/profiles/debug



