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
        struct All {};

        using Uuids    = std::vector<std::string>;
        using Filter   = std::function<bool(Device&)>;
        using Criteria = std::variant<All, Uuids, Filter, Device::Interface>;

        Criteria criteria;
    };
    struct Response {
        Devices devices;
    };
};

struct RemoveJobs {
    struct Request {
        std::vector<std::string> uuids;
    };

    struct Response {
        std::vector<std::string> missingJobs;
    };
};

}  // namespace juno
