#pragma once

#include <vector>
#include <functional>

#include <kstd/Scope.hh>

#include <grpcpp/grpcpp.h>

#include <kstd/Core.hh>
#include <kstd/memory/UniquePtr.hh>
#include <kstd/async/Utils.hh>
#include <kstd/async/Core.hh>
#include <kstd/Log.hh>

namespace juno {

template <typename Response>
using ResponseWrapper = kstd::Coro<std::pair<Response, grpc::Status>>;

namespace details {

// clang-format off
template <typename Service, typename Request, typename Response>
using InternalHandler = void (Service::*)(
    grpc::ServerContext*, Request*,
    grpc::ServerAsyncResponseWriter<Response>*,
    grpc::CompletionQueue*, grpc::ServerCompletionQueue*, void* 
);
// clang-format on

struct CallData {
    enum class Status : kstd::u8 { processing, finishing };
    virtual ~CallData() = default;

    virtual kstd::Coro<void> proceed() = 0;
};

template <typename Service, typename Request, typename Response>
class CallDataBase : public CallData {
    using Callback = std::function<ResponseWrapper<Response>(const Request&)>;

public:
    template <typename Callback>
    requires kstd::Callable<Callback, ResponseWrapper<Response>, const Request&>
    explicit CallDataBase(
      grpc::ServerCompletionQueue* cq, Service* service,
      InternalHandler<Service, Request, Response> internalHandler,
      Callback&& callback
    ) :
        m_status(Status::processing), m_cq(cq), m_service(service),
        m_responder(&m_ctx), m_internalHandler(internalHandler),
        m_callback(std::forward<Callback>(callback)) {
        // init async process
        (m_service->*m_internalHandler)(
          &m_ctx, &m_request, &m_responder, m_cq, m_cq, this
        );
    }

    kstd::Coro<void> proceed() {
        if (m_status == Status::processing) {
            new CallDataBase{ m_cq, m_service, m_internalHandler, m_callback };

            const auto [response, status] = co_await m_callback(m_request);
            m_responder.Finish(response, status, this);

            m_status = Status::finishing;
        } else {
            delete this;
        }
    }

private:
    Status m_status;

    grpc::ServerContext m_ctx;
    grpc::ServerCompletionQueue* m_cq;

    Service* m_service;
    Request m_request;
    grpc::ServerAsyncResponseWriter<Response> m_responder;
    InternalHandler<Service, Request, Response> m_internalHandler;
    Callback m_callback;
};

}  // namespace details

class AsyncGrpcServer {
public:
    class Service {
    public:
        virtual ~Service() = default;

        virtual kstd::Coro<void> start() = 0;
        virtual void shutdown()          = 0;
    };

    template <typename Impl> class ServiceBase : public Service {
        using Initializer = std::function<void()>;

    public:
        explicit ServiceBase(std::unique_ptr<grpc::ServerCompletionQueue> cq
        ) : m_cq(std::move(cq)) {}

        kstd::Coro<void> start() override {
            for (auto& initializer : m_initializers) initializer();

            void* tag = nullptr;
            bool ok   = false;

            using std::chrono::system_clock;

            const auto waitDeadline = std::chrono::nanoseconds(10);
            const auto pollInterval = std::chrono::milliseconds(10);

            while (true) {
                const auto deadline = system_clock::now() + waitDeadline;
                const auto status   = m_cq->AsyncNext(&tag, &ok, deadline);

                if (status == grpc::CompletionQueue::GOT_EVENT) {
                    kstd::log::expect(ok, "GRPC completion queue polling failed");
                    co_await static_cast<details::CallData*>(tag)->proceed();
                } else if (status == grpc::CompletionQueue::TIMEOUT) {
                    co_await kstd::asyncSleep(pollInterval);
                } else {
                    break;
                }
            }
        }

        void addInitializer(Initializer&& initializer) {
            m_initializers.push_back(std::move(initializer));
        }

        void shutdown() override { m_cq->Shutdown(); }

        Impl* getImpl() { return &m_impl; }

        grpc::ServerCompletionQueue* getCompletionQueue() { return m_cq.get(); }

    private:
        std::vector<Initializer> m_initializers;
        Impl m_impl;
        std::unique_ptr<grpc::ServerCompletionQueue> m_cq;
    };

    class Builder {
    public:
        explicit Builder(
          AsyncGrpcServer& server, grpc::ServerBuilder& serverBuilder
        );

        template <typename T> class ServiceBuilder {
        public:
            explicit ServiceBuilder(ServiceBase<T>* service) : m_service(service) {}

            template <typename Request, typename Response, typename Callback>
            requires kstd::Callable<
              Callback, ResponseWrapper<Response>, const Request&>
            ServiceBuilder& addRequest(
              details::InternalHandler<T, Request, Response> rawHandler,
              Callback&& callback
            ) {
                m_service->addInitializer(
                  [&, service = m_service, rawHandler = rawHandler,
                   callback = std::forward<Callback>(callback)]() mutable {
                      new details::CallDataBase<T, Request, Response>{
                          service->getCompletionQueue(), service->getImpl(),
                          rawHandler, std::forward<Callback>(callback)
                      };
                  }
                );
                return *this;
            }

        private:
            ServiceBase<T>* m_service;
        };

        template <typename T> ServiceBuilder<T> addService() {
            auto service =
              kstd::makeUnique<ServiceBase<T>>(m_serverBuilder.AddCompletionQueue());
            m_serverBuilder.RegisterService(service->getImpl());
            ServiceBuilder<T> builder{ service.get() };
            m_server.m_services.push_back(std::move(service));
            return builder;
        }

    private:
        AsyncGrpcServer& m_server;
        grpc::ServerBuilder& m_serverBuilder;
    };

    struct Config {
        std::string host;
        kstd::u16 port;
    };

    explicit AsyncGrpcServer(boost::asio::io_context& ctx, const Config& config);
    virtual ~AsyncGrpcServer();

    void startAsync();

private:
    virtual void build(Builder&&) = 0;

    boost::asio::io_context& m_ctx;
    Config m_config;

    std::unique_ptr<grpc::Server> m_server;
    std::vector<kstd::UniquePtr<Service>> m_services;
};

}  // namespace juno
