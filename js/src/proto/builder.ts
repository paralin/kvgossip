import { PROTO_DEFINITIONS } from './definitions';

import * as ProtoBuf from 'protobufjs';

export const ProtoRoot = ProtoBuf.Root.fromJSON(PROTO_DEFINITIONS);
