#pragma once

#include <unordered_map>

#include <kstd/async/Core.hh>
#include <kstd/Id.hh>

#include "devices/Vendor.hh"
#include "devices/Device.hh"

namespace juno {

class YeelightVendor : public Vendor {
public:
    explicit YeelightVendor(boost::asio::io_context& io);

    Devices getDevices() const override;
    kstd::Coro<void> scan() override;

private:
    kstd::Coro<void> processNewDevice(const std::string& payload);

    Devices m_devices;

    boost::asio::io_context& m_io;
    boost::asio::ip::udp::socket m_socket;
    boost::asio::ip::udp::endpoint m_yeelightEndpoint;

    std::string m_discoverMessage;
};

}  // namespace juno
