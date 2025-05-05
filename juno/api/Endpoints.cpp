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

    std::vector<Togglable*> devices;

    if (req.uuids()[0] == "*") {
        static auto filter = [](Device& device) -> bool {
            return static_cast<bool>(
              device.getImplementedInterfaces() & Device::Interface::togglable
            );
        };
        for (auto& device : co_await rpc::getDevices(mq, filter))
            devices.push_back(dynamic_cast<Togglable*>(device.get()));
    } else {
        for (auto& device : co_await rpc::getDevices(mq, toVector(req.uuids()))) {
            if (auto togglable = dynamic_cast<Togglable*>(device.get()); togglable) {
                devices.push_back(togglable);
            } else {
                throw Error{
                    Error::Code::invalidArgument,
                    "Device '{}' does not implement Togglabe interface", device->uuid
                };
            }
        }
    }

    for (auto& device : devices) co_await device->toggle();

    co_return api::AckResponse{};
}

}  // namespace juno
