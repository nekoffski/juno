#pragma once

#include "rpc/Service.hh"

#include <kstd/async/Core.hh>

#include "jobs/Job.hh"
#include "rpc/Messages.hh"

namespace juno {

class Scheduler : public RpcService<Scheduler> {
public:
    explicit Scheduler(boost::asio::io_context& io, kstd::AsyncMessenger& messenger);

    void start() override;
    void shutdown() override;

private:
    // -- message handlers
    kstd::Coro<void> handleRemoveJobsRequest(
      kstd::AsyncMessage& handle, const RemoveJobs::Request& r
    );
    kstd::Coro<void> handleAddJobRequest(
      kstd::AsyncMessage& handle, const AddJob::Request& r
    );

    // --

    boost::asio::io_context& m_io;
};

}  // namespace juno
