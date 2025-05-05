#pragma once

#include <string>
#include <vector>
#include <variant>
#include <functional>

#include "Core.hh"
#include "devices/Device.hh"

namespace juno {

struct GetDevices {
    struct Request {
        using Uuids  = std::vector<std::string>;
        using Filter = std::function<bool(Device&)>;

        std::variant<Uuids, Filter> criteria;
    };
    struct Response {
        Devices devices;
    };
};

}  // namespace juno
