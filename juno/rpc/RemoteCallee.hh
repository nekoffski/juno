#pragma once

#include <unordered_map>
#include <functional>

#include <kstd/async/Core.hh>
#include <kstd/async/AsyncMessenger.hh>

#include "Core.hh"

namespace juno {

template <typename Impl> class RemoteCallee {
    using Callback = std::function<kstd::Coro<void>(kstd::AsyncMessage&)>;

public:
    explicit RemoteCallee(
      Impl* impl, kstd::AsyncMessenger& messenger, const std::string& queueName
    ) :
        m_impl(impl), m_queueName(queueName),
        m_mq(messenger.registerQueue(queueName)) {}

    kstd::Coro<void> startHandling() {
        while (true) {
            auto message           = co_await m_mq->wait();
            const auto messageType = message->getType();
            auto handler           = m_handlers.find(messageType);

            if (handler == m_handlers.end()) {
                co_await message->template respond<Error>(
                  "'{}' - unhandled call '{}'", m_queueName, messageType.name()
                );
            }
            co_await handler->second(*message);
        }
    }

    template <typename Request>
    void registerCall(
      kstd::Coro<void> (Impl::*c)(kstd::AsyncMessage&, const Request&)
    ) {
        const auto& type = typeid(Request);
        log::expect(
          not m_handlers.contains(type), "Handler '{}' already registered",
          type.name()
        );

        m_handlers[type] =
          [c, impl = m_impl](kstd::AsyncMessage& message) -> kstd::Coro<void> {
            co_await std::invoke(c, impl, message, *message.as<Request>());
        };
    }

private:
    Impl* m_impl;
    std::string m_queueName;
    kstd::AsyncMessenger::Queue* m_mq;

    std::unordered_map<std::type_index, Callback> m_handlers;
};

}  // namespace juno
