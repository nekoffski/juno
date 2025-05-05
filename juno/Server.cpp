#include "Server.hh"

#include "api/GrpcApi.hh"
#include "devices/DeviceProxy.hh"
#include "scheduler/Scheduler.hh"

namespace juno {

Server::Server(const Config& config
) : m_signals(m_io, SIGTERM, SIGINT), m_config(config), m_messenger(m_io) {
    addService<GrpcApi>(m_io, m_messenger, m_config);
    addService<DeviceProxy>(m_io, m_messenger);
    addService<Scheduler>(m_io, m_messenger);

    m_signals.async_wait([&](const boost::system::error_code& ec, i32 signal) {
        if (!ec) {
            log::warn("Received signal: {} - requesting shut down", signal);
            shutdownServices();

            // FIXME - stop services gratefully instead of shutting down the engine
            m_io.stop();
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
    for (auto& service : m_services) service->shutdown();
}

}  // namespace juno
