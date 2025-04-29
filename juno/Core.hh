#pragma once

#include <stdexcept>

#include <kstd/Core.hh>
#include <kstd/Log.hh>
#include <kstd/FileSystem.hh>

#include <limits>
#include <fmt/core.h>

namespace juno {

using namespace std::chrono_literals;

namespace log {
using namespace kstd::log;
}

using kstd::i16;
using kstd::i32;
using kstd::i64;
using kstd::i8;

using kstd::u16;
using kstd::u32;
using kstd::u64;
using kstd::u8;

using kstd::f32;
using kstd::f64;

class Error : public std::runtime_error {
public:
    template <typename... Args>
    explicit Error(kstd::log::details::FormatWithLocation format, Args&&... args) :
        runtime_error(
          fmt::format(fmt::runtime(format.fmt), std::forward<Args>(args)...)
        ),
        m_source(format.loc) {}

    std::string where() const;

private:
    spdlog::source_loc m_source;
};

template <u64 Max> class SequenceGenerator {
    static_assert(Max < std::numeric_limits<u64>::max());

public:
    SequenceGenerator() : m_current(0u) {}

    u64 get() {
        m_current %= Max;
        return m_current++;
    }

private:
    u64 m_current;
};

}  // namespace juno
