#pragma once

#include <vector>

#include <kstd/memory/SharedPtr.hh>
#include <kstd/RTTI.hh>
#include <kstd/Id.hh>

#include "Core.hh"

namespace juno {

struct Device {
    enum class Type : u16 { undefined, bulb, fan };

    explicit Device() : uuid(kstd::generateUuid()) {}
    virtual ~Device() = default;

    virtual Type getDeviceType() const         = 0;
    virtual const std::string& getName() const = 0;

    const std::string uuid;
};

using Devices = std::vector<kstd::SharedPtr<Device>>;

struct Bulb : Device {
    Type getDeviceType() const override;
};

}  // namespace juno
