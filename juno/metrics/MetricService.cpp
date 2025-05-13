#include "MetricService.hh"

#include "rpc/Queues.hh"
#include "net/Http.hh"

#include <boost/beast/core.hpp>
#include <boost/beast/http.hpp>
#include <boost/beast/version.hpp>
#include <boost/beast/ssl.hpp>
#include <boost/asio/ssl.hpp>

namespace juno {

MetricService::MetricService(
  boost::asio::io_context& io, kstd::AsyncMessenger& messenger
) : MessageQueueDestination(this, messenger, METRIC_SERVICE_QUEUE), m_io(io) {
    boost::asio::ssl::context sslContext{ boost::asio::ssl::context::sslv23_client };
    sslContext.set_default_verify_paths();
}

void MetricService::spawn() {
    log::warn("Metric service spawning");

    kstd::spawn(m_io.get_executor(), [&]() -> kstd::Coro<void> {
        const auto lattitude = 50.049683f;
        const auto longitude = 19.944544f;

        const auto apiKey = std::getenv("JUNO_OPEN_WEATHER_API_KEY");

        auto response = co_await httpsRequest({
          .host = "api.openweathermap.org",
          .path = fmt::format(
            "/data/2.5/weather?lat={}&lon={}&appid={}", lattitude, longitude, apiKey
          ),
        });
        log::warn("kcz: {} / {}", response.code, response.body);
    });
}

void MetricService::shutdown() {}

}  // namespace juno
