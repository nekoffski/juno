#pragma once

#include "Core.hh"

namespace juno {

struct Metrics {
    struct Weather {
        f32 temp;
        f32 feelsLike;
        f32 minTemp;
        f32 maxTemp;
        u32 pressure;
        u32 humidity;
        f32 windSpeed;
    } weather;

    void log();
};

}  // namespace juno
