#include "Config.hh"

#include <nlohmann/json.hpp>

namespace juno {

void from_json(const nlohmann::json& j, Config& out) {
    for (const auto& f : j.items()) log::debug("{} - {}", f.key(), f.value().dump());

    j.at("grpc-api-host").get_to(out.grpcApiHost);
    j.at("grpc-api-port").get_to(out.grpcApiPort);
}

Config Config::fromFile(const std::string& path, const kstd::FileSystem& fs) {
    log::debug("Loading config file: {}", path);

    if (not fs.isFile(path)) throw Error{ "Config file doesn't exist: {}", path };

    try {
        return nlohmann::json::parse(fs.readFile(path)).get<Config>();
    } catch (const nlohmann::json::parse_error& e) {
        throw Error{ "Could not parse config file '{}' - {}", path, e.what() };
    }
}

}  // namespace juno
