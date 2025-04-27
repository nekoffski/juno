#pragma once

#include <vector>

#include <kstd/async/Core.hh>
#include <kstd/memory/SharedPtr.hh>

#include "Device.hh"

namespace juno {

struct Vendor {
    virtual Devices getDevices() const = 0;
    virtual kstd::Coro<void> scan()    = 0;
};

}  // namespace juno
