#pragma once

#include <kstd/async/Core.hh>
#include <kstd/memory/UniquePtr.hh>
#include <kstd/Id.hh>

#include <nlohmann/json.hpp>

#include "devices/Device.hh"
#include "net/Socket.hh"

namespace juno {

class YeelightBulb : public Bulb {
    struct Tag {};

public:
    static kstd::Coro<kstd::UniquePtr<YeelightBulb>> create(
      boost::asio::io_context& io,
      const std::unordered_map<std::string, std::string>& headers
    );
    explicit YeelightBulb(Tag, kstd::UniquePtr<StreamSocket> socket, i32 yeelightId);

    const std::string& getName() const override;
    kstd::Coro<void> toggle() override;

private:
    template <typename... Args>
    kstd::Coro<u64> request(const std::string& method, Args&&... args) {
        const auto requestId = m_idGenerator.get();
        nlohmann::json req;
        req["id"]     = requestId;
        req["method"] = method;
        req["params"] = nlohmann::json::array();

        (req["params"].push_back(args), ...);

        co_await m_socket->write(nlohmann::to_string(req) + "\r\n");
        co_return requestId;
    }

    kstd::SequenceGenerator<0u, 128u> m_idGenerator;
    kstd::UniquePtr<StreamSocket> m_socket;

public:
    const u64 yeelightId;
};

}  // namespace juno
