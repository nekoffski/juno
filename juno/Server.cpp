#include "Server.hh"

#include "api/GrpcApi.hh"
#include "devices/DeviceProxy.hh"
#include "scheduler/Scheduler.hh"
#include "metrics/MetricService.hh"

namespace juno {

Server::Server() : m_signals(m_io, SIGTERM, SIGINT), m_messenger(m_io) {
    addService<GrpcApi>(m_io, m_messenger);
    addService<MetricService>(m_io, m_messenger);
    addService<Scheduler>(m_io, m_messenger);
    addService<DeviceProxy>(m_io, m_messenger);

    m_signals.async_wait([&](const boost::system::error_code& ec, i32 signal) {
        if (!ec) {
            log::warn("Received signal: {} - requesting shut down", signal);
            stopServices();
        }
    });
}

i32 Server::start() {
    startServices();
    m_io.run();
    return 0;
}

void Server::startServices() {
    for (auto& service : m_services) service->start();
}

void Server::stopServices() {
    for (auto& service : m_services) service->stop();
    for (auto& service : m_services) service->shutdown();
}

}  // namespace juno
