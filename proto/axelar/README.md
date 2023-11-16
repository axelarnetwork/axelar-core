# Axelar Protobufs

This folder defines protobufs used by Axelar specific Cosmos SDK msg, event, and query types.

## REST API

The REST API (LCD) gets generated automatically from the gRPC service definitions.
The request/response types are defined in `query.proto` for the respective modules, and the query is defined in the `service.proto`.

Note: The request types cannot make use of custom types encoded as bytes as that would be awkward
for REST-based calls. Instead, primitive types such as string is used (for e.g. when specifying addresses, instead of using sdk.AccAddress).
