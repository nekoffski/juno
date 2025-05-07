#pragma once

#include <vector>

#include <kstd/memory/SharedPtr.hh>
#include <kstd/RTTI.hh>
#include <kstd/Id.hh>
#include <kstd/async/Core.hh>
#include <kstd/Enum.hh>

#include "Core.hh"
#include "proto/juno.pb.h"

namespace juno {

struct Device : kstd::WithUuid {
    enum class Type : u16 { undefined, bulb, fan };
    enum class Interface : u16 {
        togglable = 0x1,
    };

    virtual ~Device() = default;

    virtual void toProto(api::Device*) const = 0;

    bool implements(Interface interfaces) const;

    virtual Type getDeviceType() const                 = 0;
    virtual Interface getImplementedInterfaces() const = 0;
    virtual const std::string& getName() const         = 0;
};

using Devices = std::vector<kstd::SharedPtr<Device>>;

struct Togglable {
    virtual kstd::Coro<void> toggle() = 0;
};

class Bulb : public Device, public Togglable {
public:
    Type getDeviceType() const override;
    Interface getImplementedInterfaces() const override;

private:
    void toProto(api::Device*) const override;
};

constexpr void enableBitOperations(juno::Device::Interface);

}  // namespace juno
