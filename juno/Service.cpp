#include "Service.hh"

namespace juno {

Service::Service() : m_isRunning(true) {}

void Service::stop() { m_isRunning = false; }

bool Service::isRunning() const { return m_isRunning; }

}  // namespace juno
