#include <iostream>
#include <memory>
#include <string>
#include <thread>

#include <fmt/core.h>

#include "proto/juno.pb.h"
#include "proto/juno.grpc.pb.h"

#include "net/AsyncGrpcServer.hh"

class GrpcServer : public juno::AsyncGrpcServer {
public:
    explicit GrpcServer(boost::asio::io_context& ctx
    ) : AsyncGrpcServer(ctx, Config{ .host = "0.0.0.0", .port = 8001 }) {}

    void build(Builder&& builder) override {
        builder.addService<test::JunoService::AsyncService>()
          .addRequest<test::PingRequest, test::PongResponse>(
            &test::JunoService::AsyncService::RequestPing,
            [](const test::PingRequest& r
            ) -> juno::ResponseWrapper<test::PongResponse> {
                test::PongResponse res{};
                res.set_uuid("ping");
                co_return std::make_pair(res, grpc::Status::OK);
            }
          )
          .addRequest<test::PingRequest, test::PongResponse>(
            &test::JunoService::AsyncService::RequestTest,
            [](const test::PingRequest& r
            ) -> juno::ResponseWrapper<test::PongResponse> {
                test::PongResponse res{};
                res.set_uuid("test");
                co_return std::make_pair(res, grpc::Status::OK);
            }
          );
    }
};

int main() {
    kstd::log::init("juno");

    boost::asio::io_context ctx;

    GrpcServer server{ ctx };
    server.startAsync();

    ctx.run();

    return 0;
}
