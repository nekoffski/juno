#include "Grpc.hh"

namespace juno {

grpc::StatusCode GrpcError::code() const { return m_code; }
grpc::Status GrpcError::status() const { return grpc::Status{ m_code, what() }; }

AsyncGrpcServer::Builder::Builder(
  AsyncGrpcServer& server, grpc::ServerBuilder& serverBuilder
) : m_server(server), m_serverBuilder(serverBuilder) {}

AsyncGrpcServer::AsyncGrpcServer(
  boost::asio::io_context& ctx, const Config& config
) : m_ctx(ctx), m_config(config) {}

AsyncGrpcServer::~AsyncGrpcServer() {
    if (m_server) m_server->Shutdown();
    for (auto& service : m_services) service->shutdown();
}

void AsyncGrpcServer::startAsync() {
    grpc::ServerBuilder serverBuilder;

    const auto addr = fmt::format("{}:{}", m_config.host, m_config.port);
    kstd::log::debug("GRPC listening address: {}", addr);
    serverBuilder.AddListeningPort(addr, grpc::InsecureServerCredentials());

    build(Builder{ *this, serverBuilder });
    m_server = serverBuilder.BuildAndStart();

    auto executor = m_ctx.get_executor();

    for (auto& service : m_services) {
        kstd::spawn(executor, [&]() -> kstd::Coro<void> {
            co_await service->start();
        });
    }
}

}  // namespace juno
