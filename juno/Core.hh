#pragma once

#include <stdexcept>

#include <kstd/Core.hh>
#include <kstd/Log.hh>
#include <kstd/FileSystem.hh>
#include <kstd/Error.hh>

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

class Error : public kstd::Error {
public:
    enum class Code { unspecified = 0, notFound, invalidArgument };

    template <typename... Args>
    explicit Error(kstd::log::details::FormatWithLocation format, Args&&... args) :
        kstd::Error(std::move(format), std::forward<Args>(args)...),
        m_code(Code::unspecified) {}

    template <typename... Args>
    explicit Error(
      Code code, kstd::log::details::FormatWithLocation format, Args&&... args
    ) : kstd::Error(std::move(format), std::forward<Args>(args)...), m_code(code) {}

    Code code() const;

private:
    Code m_code;
};

}  // namespace juno
