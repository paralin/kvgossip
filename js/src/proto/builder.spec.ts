import { ProtoRoot } from './builder';
import { IBuildTransactionRequest } from './interfaces';

import * as ProtoBuf from 'protobufjs';

describe('proto builder', () => {
  it('should build a basic message properly', () => {
    let msg: IBuildTransactionRequest = {
      key: 'test',
    };
    let typ = (ProtoRoot.lookup('ctl.BuildTransactionRequest') as ProtoBuf.Type);
    let arr = typ
      .encode(msg)
      .finish();
    let res: IBuildTransactionRequest = <any>typ.decode(arr).asJSON();
    console.log(res);
    expect(res.key).toBe(msg.key);
  });
});
