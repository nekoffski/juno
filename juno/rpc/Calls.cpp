#include "Calls.hh"

#include "Messages.hh"
#include "Queues.hh"

namespace juno::rpc {

static void handleError(kstd::AsyncResponse* res, const std::string& requestName) {
    if (res->is<Error>()) throw *res->as<Error>();
    throw Error{
        Error::Code::unspecified, "Unknown error handling request: '{}'", requestName
    };
}

kstd::Coro<Devices> getDevices(
  kstd::AsyncMessenger::Queue& mq, GetDevices::Request::Criteria criteria
) {
    auto handle =
      co_await mq.send<GetDevices::Request>(std::move(criteria))
        .to(DEVICE_PROXY_QUEUE);
    auto response = co_await handle->wait();

    if (response->is<GetDevices::Response>())
        co_return response->as<GetDevices::Response>()->devices;
    handleError(response.get(), "GetDevices");
}

kstd::Coro<void> removeJobs(
  kstd::AsyncMessenger::Queue& mq, const std::vector<std::string>& uuids
) {
    auto handle   = co_await mq.send<RemoveJobs::Request>(uuids).to(SCHEDULER_QUEUE);
    auto response = co_await handle->wait();
    if (response->is<RemoveJobs::Response>()) {
        auto notRemoved = response->as<RemoveJobs::Response>()->missingJobs;
        if (notRemoved.size() > 0) {
            // FIXME
            // throw Error{
            //
            // };
        }
    } else {
        handleError(response.get(), "RemoveJobs");
    }
}

}  // namespace juno::rpc
