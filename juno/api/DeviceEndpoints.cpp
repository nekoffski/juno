#include "DeviceEndpoints.hh"

#include "net/Grpc.hh"

#include "rpc/Messages.hh"
#include "rpc/Queues.hh"
#include "rpc/Calls.hh"
#include "Helpers.hh"

namespace juno {

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

    std::vector<Togglable*> togglabeDevices;

    if (req.uuids()[0] == "*") {
        auto devices = co_await rpc::getDevices(mq, Device::Interface::togglable);
        togglabeDevices.reserve(devices.size());
        std::ranges::transform(
          devices, std::back_inserter(togglabeDevices),
          [](auto& device) { return dynamic_cast<Togglable*>(device.get()); }
        );
    } else {
        auto devices = co_await rpc::getDevices(mq, toVector(req.uuids()));
        togglabeDevices.reserve(devices.size());
        for (auto& device : devices) {
            if (auto togglable = dynamic_cast<Togglable*>(device.get()); togglable) {
                togglabeDevices.push_back(togglable);
            } else {
                throw Error{
                    Error::Code::invalidArgument,
                    "Device '{}' does not implement Togglabe interface", device->uuid
                };
            }
        }
    }
    for (auto& device : togglabeDevices) co_await device->toggle();
    co_return api::AckResponse{};
}

}  // namespace juno
