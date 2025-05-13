
set(JUNO_TARGET juno-server)
file(GLOB_RECURSE JUNO_SRC ${CMAKE_CURRENT_SOURCE_DIR}/*.cpp)

get_target_property(GRPC_CPP_PLUGIN_EXECUTABLE gRPC::grpc_cpp_plugin IMPORTED_LOCATION)

foreach(PROTO_FILE ${PROTO_FILES})
    get_filename_component(PROTO_NAME ${PROTO_FILE} NAME_WE)

    set(GENERATED_PROTO_SRCS "${CMAKE_CURRENT_BINARY_DIR}/${PROTO_NAME}.pb.cc")
    set(GENERATED_PROTO_HDRS "${CMAKE_CURRENT_BINARY_DIR}/${PROTO_NAME}.pb.h")
    set(GENERATED_GRPC_SRCS "${CMAKE_CURRENT_BINARY_DIR}/${PROTO_NAME}.grpc.pb.cc")
    set(GENERATED_GRPC_HDRS "${CMAKE_CURRENT_BINARY_DIR}/${PROTO_NAME}.grpc.pb.h")

    add_custom_command(
        OUTPUT ${GENERATED_PROTO_SRCS} ${GENERATED_PROTO_HDRS}
               ${GENERATED_GRPC_SRCS} ${GENERATED_GRPC_HDRS}
        COMMAND ${Protobuf_PROTOC_EXECUTABLE}
            --proto_path=${PROTO_DIR}
            --cpp_out=${CMAKE_CURRENT_BINARY_DIR}
            --grpc_out=${CMAKE_CURRENT_BINARY_DIR}
            --plugin=protoc-gen-grpc=${GRPC_CPP_PLUGIN_EXECUTABLE}
            ${PROTO_FILE}
        DEPENDS ${PROTO_FILE}
        COMMENT "Generating protobuf/grpc for ${PROTO_NAME}.proto"
    )

    list(APPEND PROTO_SRCS ${GENERATED_PROTO_SRCS} ${GENERATED_GRPC_SRCS})
    list(APPEND PROTO_INCLUDE ${GENERATED_PROTO_HDRS} ${GENERATED_GRPC_HDRS})
endforeach()

add_executable(${JUNO_TARGET} ${JUNO_SRC} ${PROTO_SRCS})
target_include_directories(${JUNO_TARGET} PRIVATE ${JUNO_INCLUDE} ${CMAKE_CURRENT_BINARY_DIR})
target_link_libraries(${JUNO_TARGET} PUBLIC ${JUNO_LIBS}) 
target_compile_options(${JUNO_TARGET} PRIVATE ${JUNO_COMPILER_FLAGS})
