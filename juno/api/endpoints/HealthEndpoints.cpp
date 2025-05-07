#include "HealthEndpoints.hh"

#include "Core.hh"

namespace juno {

kstd::Coro<api::PongResponse> pingEndpoint(const api::PingRequest& req) {
    const auto magic = req.magic();
    log::trace("Received ping request, magic: '{}'", magic);

    api::PongResponse res;
    res.set_magic(magic);
    co_return res;
}

}  // namespace juno
