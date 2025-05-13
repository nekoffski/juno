
#include "Core.hh"
#include "Server.hh"

int main() {
    juno::log::init("juno");

    kstd::GlobalFileSystem fs;

    try {
        return juno::Server{}.start();
    } catch (const juno::Error& e) {
        juno::log::error("Unhandled exception: {} - {}", e.what(), e.where());
        return -1;
    } catch (const std::exception& e) {
        juno::log::error("Unhandled unknown exception: {}", e.what());
        return -2;
    }
}
