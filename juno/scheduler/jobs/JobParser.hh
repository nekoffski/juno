#pragma once

#include <unordered_map>
#include <string>

#include "Core.hh"

namespace juno {

class JobParser {
public:
    void parseString(const std::string& job);

private:
    void tokenize(const std::string& job);

    std::unordered_map<std::string, std::string> m_tokens;
};

}  // namespace juno
