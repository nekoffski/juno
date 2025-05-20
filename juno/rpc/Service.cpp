#include "Service.hh"

namespace juno {

Service::Service(boost::asio::io_context& io) : m_io(io), m_isRunning(true) {}

}  // namespace juno
