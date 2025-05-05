#pragma once

#include <vector>

#include <kstd/memory/SharedPtr.hh>
#include <kstd/RTTI.hh>
#include <kstd/Id.hh>
#include <kstd/async/Core.hh>

#include "Core.hh"
#include "proto/juno.pb.h"

namespace juno {

struct Device {
    enum class Type : u16 { undefined, bulb, fan };

    explicit Device() : uuid(kstd::generateUuid()) {}
    virtual ~Device() = default;

    virtual void toProto(api::Device*) const = 0;

    virtual Type getDeviceType() const         = 0;
    virtual const std::string& getName() const = 0;

    const std::string uuid;
};

using Devices = std::vector<kstd::SharedPtr<Device>>;

struct Togglable {
    virtual kstd::Coro<void> toggle() = 0;
};

class Bulb : public Device, public Togglable {
public:
    Type getDeviceType() const override;

private:
    void toProto(api::Device*) const override;
};

}  // namespace juno
