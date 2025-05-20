#include "GrpcApi.hh"

#include <kstd/Env.hh>

#include "juno.pb.h"
#include "juno.grpc.pb.h"

#include "endpoints/HealthEndpoints.hh"
#include "endpoints/DeviceEndpoints.hh"
#include "endpoints/SchedulerEndpoints.hh"

namespace juno {

const auto grpcApiHost = kstd::getEnv("JUNO_GRPC_API_HOST").value_or("0.0.0.0");
const auto grpcApiPort = kstd::getEnv<u64>("JUNO_GRPC_API_PORT").value_or(8888);

GrpcApi::GrpcApi(boost::asio::io_context& io, kstd::AsyncMessenger& messenger) :
    RpcService(io, this, messenger, GRPC_API_QUEUE),
    AsyncGrpcServer(
      io,
      Config{
        .host = grpcApiHost,
        .port = static_cast<u16>(grpcApiPort),
      }
    ) {}

void GrpcApi::start() { startAsync(); }

void GrpcApi::shutdown() { AsyncGrpcServer::stop(); }

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
            co_return co_await listDevicesEndpoint(getMessageQueue());
        }
      )
      .addRequest<api::ToggleDevicesRequest, api::AckResponse>(
        &api::DeviceService::AsyncService::RequestToggle,
        [&](const auto& req) -> kstd::Coro<api::AckResponse> {
            co_return co_await toggleDevicesEndpoint(getMessageQueue(), req);
        }
      );

    builder.addService<api::SchedulerService::AsyncService>()
      .addRequest<api::AddJobRequest, api::AddJobResponse>(
        &api::SchedulerService::AsyncService::RequestAddJob,
        [&](const auto& req) -> kstd::Coro<api::AddJobResponse> {
            co_return co_await addJobEndpoint(getMessageQueue(), req);
        }
      )
      .addRequest<api::RemoveJobsRequest, api::AckResponse>(
        &api::SchedulerService::AsyncService::RequestRemoveJobs,
        [&](const auto& req) -> kstd::Coro<api::AckResponse> {
            co_return co_await removeJobsEndpoint(getMessageQueue(), req);
        }
      );
}

}  // namespace juno
