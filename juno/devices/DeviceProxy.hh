#pragma once

#include <kstd/async/Core.hh>
#include <kstd/async/AsyncMessenger.hh>
#include <kstd/containers/FlatMap.hh>

#include "Vendor.hh"
#include "Device.hh"

#include "rpc/Messages.hh"
#include "rpc/Service.hh"

namespace juno {

class DeviceProxy : public RpcService<DeviceProxy> {
public:
    explicit DeviceProxy(
      boost::asio::io_context& io, kstd::AsyncMessenger& messenger
    );

    void start() override;
    void shutdown() override;

private:
    template <typename T, typename... Args>
    requires(std::derived_from<T, Vendor> && std::constructible_from<T, Args...>)
    void addVendor(Args&&... args) {
        m_vendors.push_back(kstd::makeUnique<T>(std::forward<Args>(args)...));
    }

    // -- message handlers
    kstd::Coro<void> handleGetDevicesRequest(
      kstd::AsyncMessage& handle, const GetDevices::Request& r
    );

    // --
    kstd::Coro<void> scan();
    std::vector<Device*> getDevices();

    boost::asio::io_context& m_io;
    kstd::AsyncMessenger::Queue* m_messageQueue;

    std::vector<kstd::UniquePtr<Vendor>> m_vendors;
    kstd::DynamicFlatMap<std::string, Device*> m_devices;
};

}  // namespace juno
