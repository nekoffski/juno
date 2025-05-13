#pragma once

#include <kstd/async/AsyncMessenger.hh>

#include "net/Grpc.hh"
#include "Service.hh"

namespace juno {

class GrpcApi : public Service, private AsyncGrpcServer {
public:
    explicit GrpcApi(boost::asio::io_context& io, kstd::AsyncMessenger& messenger);

    void spawn() override;
    void shutdown() override;

private:
    void build(Builder&&) override;

    kstd::AsyncMessenger::Queue* m_mq;
};

}  // namespace juno
