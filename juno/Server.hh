#pragma once

#include "Core.hh"
#include "Config.hh"

namespace juno {

class Server {
public:
    explicit Server(const Config& config, const kstd::FileSystem& fileSystem);

    i32 start();

private:
    Config m_config;
    const kstd::FileSystem& m_fileSystem;
};

}  // namespace juno
