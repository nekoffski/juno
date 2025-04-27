#include "DeviceProxy.hh"

#include "vendors/Yeelight.hh"
#include "Messages.hh"

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

kstd::Coro<void> DeviceProxy::handleMessages() {
    while (true) {
        auto message = co_await m_messageQueue->wait();

        if (message->is<ListDevices::Request>())
            co_await message->respond<ListDevices::Response>(getDevices());
    }
}

Devices DeviceProxy::getDevices() const {
    Devices allDevices;

    for (auto& vendor : m_vendors) {
        auto devices = vendor->getDevices();
        std::ranges::move(devices, std::back_inserter(allDevices));
    }

    return allDevices;
}

kstd::Coro<void> DeviceProxy::scan() {
    for (auto& vendor : m_vendors) co_await vendor->scan();
}

}  // namespace juno
