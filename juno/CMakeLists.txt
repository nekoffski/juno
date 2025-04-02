
set(JUNO_TARGET juno-server)
file(GLOB_RECURSE JUNO_SRC ${CMAKE_CURRENT_SOURCE_DIR}/*.cpp)

add_executable(${JUNO_TARGET} ${JUNO_SRC})
target_include_directories(${JUNO_TARGET} PRIVATE ${JUNO_INCLUDE})
target_link_libraries(${JUNO_TARGET} PUBLIC ${JUNO_LIBS})
target_compile_options(${JUNO_TARGET} PRIVATE ${JUNO_COMPILER_FLAGS})
