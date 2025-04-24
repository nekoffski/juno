#pragma once

#include <string>

#include "Core.hh"

namespace juno {

struct Config {
    static Config fromFile(const std::string& path, const kstd::FileSystem& fs);

    std::string grpcApiHost;
    u16 grpcApiPort;
};

}  // namespace juno
