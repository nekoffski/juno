#pragma once

#include "Service.hh"

#include <kstd/async/Core.hh>

#include "rpc/Messages.hh"
#include "rpc/MessageQueueDestination.hh"

namespace juno {

class MetricService : public Service, public MessageQueueDestination<MetricService> {
public:
    explicit MetricService(
      boost::asio::io_context& io, kstd::AsyncMessenger& messenger
    );

    void spawn() override;
    void shutdown() override;

private:
    boost::asio::io_context& m_io;
};

}  // namespace juno
