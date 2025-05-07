#pragma once

#include <kstd/async/Core.hh>

#include "proto/juno.grpc.pb.h"
#include "proto/juno.pb.h"

namespace juno {

kstd::Coro<api::PongResponse> pingEndpoint(const api::PingRequest&);

}  // namespace juno
