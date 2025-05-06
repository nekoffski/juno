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

bool Device::implements(Interface interfaces) const {
    return static_cast<bool>(getImplementedInterfaces() & interfaces);
}

}  // namespace juno
