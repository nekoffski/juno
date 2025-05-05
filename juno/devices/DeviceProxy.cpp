#include "DeviceProxy.hh"

#include <kstd/Functional.hh>

#include "yeelight/Yeelight.hh"
#include "rpc/Messages.hh"

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

kstd::Coro<void> DeviceProxy::handleMessage(kstd::AsyncMessage& message) {
    if (message.is<GetDevices::Request>()) {
        co_await std::visit(
          kstd::Overloader{
            [&](const GetDevices::Request::Uuids& uuids) -> kstd::Coro<void> {
                if (uuids.size() == 0) {
                    co_await message.respond<GetDevices::Response>(getDevices());
                } else {
                    Devices devices;
                    for (const auto& uuid : uuids) {
                        if (not m_devices.contains(uuid)) {
                            co_await message.respond<Error>(
                              Error::Code::notFound,
                              "Could not find device with uuid: '{}'", uuid
                            );
                            co_return;
                        }
                        devices.push_back(m_devices.at(uuid));
                    }
                    co_await message.respond<GetDevices::Response>(std::move(devices)
                    );
                }
            },
            [&](const GetDevices::Request::Filter& filter) -> kstd::Coro<void> {
                Devices devices;
                for (auto& device : m_devices | std::views::values)
                    if (filter(*device)) devices.push_back(device);
                co_await message.respond<GetDevices::Response>(std::move(devices));
            },
          },
          message.as<GetDevices::Request>()->criteria
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
    auto values = m_devices | std::views::values;
    return Devices{ values.begin(), values.end() };
}

kstd::Coro<void> DeviceProxy::scan() {
    for (auto& vendor : m_vendors) {
        co_await vendor->scan();
        for (auto& device : vendor->getDevices()) m_devices[device->uuid] = device;
    }
}

}  // namespace juno
