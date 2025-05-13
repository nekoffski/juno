#include "GrpcApi.hh"

#include "juno.pb.h"
#include "juno.grpc.pb.h"

#include "endpoints/HealthEndpoints.hh"
#include "endpoints/DeviceEndpoints.hh"
#include "endpoints/SchedulerEndpoints.hh"

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
    m_mq(messenger.registerQueue(GRPC_API_QUEUE)) {}

void GrpcApi::spawn() { startAsync(); }

void GrpcApi::shutdown() {
    // FIXME
}

void GrpcApi::build(Builder&& builder) {
    builder.addService<api::HealthService::AsyncService>()
      .addRequest<api::PingRequest, api::PongResponse>(
        &api::HealthService::AsyncService::RequestPing,
        [&](const auto& req) -> kstd::Coro<api::PongResponse> {
            co_return co_await pingEndpoint(req);
        }
      );

    builder.addService<api::DeviceService::AsyncService>()
      .addRequest<api::ListDevicesRequest, api::ListDevicesResponse>(
        &api::DeviceService::AsyncService::RequestList,
        [&]([[maybe_unused]] const auto&) -> kstd::Coro<api::ListDevicesResponse> {
            co_return co_await listDevicesEndpoint(*m_mq);
        }
      )
      .addRequest<api::ToggleDevicesRequest, api::AckResponse>(
        &api::DeviceService::AsyncService::RequestToggle,
        [&](const auto& req) -> kstd::Coro<api::AckResponse> {
            co_return co_await toggleDevicesEndpoint(*m_mq, req);
        }
      );

    builder.addService<api::SchedulerService::AsyncService>()
      .addRequest<api::AddJobRequest, api::AddJobResponse>(
        &api::SchedulerService::AsyncService::RequestAddJob,
        [&](const auto& req) -> kstd::Coro<api::AddJobResponse> {
            co_return co_await addJobEndpoint(*m_mq, req);
        }
      )
      .addRequest<api::RemoveJobsRequest, api::AckResponse>(
        &api::SchedulerService::AsyncService::RequestRemoveJobs,
        [&](const auto& req) -> kstd::Coro<api::AckResponse> {
            co_return co_await removeJobsEndpoint(*m_mq, req);
        }
      );
}

}  // namespace juno
