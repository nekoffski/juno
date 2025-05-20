#pragma once

#include <vector>
#include <concepts>

#include <kstd/memory/UniquePtr.hh>
#include <kstd/async/Core.hh>
#include <kstd/async/AsyncMessenger.hh>

#include "Core.hh"

#include "rpc/Service.hh"

namespace juno {

class Server {
public:
    explicit Server();

    i32 start();

private:
    template <typename T, typename... Args>
    requires(std::derived_from<T, Service> && std::constructible_from<T, Args...>)
    void addService(Args&&... args) {
        m_services.push_back(kstd::makeUnique<T>(std::forward<Args>(args)...));
    }

    void startServices();
    void stopServices();

    boost::asio::io_context m_io;
    boost::asio::signal_set m_signals;

    kstd::AsyncMessenger m_messenger;

    std::vector<kstd::UniquePtr<Service>> m_services;
};

}  // namespace juno
