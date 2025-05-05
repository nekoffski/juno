#pragma once

#include <kstd/async/AsyncMessenger.hh>

#include "proto/juno.grpc.pb.h"
#include "proto/juno.pb.h"

namespace juno {

kstd::Coro<api::PongResponse> pingEndpoint(const api::PingRequest&);

kstd::Coro<api::ListDevicesResponse>
  listDevicesEndpoint(kstd::AsyncMessenger::Queue&);

kstd::Coro<api::AckResponse>
  toggleDevicesEndpoint(kstd::AsyncMessenger::Queue&, const api::ToggleDevicesRequest&);

}  // namespace juno
