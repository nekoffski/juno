#include "GrpcApi.hh"

#include "messages/Queues.hh"

#include "proto/juno.pb.h"
#include "proto/juno.grpc.pb.h"

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

    service.addRequest<api::PingRequest, api::PongResponse>(
      &api::JunoService::AsyncService::RequestPing,
      [&](const api::PingRequest& req) -> ResponseWrapper<api::PongResponse> {
          const auto magic = req.magic();
          log::debug("Received ping request, magic: '{}'", magic);

          api::PongResponse res;
          res.set_magic(magic);
          co_return std::make_pair(res, grpc::Status::OK);
      }
    );
}

}  // namespace juno
