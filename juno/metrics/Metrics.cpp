#include "Metrics.hh"

#include "Core.hh"

#include <boost/date_time/posix_time/posix_time.hpp>

namespace juno {

void Metrics::log() {
    log::debug(
      "Weather: temp={}, feels_like={}, min temp={}, max temp={}, pressure={}, humidity={},"
      " wind speed={}, sunrise='{}', sunset='{}'",
      weather.temp, weather.feelsLike, weather.minTemp, weather.maxTemp,
      weather.pressure, weather.humidity, weather.windSpeed,
      boost::posix_time::to_simple_string(
        boost::posix_time::from_time_t(weather.sunrise)
      ),
      boost::posix_time::to_simple_string(
        boost::posix_time::from_time_t(weather.sunset)
      )
    );
}

}  // namespace juno