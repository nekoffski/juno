#pragma once

#include <kstd/async/AsyncMessenger.hh>

#include "devices/Device.hh"

namespace juno::rpc {

kstd::Coro<Devices> getDevices(
  kstd::AsyncMessenger::Queue& mq, const std::vector<std::string>& uuids = {}
);

}
