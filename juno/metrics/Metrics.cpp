#include "Metrics.hh"

#include "Core.hh"

namespace juno {

void Metrics::log() {
    log::debug(
      "Weather: temp={}, feels_like={}, min temp={}, max temp={}, pressure={}, humidity={}, wind speed={}",
      weather.temp, weather.feelsLike, weather.minTemp, weather.maxTemp,
      weather.pressure, weather.humidity, weather.windSpeed
    );
}

}  // namespace juno