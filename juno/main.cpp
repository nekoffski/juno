
#include "Core.hh"
#include "Config.hh"
#include "Server.hh"

int main(int argc, char** argv) {
    juno::log::init("juno");
    juno::log::expect(argc == 2, "Config path is required as input argument");

    kstd::GlobalFileSystem fs;

    try {
        auto config = juno::Config::fromFile(std::string{ argv[1] }, fs);
        return juno::Server{ config }.start();
    } catch (const juno::Error& e) {
        juno::log::error("Unhandled exception: {} - {}", e.what(), e.where());
        return -1;
    } catch (const std::exception& e) {
        juno::log::error("Unhandled unknown exception: {}", e.what());
        return -2;
    }
}
