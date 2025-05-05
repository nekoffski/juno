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

static kstd::Coro<Devices> getDevices(
  kstd::AsyncMessenger::Queue& mq, GetDevices::Request&& request
) {
    auto handle =
      co_await mq.send<GetDevices::Request>(std::move(request))
        .to(DEVICE_PROXY_QUEUE);
    auto response = co_await handle->wait();

    if (response->is<GetDevices::Response>())
        co_return response->as<GetDevices::Response>()->devices;
    handleError(response.get(), "GetDevices");
}

kstd::Coro<Devices> getDevices(
  kstd::AsyncMessenger::Queue& mq, const GetDevices::Request::Uuids& uuids
) {
    co_return (co_await getDevices(mq, GetDevices::Request{ uuids }));
}

kstd::Coro<Devices> getDevices(
  kstd::AsyncMessenger::Queue& mq, const GetDevices::Request::Filter& filter
) {
    co_return (co_await getDevices(mq, GetDevices::Request{ filter }));
}

}  // namespace juno::rpc
