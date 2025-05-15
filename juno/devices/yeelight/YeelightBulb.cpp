#include "YeelightBulb.hh"

#include "Core.hh"

namespace juno {

kstd::Coro<kstd::UniquePtr<YeelightBulb>> YeelightBulb::create(
  boost::asio::io_context& io,
  const std::unordered_map<std::string, std::string>& headers
) {
    const auto id  = std::stoi(headers.at("id"), 0, 16u);
    const auto loc = headers.at("Location");

    const auto colon     = loc.find_last_of(":");
    const auto addrStart = loc.find_last_of("/");
    const auto addr      = loc.substr(addrStart + 1, colon - addrStart - 1);
    const auto port      = std::stoull(loc.substr(colon + 1));

    log::info("Yeelight Bulb endpoint {}:{}", addr, port);

    co_return kstd::makeUnique<YeelightBulb>(
      Tag{}, co_await StreamSocket::create(io, addr, port), id
    );
}

YeelightBulb::YeelightBulb(
  Tag, kstd::UniquePtr<StreamSocket> socket, i32 yeelightId
) : m_socket(std::move(socket)), yeelightId(yeelightId) {}

const std::string& YeelightBulb::getName() const { return "somename"; }

kstd::Coro<void> YeelightBulb::toggle() { co_await request("toggle"); }

}  // namespace juno
