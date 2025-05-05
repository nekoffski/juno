#pragma once

#include <kstd/async/AsyncMessenger.hh>

#include "devices/Device.hh"
#include "Messages.hh"

namespace juno::rpc {

// FIXME: separate overload for getting all devices
kstd::Coro<Devices> getDevices(
  kstd::AsyncMessenger::Queue& mq, const GetDevices::Request::Uuids& uuids = {}
);

kstd::Coro<Devices> getDevices(
  kstd::AsyncMessenger::Queue& mq, const GetDevices::Request::Filter& filter
);

}  // namespace juno::rpc
