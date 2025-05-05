#include "Device.hh"

namespace juno {

Device::Type Bulb::getDeviceType() const { return Type::bulb; }

Device::Interface Bulb::getImplementedInterfaces() const {
    return Interface::togglable;
}

void Bulb::toProto(api::Device* device) const {
    device->set_uuid(uuid);
    device->set_type(api::BULB);
}

}  // namespace juno
