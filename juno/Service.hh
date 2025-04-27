#pragma once

#include "Queues.hh"

namespace juno {

struct Service {
    virtual ~Service() = default;

    virtual void spawn()    = 0;
    virtual void shutdown() = 0;
};

}  // namespace juno
