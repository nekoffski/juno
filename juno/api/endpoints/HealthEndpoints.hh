#pragma once

#include <kstd/async/Core.hh>

#include "juno.grpc.pb.h"
#include "juno.pb.h"

namespace juno {

kstd::Coro<api::PongResponse> pingEndpoint(const api::PingRequest&);

}  // namespace juno
