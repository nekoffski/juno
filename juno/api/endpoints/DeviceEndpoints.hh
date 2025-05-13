#pragma once

#include <kstd/async/AsyncMessenger.hh>

#include "juno.grpc.pb.h"
#include "juno.pb.h"

namespace juno {

kstd::Coro<api::ListDevicesResponse>
  listDevicesEndpoint(kstd::AsyncMessenger::Queue&);

kstd::Coro<api::AckResponse>
  toggleDevicesEndpoint(kstd::AsyncMessenger::Queue&, const api::ToggleDevicesRequest&);

}  // namespace juno
