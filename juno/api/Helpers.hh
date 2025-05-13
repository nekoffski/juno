#pragma once

#include <vector>

#include "juno.pb.h"
#include "devices/Device.hh"

namespace juno {

template <typename T>
std::vector<T> toVector(const ::google::protobuf::RepeatedPtrField<T>& r) {
    return std::vector<T>{ r.rbegin(), r.rend() };
}

}  // namespace juno
