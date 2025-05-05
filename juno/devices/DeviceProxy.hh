#pragma once

#include <kstd/async/Core.hh>
#include <kstd/async/AsyncMessenger.hh>

#include "Vendor.hh"
#include "Device.hh"

#include "Service.hh"

namespace juno {

class DeviceProxy : public Service {
public:
    explicit DeviceProxy(
      boost::asio::io_context& io, kstd::AsyncMessenger& messenger
    );

    void spawn() override;
    void shutdown() override;

private:
    kstd::Coro<void> handleMessages();

    template <typename T, typename... Args>
    requires(std::derived_from<T, Vendor> && std::constructible_from<T, Args...>)
    void addVendor(Args&&... args) {
        m_vendors.push_back(kstd::makeUnique<T>(std::forward<Args>(args)...));
    }

    kstd::Coro<void> scan();
    Devices getDevices() const;

    boost::asio::io_context& m_io;
    kstd::AsyncMessenger::Queue* m_messageQueue;

    std::vector<kstd::UniquePtr<Vendor>> m_vendors;
    std::unordered_map<std::string, kstd::SharedPtr<Device>> m_devices;
};

}  // namespace juno
