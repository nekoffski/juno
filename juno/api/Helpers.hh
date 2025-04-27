#pragma once

#include "proto/juno.pb.h"
#include "devices/Device.hh"

namespace juno {

void toProto(const Device* in, api::Device* out);

}
