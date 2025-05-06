#include "DeviceProxy.hh"

#include <kstd/Functional.hh>

#include "yeelight/Yeelight.hh"

namespace juno {

DeviceProxy::DeviceProxy(
  boost::asio::io_context& io, kstd::AsyncMessenger& messenger
) : m_io(io), m_messageQueue(messenger.registerQueue(DEVICE_PROXY_QUEUE)) {
    addVendor<YeelightVendor>(m_io);
}

void DeviceProxy::spawn() {
    kstd::spawn(m_io.get_executor(), [&]() -> kstd::Coro<void> { co_await scan(); });
    kstd::spawn(m_io.get_executor(), [&]() -> kstd::Coro<void> {
        co_await handleMessages();
    });
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
            Devices devices;
            devices.reserve(uuids.size());
            for (const auto& uuid : uuids) {
                if (not m_devices.contains(uuid)) {
                    co_await handle.respond<Error>(
                      Error::Code::notFound, "Could not find device with uuid: '{}'",
                      uuid
                    );
                    co_return;
                }
                devices.push_back(m_devices.at(uuid));
            }
            co_await handle.respond<GetDevices::Response>(std::move(devices));
        },
        [&](const GetDevices::Request::Filter& filter) -> kstd::Coro<void> {
            co_await handle.respond<GetDevices::Response>(
              m_devices | std::views::values
              | std::views::filter([&](auto& device) { return filter(*device); })
              | kstd::toVector<kstd::SharedPtr<Device>>()
            );
        },
        [&](const Device::Interface& interfaces) -> kstd::Coro<void> {
            co_await handle.respond<GetDevices::Response>(
              m_devices | std::views::values | std::views::filter([&](auto& device) {
                  return device->implements(interfaces);
              })
              | kstd::toVector<kstd::SharedPtr<Device>>()
            );
        },
      },
      r.criteria
    );
}

kstd::Coro<void> DeviceProxy::handleMessage(kstd::AsyncMessage& message) {
    if (message.is<GetDevices::Request>()) {
        co_await handleGetDevicesRequest(
          message, *message.as<GetDevices::Request>()
        );
    }
}

kstd::Coro<void> DeviceProxy::handleMessages() {
    while (true) {
        auto message = co_await m_messageQueue->wait();
        co_await handleMessage(*message);
    }
}

Devices DeviceProxy::getDevices() const {
    return m_devices | std::views::values
           | kstd::toVector<kstd::SharedPtr<Device>>();
}

kstd::Coro<void> DeviceProxy::scan() {
    for (auto& vendor : m_vendors) {
        co_await vendor->scan();
        for (auto& device : vendor->getDevices()) m_devices[device->uuid] = device;
    }
}

}  // namespace juno
