#include "YeelightVendor.hh"

#include <boost/algorithm/string.hpp>
#include <kstd/async/Utils.hh>
#include <kstd/String.hh>

#include "Core.hh"

#include "YeelightBulb.hh"

namespace juno {

static const std::string yeelightMulticastAddr = "239.255.255.250";
static const u16 yeelightMulticastPort         = 1982u;

YeelightVendor::YeelightVendor(boost::asio::io_context& io) :
    m_io(io), m_socket(m_io),
    m_yeelightEndpoint(
      boost::asio::ip::make_address(yeelightMulticastAddr), yeelightMulticastPort
    ) {
    m_socket.open(boost::asio::ip::udp::v4());

    m_socket.set_option(boost::asio::ip::multicast::hops{ 3 });
    m_socket.set_option(boost::asio::ip::multicast::enable_loopback{ true });

    std::vector<std::string> headers = {
        "M-SEARCH * HTTP/1.1",
        fmt::format("HOST: {}:{}", yeelightMulticastAddr, yeelightMulticastPort),
        "MAN: \"ssdp:discover\"",
        "ST: wifi_bulb",
    };

    m_discoverMessage = boost::algorithm::join(headers, "\r\n");
    log::debug("Yeelight discover message: \n{}", m_discoverMessage);
}

kstd::Coro<void> YeelightVendor::scan() {
    log::info("Scanning for Yeelight devices, multicasting discover message");

    co_await m_socket.async_send_to(
      boost::asio::buffer(m_discoverMessage), m_yeelightEndpoint,
      boost::asio::use_awaitable
    );

    const auto scanDeadline = 5s;
    const auto ex           = co_await boost::asio::this_coro::executor;

    kstd::callLater(ex, scanDeadline, [&]() { m_socket.cancel(); });

    log::info("Waiting for devices to respond");
    std::array<char, 1024> buffer;
    while (true) {
        try {
            boost::asio::ip::udp::endpoint sender;
            const auto bytes = co_await m_socket.async_receive_from(
              boost::asio::buffer(buffer), sender, boost::asio::use_awaitable
            );
            co_await processNewDevice(std::string{ buffer.data(), bytes });
        } catch (const boost::system::system_error& e) {
            if (e.code() != boost::asio::error::operation_aborted)
                log::warn("Socket error: {}", e.what());
            break;
        }
    }
    log::info("Scanning finished, {} devices discovered", m_devices.size());
}

kstd::Coro<void> YeelightVendor::processNewDevice(const std::string& payload) {
    log::info("Processing new device");
    std::unordered_map<std::string, std::string> headers;

    for (const auto& line : kstd::split(payload, "\r\n")) {
        if (line.empty() || line.find("HTTP/") != line.npos) continue;

        const auto separator = line.find_first_of(":");
        if (separator == line.npos) {
            log::warn("Could not parse header: {}", line);
            continue;
        }
        const auto key   = line.substr(0, separator);
        const auto value = line.substr(separator + 2);
        headers[key]     = value;

        log::debug("Parsed header: '{}'='{}'", key, value);
    }

    static const auto requiredHeaders = { "model", "id" };

    for (const auto requiredHeader : requiredHeaders) {
        if (not headers.contains(requiredHeader)) {
            log::warn("Could not find '{}' header", requiredHeader);
            co_return;
        }
    }

    const auto id = std::stoull(headers.at("id"), 0, 16u);

    for (auto& device : m_devices) {
        if (static_cast<YeelightBulb*>(device.get())->yeelightId == id) {
            log::info("Device ({}) already stored, skipping", id);
            co_return;
        }
    }

    if (const auto model = headers["model"]; model == "strip6") {
        m_devices.push_back(co_await YeelightBulb::create(m_io, headers));
    } else {
        log::warn("Device model not supported: '{}'", model);
    }
}

Devices YeelightVendor::getDevices() const { return m_devices; }

}  // namespace juno
