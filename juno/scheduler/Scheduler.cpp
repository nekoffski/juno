#include "Scheduler.hh"

#include "rpc/Queues.hh"
#include "jobs/JobParser.hh"

namespace juno {

Scheduler::Scheduler(boost::asio::io_context& io, kstd::AsyncMessenger& messenger) :
    m_io(io), m_messageQueue(messenger.registerQueue(SCHEDULER_QUEUE)) {}

void Scheduler::spawn() {
    std::string job =
      "DECLARE JOB TYPE=oneshot,  NAME=job_0,  COMMAND=toggle_bulb, ARGS=(), DELAY=5s";

    try {
        JobParser{}.parseString(job);
    } catch (const Error& e) {
        log::warn("Could not parse job: {} - {}", e.what(), e.where());
    }
}

void Scheduler::shutdown() {}

}  // namespace juno
