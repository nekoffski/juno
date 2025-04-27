#include "Helpers.hh"

namespace juno {

void toProto(const Device* in, api::Device* out) { out->set_uuid(in->uuid); }

}  // namespace juno
