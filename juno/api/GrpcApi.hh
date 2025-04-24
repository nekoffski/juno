#pragma once

#include <kstd/async/AsyncMessenger.hh>

#include "messages/Queues.hh"
#include "net/AsyncGrpcServer.hh"
#include "Config.hh"

#include "Service.hh"

namespace juno {

class GrpcApi : public Service, private AsyncGrpcServer {
public:
    explicit GrpcApi(
      boost::asio::io_context& io, kstd::AsyncMessenger& messenger,
      const juno::Config& cfg
    );

    void spawn() override;
    void shutdown() override;

private:
    void build(Builder&&) override;

    kstd::AsyncMessenger::Queue* m_messageQueue;
};

}  // namespace juno
