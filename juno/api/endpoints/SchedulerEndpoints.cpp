#include "SchedulerEndpoints.hh"

#include "api/Helpers.hh"
#include "rpc/Calls.hh"
#include "Core.hh"

namespace juno {

kstd::Coro<api::AddJobResponse> addJobEndpoint(
  kstd::AsyncMessenger::Queue& mq, const api::AddJobRequest& req
) {
    const auto& jobBody = req.job();
    if (jobBody.empty())
        throw Error{ Error::Code::invalidArgument, "Empty Job body is not allowed" };

    const auto& jobUuid = co_await rpc::addJob(mq, jobBody);

    api::AddJobResponse res{};
    res.set_uuid(jobUuid);
    co_return res;
}

kstd::Coro<api::AckResponse> removeJobsEndpoint(
  kstd::AsyncMessenger::Queue& mq, const api::RemoveJobsRequest& req
) {
    if (req.uuids_size() == 0) {
        throw Error{
            Error::Code::invalidArgument, "At least one job uuid must be specified"
        };
    }
    co_await rpc::removeJobs(mq, toVector(req.uuids()));
    co_return api::AckResponse{};
}

}  // namespace juno
