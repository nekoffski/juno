#pragma once

#include <kstd/async/AsyncMessenger.hh>

#include "devices/Device.hh"
#include "Messages.hh"

namespace juno::rpc {

kstd::Coro<std::vector<Device*>> getDevices(
  kstd::AsyncMessenger::Queue& mq,
  GetDevices::Request::Criteria criteria = GetDevices::Request::All{}
);

kstd::Coro<void> removeJobs(
  kstd::AsyncMessenger::Queue& mq, const std::vector<std::string>& uuids
);

kstd::Coro<std::string> addJob(
  kstd::AsyncMessenger::Queue& mq, const std::string& jobBody
);

}  // namespace juno::rpc
