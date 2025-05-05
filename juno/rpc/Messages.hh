#pragma once

#include <string>
#include <vector>

#include "Core.hh"
#include "devices/Device.hh"

namespace juno {

struct ListDevices {
    struct Request {
        std::vector<std::string> uuids;
    };
    struct Response {
        Devices devices;
    };
};

}  // namespace juno
