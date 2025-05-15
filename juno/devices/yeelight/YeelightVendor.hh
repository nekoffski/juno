#pragma once

#include <unordered_map>

#include <kstd/async/Core.hh>
#include <kstd/Id.hh>
#include <kstd/memory/UniquePtr.hh>

#include "devices/Vendor.hh"
#include "devices/Device.hh"

namespace juno {

class YeelightVendor : public Vendor {
public:
    explicit YeelightVendor(boost::asio::io_context& io);

    kstd::Coro<std::vector<Device*>> scan() override;

private:
    kstd::Coro<Device*> processNewDevice(const std::string& payload);

    std::vector<kstd::UniquePtr<Device>> m_devices;

    boost::asio::io_context& m_io;
    boost::asio::ip::udp::socket m_socket;
    boost::asio::ip::udp::endpoint m_yeelightEndpoint;

    std::string m_discoverMessage;
};

}  // namespace juno
