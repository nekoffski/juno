#pragma once

#include "Service.hh"

#include <kstd/async/Core.hh>
#include <kstd/async/AsyncMessenger.hh>

#include "jobs/Job.hh"

namespace juno {

class Scheduler : public Service {
public:
    explicit Scheduler(boost::asio::io_context& io, kstd::AsyncMessenger& messenger);

    void spawn() override;
    void shutdown() override;

private:
    boost::asio::io_context& m_io;
    kstd::AsyncMessenger::Queue* m_messageQueue;
};

}  // namespace juno
