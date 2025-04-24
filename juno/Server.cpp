#include "Server.hh"

namespace juno {

Server::Server(const Config& config, const kstd::FileSystem& fileSystem) :
    m_config(config), m_fileSystem(fileSystem) {}

i32 Server::start() { return 0; }

}  // namespace juno
