#pragma once

#include <atomic>

#include "rpc/Queues.hh"

namespace juno {

class Service {
public:
    explicit Service();
    virtual ~Service() = default;

    virtual void spawn() = 0;
    virtual void shutdown() {}

    void stop();
    bool isRunning() const;

private:
    std::atomic_bool m_isRunning;
};

}  // namespace juno
