#pragma once

#include <unordered_map>

#include <kstd/async/Core.hh>

#include "Core.hh"

namespace juno {

struct HttpRequest {
    std::string host;
    std::string path;
};

struct HttpResponse {
    u32 code;
    std::string body;
};

kstd::Coro<HttpResponse> httpRequest(const HttpRequest& r);
kstd::Coro<HttpResponse> httpsRequest(const HttpRequest& r);

}  // namespace juno
