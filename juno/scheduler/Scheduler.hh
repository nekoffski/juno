#pragma once

#include "Service.hh"

#include <kstd/async/Core.hh>

#include "jobs/Job.hh"
#include "rpc/Messages.hh"
#include "rpc/RemoteCallee.hh"

namespace juno {

class Scheduler : public Service, public RemoteCallee<Scheduler> {
public:
    explicit Scheduler(boost::asio::io_context& io, kstd::AsyncMessenger& messenger);

    void spawn() override;
    void shutdown() override;

private:
    // -- message handlers
    kstd::Coro<void> handleRemoveJobsRequest(
      kstd::AsyncMessage& handle, const RemoveJobs::Request& r
    );

    // --

    boost::asio::io_context& m_io;
};

}  // namespace juno
