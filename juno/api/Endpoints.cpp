#include "Endpoints.hh"

#include "net/Grpc.hh"

#include "Messages.hh"
#include "Queues.hh"
#include "Helpers.hh"

namespace juno {

kstd::Coro<api::PongResponse> pingEndpoint(const api::PingRequest& req) {
    const auto magic = req.magic();
    log::debug("Received ping request, magic: '{}'", magic);

    api::PongResponse res;
    res.set_magic(magic);
    co_return res;
}

kstd::Coro<api::ListDevicesResponse> listDevicesEndpoint(
  kstd::AsyncMessenger::Queue* mq
) {
    auto future   = co_await mq->send<ListDevices::Request>().to(DEVICE_PROXY_QUEUE);
    auto response = co_await future->wait();

    api::ListDevicesResponse res;
    for (auto& device : response->as<ListDevices::Response>()->devices)
        toProto(device.get(), res.add_devices());
    co_return res;
}

}  // namespace juno
