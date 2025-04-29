#include "Socket.hh"

namespace juno {

kstd::Coro<kstd::UniquePtr<StreamSocket>> StreamSocket::create(
  boost::asio::io_context& io, const std::string& host, u16 port
) {
    auto socket = kstd::makeUnique<StreamSocket>(Tag{}, io, host, port);
    co_await socket->connect();
    co_return socket;
}

StreamSocket::StreamSocket(
  Tag, boost::asio::io_context& io, const std::string& host, u16 port
) :
    m_endpoint(boost::asio::ip::make_address(host), port),
    m_socket(io.get_executor()) {}

kstd::Coro<void> StreamSocket::write(const std::string& message) {
    log::debug("Sending message: {}", message);
    co_await boost::asio::async_write(
      m_socket, boost::asio::buffer(message), boost::asio::use_awaitable
    );
}

kstd::Coro<std::string> StreamSocket::read() {
    std::string buffer(1024u, '\0');
    const auto bytes = co_await m_socket.async_read_some(
      boost::asio::buffer(buffer), boost::asio::use_awaitable
    );
    log::debug("Got message: {}", buffer);
    co_return buffer.substr(0u, bytes);
}

kstd::Coro<void> StreamSocket::connect() {
    co_await m_socket.async_connect(m_endpoint, boost::asio::use_awaitable);
}

}  // namespace juno
