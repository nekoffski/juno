#include "MetricService.hh"

#include <kstd/async/Utils.hh>
#include <kstd/Env.hh>

#include "rpc/Queues.hh"
#include "net/Http.hh"

namespace juno {

MetricService::MetricService(
  boost::asio::io_context& io, kstd::AsyncMessenger& messenger
) :
    RpcService(io, this, messenger, METRIC_SERVICE_QUEUE), m_io(io),
    m_conf(readConfig()) {}

void MetricService::start() {
    log::warn("Metric service spawning");

    static const auto updateInterval = 10s;

    spawn([&]() -> kstd::Coro<void> {
        kstd::AsyncTimer timer{ co_await boost::asio::this_coro::executor };

        onServiceCancel([&]() { timer.cancel(); });

        while (isRunning()) {
            co_await updateMetrics();
            co_await timer.sleep(updateInterval);
        }
    });
}

kstd::Coro<void> MetricService::updateMetrics() {
    m_metrics.weather = co_await queryOpenWeather();

    log::debug("System metrics updated");
    m_metrics.log();
}

kstd::Coro<Metrics::Weather> MetricService::queryOpenWeather() {
    const auto response = co_await httpsRequest({
      .host = "api.openweathermap.org",
      .path = fmt::format(
        "/data/2.5/weather?units=metric&lat={}&lon={}&appid={}", m_conf.lattitude,
        m_conf.longitude, m_conf.openWeatherApiKey
      ),
    });
    if (response.code != 200) {
        // TODO: handle error
        co_return Metrics::Weather{};
    }

    const auto body = response.toJson();
    log::debug("open weather response: {}", response.body);

    const auto timezone = body["timezone"].get<i64>();

    co_return Metrics::Weather{
        .temp      = body["main"]["temp"].get<f32>(),
        .feelsLike = body["main"]["feels_like"].get<f32>(),
        .minTemp   = body["main"]["temp_min"].get<f32>(),
        .maxTemp   = body["main"]["temp_max"].get<f32>(),
        .pressure  = body["main"]["pressure"].get<u32>(),
        .humidity  = body["main"]["humidity"].get<u32>(),
        .windSpeed = body["wind"]["speed"].get<f32>(),
        .sunrise   = body["sys"]["sunrise"].get<i64>() + timezone,
        .sunset    = body["sys"]["sunset"].get<i64>() + timezone,

    };
}

MetricService::Config MetricService::readConfig() const {
    const auto cracowLattitude = 50.049683f;
    const auto cracowLongitude = 19.944544f;

    auto lattitude = kstd::getEnv<f32>("JUNO_LATTITUDE").value_or(cracowLattitude);
    auto longitude = kstd::getEnv<f32>("JUNO_LONGITUDE").value_or(cracowLongitude);
    auto openWeatherApiKey = kstd::getEnv("JUNO_OPEN_WEATHER_API_KEY");

    log::expect(
      openWeatherApiKey.has_value(), "JUNO_OPEN_WEATHER_API_KEY is not set"
    );

    log::info("System location: lat={}, lon={}", lattitude, longitude);

    return Config{
        .lattitude         = lattitude,
        .longitude         = longitude,
        .openWeatherApiKey = *openWeatherApiKey,
    };
}

void MetricService::shutdown() {}

}  // namespace juno
