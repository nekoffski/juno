#pragma once

#include "Service.hh"

#include <kstd/async/Core.hh>

#include "rpc/Messages.hh"
#include "rpc/MessageQueueDestination.hh"
#include "Metrics.hh"

namespace juno {

class MetricService : public Service, public MessageQueueDestination<MetricService> {
    struct Config {
        f32 lattitude;
        f32 longitude;
        std::string openWeatherApiKey;
    };

public:
    explicit MetricService(
      boost::asio::io_context& io, kstd::AsyncMessenger& messenger
    );

    void spawn() override;
    void shutdown() override;

private:
    Config readConfig() const;

    kstd::Coro<void> updateMetrics();
    kstd::Coro<Metrics::Weather> queryOpenWeather();

    boost::asio::io_context& m_io;
    Metrics m_metrics;
    Config m_conf;
};

}  // namespace juno
