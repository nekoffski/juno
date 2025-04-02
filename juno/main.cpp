
#include <kstd/Log.hh>

int main() {
    kstd::log::init("Juno-Server");
    kstd::log::debug("Hello world!");
    return 0;
}
