export interface IPutGrantRequest {
  pool?: IGrantAuthorizationPool;
}

export interface IPutGrantResponse {
  revocations?: ISignedData[];
}

export interface IPutRevocationRequest {
  revocation?: ISignedData;
}

export interface IPutRevocationResponse {
}

export interface IBuildTransactionRequest {
  entityPublicKey?: Uint8Array;
  key?: string;
}

export interface IBuildTransactionResponse {
  transaction?: ITransaction;
  revocations?: ISignedData[];
  invalid?: ISignedData[];
}

export interface IPutTransactionRequest {
  transaction?: ITransaction;
}

export interface IPutTransactionResponse {
}

export interface IGetGrantsRequest {
}

export interface IGetGrantsResponse {
  grants?: ISignedData[];
}

export interface IGetKeyRequest {
  key?: string;
}

export interface IGetKeyResponse {
  transaction?: ITransaction;
}

export interface ISignedData {
  bodyType?: SignedDataType;
  body?: Uint8Array;
  signature?: Uint8Array;
}

export const enum SignedDataType {
  SIGNED_GRANT = 0,
  SIGNED_GRANT_REVOCATION = 2,
}

export interface IGrant {
  keyRegex?: string;
  subgrantAllowed?: boolean;
  issueTimestamp?: number;
  issueeKey?: Uint8Array;
  issuerKey?: Uint8Array;
}

export interface IGrantRevocation {
  revokeTimestamp?: number;
  grant?: ISignedData;
}

export interface IGrantAuthorizationPool {
  signedGrants?: ISignedData[];
}

export interface ISerfQueryMessage {
  treeHash?: ISerfTreeHashBroadcast;
  hostNonce?: string;
}

export interface ISerfTreeHashBroadcast {
  treeHash?: Uint8Array;
  syncPort?: number;
}

export interface ITransaction {
  key?: string;
  value?: Uint8Array;
  verification?: ITransactionVerification;
  transactionType?: TransactionType;
}

export const enum TransactionType {
  TRANSACTION_SET = 0,
}

export interface ITransactionValue {
}

export interface ITransactionVerification {
  valueSignature?: Uint8Array;
  signerPublicKey?: Uint8Array;
  grant?: IGrantAuthorizationPool;
  timestamp?: number;
}

export interface ISyncGlobalHash {
  kvgossipVersion?: string;
  globalTreeHash?: Uint8Array;
  hostNonce?: string;
}

export interface ISyncKeyHash {
  key?: string;
  hash?: Uint8Array;
  timestamp?: number;
}

export interface ISyncKey {
  requestKey?: string;
  transaction?: ITransaction;
}

export interface ISyncKeyResult {
  revocations?: ISignedData[];
  updatedKey?: string;
}

export interface ISyncSessionMessage {
  syncGlobalHash?: ISyncGlobalHash;
  syncKeyHash?: ISyncKeyHash;
  syncKey?: ISyncKey;
  syncKeyResult?: ISyncKeyResult;
}
