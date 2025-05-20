#include "DeviceProxy.hh"

#include <kstd/Functional.hh>

#include "yeelight/Yeelight.hh"

namespace juno {

DeviceProxy::DeviceProxy(
  boost::asio::io_context& io, kstd::AsyncMessenger& messenger
) : RpcService(io, this, messenger, DEVICE_PROXY_QUEUE), m_io(io) {
    addVendor<YeelightVendor>(m_io);

    registerCall<GetDevices::Request>(&juno::DeviceProxy::handleGetDevicesRequest);
}

void DeviceProxy::start() {
    spawn([&]() -> kstd::Coro<void> { co_await scan(); });
}

void DeviceProxy::shutdown() {}

kstd::Coro<void> DeviceProxy::handleGetDevicesRequest(
  kstd::AsyncMessage& handle, const GetDevices::Request& r
) {
    co_await std::visit(
      kstd::Overloader{
        [&]([[maybe_unused]] const GetDevices::Request::All&) -> kstd::Coro<void> {
            co_await handle.respond<GetDevices::Response>(getDevices());
        },
        [&](const GetDevices::Request::Uuids& uuids) -> kstd::Coro<void> {
            std::vector<Device*> devices;
            devices.reserve(uuids.size());
            for (const auto& uuid : uuids) {
                if (not m_devices.contains(uuid)) {
                    co_await handle.respond<Error>(
                      Error::Code::notFound, "Could not find device with uuid: '{}'",
                      uuid
                    );
                    co_return;
                }
                devices.push_back(*m_devices.get(uuid));
            }
            co_await handle.respond<GetDevices::Response>(std::move(devices));
        },
        [&](const GetDevices::Request::Filter& filter) -> kstd::Coro<void> {
            co_await handle.respond<GetDevices::Response>(
              getDevices()
              | std::views::filter([&](auto& device) { return filter(*device); })
              | kstd::toVector<Device*>()
            );
        },
        [&](const Device::Interface& interfaces) -> kstd::Coro<void> {
            co_await handle.respond<GetDevices::Response>(
              getDevices() | std::views::filter([&](auto& device) {
                  return device->implements(interfaces);
              })
              | kstd::toVector<Device*>()
            );
        },
      },
      r.criteria
    );
}

std::vector<Device*> DeviceProxy::getDevices() { return m_devices.getValues(); }

kstd::Coro<void> DeviceProxy::scan() {
    u64 devicesDiscovered = 0u;
    for (auto& vendor : m_vendors) {
        for (auto& device : co_await vendor->scan()) {
            m_devices.insert(device->getUuid(), device);
            ++devicesDiscovered;
        }
    }
    log::info("Scan finished, {} new devices discovered", devicesDiscovered);
}

}  // namespace juno
