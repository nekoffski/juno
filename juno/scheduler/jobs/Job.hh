#pragma once

#include <kstd/async/Core.hh>

#include "Core.hh"

namespace juno {

struct Job {
    enum class Type { oneshot, repeated, rule };
};

}  // namespace juno
