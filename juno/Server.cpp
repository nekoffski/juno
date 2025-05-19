#include "Server.hh"

#include "api/GrpcApi.hh"
#include "devices/DeviceProxy.hh"
#include "scheduler/Scheduler.hh"
#include "metrics/MetricService.hh"

namespace juno {

Server::Server() : m_signals(m_io, SIGTERM, SIGINT), m_messenger(m_io) {
    addService<GrpcApi>(m_io, m_messenger);
    // addService<DeviceProxy>(m_io, m_messenger);
    // addService<Scheduler>(m_io, m_messenger);
    // addService<MetricService>(m_io, m_messenger);

    m_signals.async_wait([&](const boost::system::error_code& ec, i32 signal) {
        if (!ec) {
            log::warn("Received signal: {} - requesting shut down", signal);
            shutdownServices();
        }
    });
}

i32 Server::start() {
    spawnServices();
    m_io.run();
    return 0;
}

void Server::spawnServices() {
    for (auto& service : m_services) service->spawn();
}

void Server::shutdownServices() {
    for (auto& service : m_services) service->stop();
    for (auto& service : m_services) service->shutdown();
}

}  // namespace juno
