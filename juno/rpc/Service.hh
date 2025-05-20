#pragma once

#include <atomic>
#include <vector>
#include <functional>

#include <boost/asio/cancellation_signal.hpp>

#include <kstd/containers/FlatMap.hh>
#include <kstd/Concepts.hh>
#include <kstd/async/Core.hh>
#include <kstd/async/AsyncMessenger.hh>

#include "Core.hh"
#include "Queues.hh"

namespace juno {

class Service {
public:
    explicit Service(boost::asio::io_context& io);
    virtual ~Service() = default;

    virtual void start() = 0;
    virtual void shutdown() {}

    void stop() {
        m_isRunning = false;
        m_cancelationSignal.emit(boost::asio::cancellation_type::all);
    }

protected:
    template <typename Callback>
    requires kstd::Callable<Callback, kstd::Coro<void>>
    void spawn(Callback&& c) {
        kstd::spawn(
          m_io.get_executor(),
          [&, c = std::forward<Callback>(c)]() -> kstd::Coro<void> {
              try {
                  co_await c();
              } catch ([[maybe_unused]] boost::system::system_error& e) {
                  log::info("Coroutine ended with: {}", e.what());
              }
          }
        );
    }

    bool isRunning() const { return m_isRunning; }

    template <typename Callback>
    requires kstd::Callable<Callback>
    void onServiceCancel(Callback&& c) {
        m_cancelCallbacks.emplace_back(std::forward<Callback>(c));
        m_cancelationSignal.slot().assign([&]([[maybe_unused]] auto) {
            for (auto& c : m_cancelCallbacks) c();
        });
    }

    boost::asio::io_context& m_io;

    std::atomic_bool m_isRunning;
    boost::asio::cancellation_signal m_cancelationSignal;
    std::vector<std::function<void()>> m_cancelCallbacks;
};

template <typename Impl> class RpcService : public Service {
    using Callback = std::function<kstd::Coro<void>(kstd::AsyncMessage&)>;

public:
    explicit RpcService(
      boost::asio::io_context& io, Impl* impl, kstd::AsyncMessenger& messenger,
      const std::string& queueName
    ) :
        Service(io), m_impl(impl), m_queueName(queueName),
        m_mq(messenger.registerQueue(queueName)) {
        spawn([&]() -> kstd::Coro<void> { co_await startHandling(); });
    }

protected:
    template <typename Request>
    void registerCall(
      kstd::Coro<void> (Impl::*c)(kstd::AsyncMessage&, const Request&)
    ) {
        const auto& type = typeid(Request);
        log::expect(
          not m_handlers.has(type), "Handler '{}' already registered", type.name()
        );

        m_handlers.insert(
          type,
          [c, impl = m_impl](kstd::AsyncMessage& message) -> kstd::Coro<void> {
              co_await std::invoke(c, impl, message, *message.as<Request>());
          }
        );
    }

    kstd::AsyncMessenger::Queue& getMessageQueue() { return *m_mq; }

private:
    kstd::Coro<void> startHandling() {
        onServiceCancel([&]() { m_mq->cancel(); });

        while (isRunning()) {
            auto message           = co_await m_mq->wait();
            const auto messageType = message->getType();

            if (auto handler = m_handlers.get(messageType); not handler) {
                co_await message->template respond<Error>(
                  "'{}' - unhandled call '{}'", m_queueName, messageType.name()
                );
            } else {
                co_await std::invoke(*handler, *message);
            }
        }
    }

    Impl* m_impl;
    std::string m_queueName;
    kstd::AsyncMessenger::Queue* m_mq;
    kstd::DynamicFlatMap<std::type_index, Callback> m_handlers;
};
}  // namespace juno
