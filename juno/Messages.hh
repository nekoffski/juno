#pragma once

#include "Core.hh"
#include "devices/Device.hh"

namespace juno {

struct ListDevices {
    struct Request {};
    struct Response {
        Devices devices;
    };
};

}  // namespace juno
