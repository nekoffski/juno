#pragma once

#include <ctime>
#include <chrono>

#include "Core.hh"

namespace juno {

struct Metrics {
    using TimePoint = std::chrono::system_clock::time_point;

    struct Weather {
        f32 temp;
        f32 feelsLike;
        f32 minTemp;
        f32 maxTemp;
        u32 pressure;
        u32 humidity;
        f32 windSpeed;
        std::time_t sunrise;
        std::time_t sunset;
    } weather;

    void log();
};

}  // namespace juno
