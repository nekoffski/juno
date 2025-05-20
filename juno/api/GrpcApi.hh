#pragma once

#include <kstd/async/AsyncMessenger.hh>

#include "net/Grpc.hh"
#include "rpc/Service.hh"

namespace juno {

class GrpcApi : public RpcService<GrpcApi>, private AsyncGrpcServer {
public:
    explicit GrpcApi(boost::asio::io_context& io, kstd::AsyncMessenger& messenger);

    void start() override;
    void shutdown() override;

private:
    void build(Builder&&) override;
};

}  // namespace juno
