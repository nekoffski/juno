#include "Http.hh"

#include <boost/beast/core.hpp>
#include <boost/beast/http.hpp>
#include <boost/beast/version.hpp>
#include <boost/beast/ssl.hpp>
#include <boost/asio/ssl.hpp>

namespace juno {

template <typename Stream>
static kstd::Coro<HttpResponse> requestImpl(const HttpRequest& r, Stream& s) {
    boost::beast::http::request<boost::beast::http::string_body> request{
        boost::beast::http::verb::get, r.path, 11
    };
    request.set(boost::beast::http::field::host, r.host);
    request.set(boost::beast::http::field::user_agent, BOOST_BEAST_VERSION_STRING);

    co_await boost::beast::http::async_write(s, request, boost::asio::use_awaitable);

    boost::beast::flat_buffer buffer;
    boost::beast::http::response<boost::beast::http::dynamic_body> response;

    co_await boost::beast::http::async_read(
      s, buffer, response, boost::asio::use_awaitable
    );

    co_return HttpResponse{
        .code = static_cast<u32>(response.result_int()),
        .body = boost::beast::buffers_to_string(response.body().data()),
    };
}

kstd::Coro<HttpResponse> httpRequest(const HttpRequest& r) {
    auto executor = co_await boost::asio::this_coro::executor;
    boost::asio::ip::tcp::resolver resolver{ executor };
    boost::beast::tcp_stream stream{ executor };

    auto endpoints =
      co_await resolver.async_resolve(r.host, "80", boost::asio::use_awaitable);
    co_await stream.async_connect(endpoints, boost::asio::use_awaitable);

    const auto response = co_await requestImpl(r, stream);
    boost::beast::error_code ec;
    stream.socket().shutdown(boost::asio::ip::tcp::socket::shutdown_both, ec);

    co_return response;
}

kstd::Coro<HttpResponse> httpsRequest(const HttpRequest& r) {
    auto executor = co_await boost::asio::this_coro::executor;
    boost::asio::ip::tcp::resolver resolver{ executor };

    boost::asio::ssl::context sslContext{ boost::asio::ssl::context::sslv23_client };
    sslContext.set_default_verify_paths();

    boost::beast::ssl_stream<boost::beast::tcp_stream> stream{
        executor, sslContext
    };
    auto endpoints =
      co_await resolver.async_resolve(r.host, "443", boost::asio::use_awaitable);

    co_await boost::beast::get_lowest_layer(stream)
      .async_connect(endpoints, boost::asio::use_awaitable);
    co_await stream.async_handshake(
      boost::asio::ssl::stream_base::client, boost::asio::use_awaitable
    );

    const auto response = co_await requestImpl(r, stream);
    boost::beast::error_code ec;
    co_await stream.async_shutdown(
      boost::asio::redirect_error(boost::asio::use_awaitable, ec)
    );

    co_return response;
}

nlohmann::json HttpResponse::toJson() const { return nlohmann::json::parse(body); }

}  // namespace juno
