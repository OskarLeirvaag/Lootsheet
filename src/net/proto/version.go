package proto

// ProtocolVersion is the wire protocol version exchanged during AUTH.
// Increment this when the protocol changes in a backwards-incompatible way
// (new required fields, changed semantics, removed methods, etc.).
// Additive changes (new optional fields, new methods) do not require a bump.
const ProtocolVersion uint32 = 1
