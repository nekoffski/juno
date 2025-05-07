#pragma once

#include <kstd/async/Core.hh>

namespace juno {

struct Action {
    virtual ~Action()                  = default;
    virtual kstd::Coro<void> execute() = 0;
};

}  // namespace juno
