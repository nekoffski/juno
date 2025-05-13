#pragma once

#include <kstd/async/AsyncMessenger.hh>

#include "juno.grpc.pb.h"
#include "juno.pb.h"

namespace juno {

kstd::Coro<api::AddJobResponse>
  addJobEndpoint(kstd::AsyncMessenger::Queue&, const api::AddJobRequest&);

kstd::Coro<api::AckResponse>
  removeJobsEndpoint(kstd::AsyncMessenger::Queue&, const api::RemoveJobsRequest&);

}  // namespace juno
