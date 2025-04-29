#pragma once

#include <nlohmann/json.hpp>

#include <kstd/async/Core.hh>
#include <kstd/memory/UniquePtr.hh>

#include "Core.hh"

namespace juno {

class StreamSocket {
protected:
    struct Tag {};

public:
    static kstd::Coro<kstd::UniquePtr<StreamSocket>> create(
      boost::asio::io_context& io, const std::string& host, u16 port
    );

    explicit StreamSocket(
      Tag, boost::asio::io_context& io, const std::string& host, u16 port
    );

    kstd::Coro<void> write(const std::string& message);
    kstd::Coro<std::string> read();

protected:
    kstd::Coro<void> connect();

private:
    boost::asio::ip::tcp::endpoint m_endpoint;
    boost::asio::ip::tcp::socket m_socket;
};

}  // namespace juno
