#pragma once

#include <kstd/async/AsyncMessenger.hh>

#include "devices/Device.hh"
#include "Messages.hh"

namespace juno::rpc {

kstd::Coro<Devices> getDevices(
  kstd::AsyncMessenger::Queue& mq,
  GetDevices::Request::Criteria criteria = GetDevices::Request::All{}
);

kstd::Coro<void> removeJobs(
  kstd::AsyncMessenger::Queue& mq, const std::vector<std::string>& uuids
);

}  // namespace juno::rpc
