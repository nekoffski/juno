#include "Core.hh"

#include <kstd/String.hh>

namespace juno {

std::string Error::where() const {
    return fmt::format(
      "{}:{}",
      kstd::nameFromPath(m_source.filename, kstd::NameExtractionMode::withExtension),
      m_source.line
    );
}

}  // namespace juno
