#include "JobParser.hh"

#include <kstd/String.hh>
#include <boost/algorithm/string.hpp>

#include "Core.hh"

namespace juno {

void JobParser::parseString(const std::string& job) { tokenize(job); }

void JobParser::tokenize(const std::string& job) {
    static const std::string prefix = "DECLARE JOB";
    static const u64 prefixLen      = prefix.length();

    log::debug("Parsing job: '{}'", job);
    if (not job.starts_with(prefix)) throw Error{ "Invalid job, prefix not found" };

    for (const auto& token : kstd::split(job.substr(prefixLen + 1), ",")) {
        const auto assignment = token.find_first_of("=");
        if (assignment == token.npos)
            throw Error{ "Could not tokenize job: '{}'", token };
        auto k = token.substr(0, assignment);
        auto v = token.substr(assignment + 1);

        boost::algorithm::trim(k);
        boost::algorithm::trim(v);

        log::debug("Parsed token: '{}' = '{}'", k, v);

        if (m_tokens.contains(k))
            throw Error{ "Token '{}' defined more than once", k };
        m_tokens.insert({ k, v });
    }
}

}  // namespace juno
