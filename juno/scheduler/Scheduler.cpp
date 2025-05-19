#include "Scheduler.hh"

#include "rpc/Queues.hh"
#include "jobs/JobParser.hh"

namespace juno {

Scheduler::Scheduler(boost::asio::io_context& io, kstd::AsyncMessenger& messenger) :
    MessageQueueDestination(this, messenger, SCHEDULER_QUEUE), m_io(io) {
    registerCall<RemoveJobs::Request>(&juno::Scheduler::handleRemoveJobsRequest);
    registerCall<AddJob::Request>(&juno::Scheduler::handleAddJobRequest);
}

void Scheduler::spawn() {
    kstd::spawn(m_io.get_executor(), [&]() -> kstd::Coro<void> {
        co_await startHandling([&]() -> bool { return not isRunning(); });
    });

    std::string job =
      "DECLARE JOB TYPE=oneshot,  NAME=job_0,  COMMAND=toggle_bulb, ARGS=(), DELAY=5s";

    try {
        JobParser{}.parseString(job);
    } catch (const Error& e) {
        log::warn("Could not parse job: {} - {}", e.what(), e.where());
    }
}

kstd::Coro<void> Scheduler::handleRemoveJobsRequest(
  kstd::AsyncMessage& handle, const RemoveJobs::Request& r
) {
    co_await handle.respond<Error>(Error::Code::notFound, "Could not find jobs");
}

kstd::Coro<void> Scheduler::handleAddJobRequest(
  kstd::AsyncMessage& handle, const AddJob::Request& r
) {
    co_await handle.respond<Error>(Error::Code::notFound, "not implemented yet");
}

void Scheduler::shutdown() {}

}  // namespace juno
