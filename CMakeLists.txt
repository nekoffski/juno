cmake_minimum_required(VERSION 3.10)
project(juno-server)

set(CMAKE_EXPORT_COMPILE_COMMANDS ON)
set(CMAKE_BUILD_TYPE Debug)
set(JUNO_INCLUDE ${CMAKE_CURRENT_SOURCE_DIR}/src)

# set (CMAKE_CXX_FLAGS_DEBUG "${CMAKE_CXX_FLAGS_DEBUG} -fno-omit-frame-pointer -fsanitize=address")
# set (CMAKE_LINKER_FLAGS_DEBUG "${CMAKE_LINKER_FLAGS_DEBUG} -fno-omit-frame-pointer -fsanitize=address")

set(GCC_COVERAGE_COMPILE_FLAGS "--coverage -O0 -fprofile-arcs -ftest-coverage -fno-elide-constructors -fno-inline")
set(GCC_COVERAGE_LINK_FLAGS "-lgcov")
set(JUNO_COMPILER_FLAGS "")

find_package(kstd CONFIG)
set(JUNO_LIBS kstd::kstd)

option(JUNO_ENABLE_UNIT_TESTS "Build unit tests" OFF)
option(JUNO_ENABLE_BENCHMARKS "Build benchmakrs " OFF)
option(JUNO_ENABLE_COVERAGE "Enable code coverage" OFF)

if(JUNO_ENABLE_COVERAGE)
    message("-- Triggering build with coverage support")
    set(CMAKE_CXX_FLAGS "${CMAKE_CXX_FLAGS} ${GCC_COVERAGE_COMPILE_FLAGS}")
    set(CMAKE_EXE_LINKER_FLAGS "${CMAKE_EXE_LINKER_FLAGS} ${GCC_COVERAGE_LINK_FLAGS}")
    string(REPLACE "-O3" "" CMAKE_CXX_FLAGS_RELEASE ${CMAKE_CXX_FLAGS_RELEASE})
else()
    set(CMAKE_CXX_FLAGS "${CMAKE_CXX_FLAGS} -O1")
endif()

set(JUNO_COMPILER_FLAGS -Wall -Wextra -Wpedantic -Werror)

if(JUNO_ENABLE_UNIT_TESTS)
    enable_testing()
endif()

add_subdirectory(${CMAKE_CURRENT_SOURCE_DIR}/juno)
