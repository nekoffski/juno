#include "GrpcApi.hh"

#include "proto/juno.pb.h"
#include "proto/juno.grpc.pb.h"

#include "Endpoints.hh"

namespace juno {

GrpcApi::GrpcApi(
  boost::asio::io_context& io, kstd::AsyncMessenger& messenger,
  const juno::Config& cfg
) :
    AsyncGrpcServer(
      io,
      Config{
        .host = cfg.grpcApiHost,
        .port = cfg.grpcApiPort,
      }
    ),
    m_messageQueue(messenger.registerQueue(GRPC_API_QUEUE)) {}

void GrpcApi::spawn() { startAsync(); }

void GrpcApi::shutdown() {
    // FIXME
}

void GrpcApi::build(Builder&& builder) {
    auto&& service = builder.addService<api::JunoService::AsyncService>();
    using Service  = api::JunoService::AsyncService;

    service
      .addRequest<api::PingRequest, api::PongResponse>(
        &Service::RequestPing,
        [&](const auto& req) -> kstd::Coro<api::PongResponse> {
            co_return (co_await pingEndpoint(req));
        }
      )
      .addRequest<api::ListDevicesRequest, api::ListDevicesResponse>(
        &Service::RequestListDevices,
        [&]([[maybe_unused]] const auto&) -> kstd::Coro<api::ListDevicesResponse> {
            co_return (co_await listDevicesEndpoint(m_messageQueue));
        }
      );
}

}  // namespace juno
