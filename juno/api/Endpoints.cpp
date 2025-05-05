#include "Endpoints.hh"

#include "net/Grpc.hh"

#include "rpc/Messages.hh"
#include "rpc/Queues.hh"
#include "rpc/Calls.hh"
#include "Helpers.hh"

namespace juno {

kstd::Coro<api::PongResponse> pingEndpoint(const api::PingRequest& req) {
    const auto magic = req.magic();
    log::trace("Received ping request, magic: '{}'", magic);

    api::PongResponse res;
    res.set_magic(magic);
    co_return res;
}

kstd::Coro<api::ListDevicesResponse> listDevicesEndpoint(
  kstd::AsyncMessenger::Queue& mq
) {
    api::ListDevicesResponse res;
    for (const auto& device : co_await rpc::getDevices(mq))
        device->toProto(res.add_devices());
    co_return res;
}

kstd::Coro<api::AckResponse> toggleDevicesEndpoint(
  kstd::AsyncMessenger::Queue& mq, const api::ToggleDevicesRequest& req
) {
    if (req.uuids_size() == 0) {
        throw Error{
            Error::Code::invalidArgument, "Need at least one device's uuid"
        };
    }

    auto devices = co_await rpc::getDevices(mq, toVector(req.uuids()));
    for (auto& device : devices) {
        // todo: extend rtti to support flags about provided interfaces
        co_await reinterpret_cast<Togglable*>(device.get())->toggle();
    }

    co_return api::AckResponse{};
}

}  // namespace juno
