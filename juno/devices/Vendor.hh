#pragma once

#include <vector>

#include <kstd/async/Core.hh>
#include <kstd/memory/SharedPtr.hh>

#include "Device.hh"

namespace juno {

struct Vendor {
    virtual kstd::Coro<std::vector<Device*>> scan() = 0;
};

}  // namespace juno
