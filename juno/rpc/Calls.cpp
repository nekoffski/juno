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
  kstd::AsyncMessenger::Queue& mq, const std::vector<std::string>& uuids
) {
    auto h = co_await mq.send<ListDevices::Request>(uuids).to(DEVICE_PROXY_QUEUE);
    auto response = co_await h->wait();

    if (response->is<ListDevices::Response>())
        co_return response->as<ListDevices::Response>()->devices;
    handleError(response.get(), "ListDevices");
}

}  // namespace juno::rpc
