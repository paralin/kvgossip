/* tslint:disable:trailing-comma */
/* tslint:disable:quotemark */
/* tslint:disable:max-line-length */
export const PROTO_DEFINITIONS = {
  "nested": {
    "ctl": {
      "nested": {
        "PutGrantRequest": {
          "fields": {
            "pool": {
              "type": "grant.GrantAuthorizationPool",
              "id": 1
            }
          }
        },
        "PutGrantResponse": {
          "fields": {
            "revocations": {
              "rule": "repeated",
              "type": "data.SignedData",
              "id": 1,
              "options": {
                "packed": true
              }
            }
          }
        },
        "PutRevocationRequest": {
          "fields": {
            "revocation": {
              "type": "data.SignedData",
              "id": 1
            }
          }
        },
        "PutRevocationResponse": {
          "fields": {}
        },
        "BuildTransactionRequest": {
          "fields": {
            "entityPublicKey": {
              "type": "bytes",
              "id": 1
            },
            "key": {
              "type": "string",
              "id": 2
            }
          }
        },
        "BuildTransactionResponse": {
          "fields": {
            "transaction": {
              "type": "tx.Transaction",
              "id": 1
            },
            "revocations": {
              "rule": "repeated",
              "type": "data.SignedData",
              "id": 2,
              "options": {
                "packed": true
              }
            },
            "invalid": {
              "rule": "repeated",
              "type": "data.SignedData",
              "id": 3,
              "options": {
                "packed": true
              }
            }
          }
        },
        "PutTransactionRequest": {
          "fields": {
            "transaction": {
              "type": "tx.Transaction",
              "id": 1
            }
          }
        },
        "PutTransactionResponse": {
          "fields": {}
        },
        "GetGrantsRequest": {
          "fields": {}
        },
        "GetGrantsResponse": {
          "fields": {
            "grants": {
              "rule": "repeated",
              "type": "data.SignedData",
              "id": 1,
              "options": {
                "packed": true
              }
            }
          }
        },
        "GetKeyRequest": {
          "fields": {
            "key": {
              "type": "string",
              "id": 1
            }
          }
        },
        "GetKeyResponse": {
          "fields": {
            "transaction": {
              "type": "tx.Transaction",
              "id": 1
            }
          }
        },
        "ControlService": {
          "methods": {
            "PutGrant": {
              "requestType": "PutGrantRequest",
              "responseType": "PutGrantResponse"
            },
            "PutRevocation": {
              "requestType": "PutRevocationRequest",
              "responseType": "PutRevocationResponse"
            },
            "BuildTransaction": {
              "requestType": "BuildTransactionRequest",
              "responseType": "BuildTransactionResponse"
            },
            "PutTransaction": {
              "requestType": "PutTransactionRequest",
              "responseType": "PutTransactionResponse"
            },
            "GetGrants": {
              "requestType": "GetGrantsRequest",
              "responseType": "GetGrantsResponse"
            },
            "GetKey": {
              "requestType": "GetKeyRequest",
              "responseType": "GetKeyResponse"
            }
          }
        }
      }
    },
    "data": {
      "nested": {
        "SignedData": {
          "fields": {
            "bodyType": {
              "type": "SignedDataType",
              "id": 1
            },
            "body": {
              "type": "bytes",
              "id": 2
            },
            "signature": {
              "type": "bytes",
              "id": 3
            }
          },
          "nested": {
            "SignedDataType": {
              "values": {
                "SIGNED_GRANT": 0,
                "SIGNED_GRANT_REVOCATION": 2
              }
            }
          }
        }
      }
    },
    "grant": {
      "nested": {
        "Grant": {
          "fields": {
            "keyRegex": {
              "type": "string",
              "id": 1
            },
            "subgrantAllowed": {
              "type": "bool",
              "id": 2
            },
            "issueTimestamp": {
              "type": "int64",
              "id": 3
            },
            "issueeKey": {
              "type": "bytes",
              "id": 4
            },
            "issuerKey": {
              "type": "bytes",
              "id": 5
            }
          }
        },
        "GrantRevocation": {
          "fields": {
            "revokeTimestamp": {
              "type": "int64",
              "id": 1
            },
            "grant": {
              "type": "data.SignedData",
              "id": 2
            }
          }
        },
        "GrantAuthorizationPool": {
          "fields": {
            "signedGrants": {
              "rule": "repeated",
              "type": "data.SignedData",
              "id": 1,
              "options": {
                "packed": true
              }
            }
          }
        }
      }
    },
    "serf": {
      "nested": {
        "SerfQueryMessage": {
          "fields": {
            "treeHash": {
              "type": "SerfTreeHashBroadcast",
              "id": 1
            },
            "hostNonce": {
              "type": "string",
              "id": 2
            }
          }
        },
        "SerfTreeHashBroadcast": {
          "fields": {
            "treeHash": {
              "type": "bytes",
              "id": 1
            },
            "syncPort": {
              "type": "uint32",
              "id": 2
            }
          }
        }
      }
    },
    "tx": {
      "nested": {
        "Transaction": {
          "fields": {
            "key": {
              "type": "string",
              "id": 1
            },
            "value": {
              "type": "bytes",
              "id": 2
            },
            "verification": {
              "type": "TransactionVerification",
              "id": 3
            },
            "transactionType": {
              "type": "TransactionType",
              "id": 4
            }
          },
          "nested": {
            "TransactionType": {
              "values": {
                "TRANSACTION_SET": 0
              }
            }
          }
        },
        "TransactionValue": {
          "fields": {}
        },
        "TransactionVerification": {
          "fields": {
            "valueSignature": {
              "type": "bytes",
              "id": 1
            },
            "signerPublicKey": {
              "type": "bytes",
              "id": 2
            },
            "grant": {
              "type": "grant.GrantAuthorizationPool",
              "id": 3
            },
            "timestamp": {
              "type": "uint64",
              "id": 4
            }
          }
        }
      }
    },
    "sync": {
      "nested": {
        "SyncGlobalHash": {
          "fields": {
            "kvgossipVersion": {
              "type": "string",
              "id": 1
            },
            "globalTreeHash": {
              "type": "bytes",
              "id": 2
            },
            "hostNonce": {
              "type": "string",
              "id": 3
            }
          }
        },
        "SyncKeyHash": {
          "fields": {
            "key": {
              "type": "string",
              "id": 1
            },
            "hash": {
              "type": "bytes",
              "id": 2
            },
            "timestamp": {
              "type": "uint64",
              "id": 3
            }
          }
        },
        "SyncKey": {
          "fields": {
            "requestKey": {
              "type": "string",
              "id": 1
            },
            "transaction": {
              "type": "tx.Transaction",
              "id": 2
            }
          }
        },
        "SyncKeyResult": {
          "fields": {
            "revocations": {
              "rule": "repeated",
              "type": "data.SignedData",
              "id": 1,
              "options": {
                "packed": true
              }
            },
            "updatedKey": {
              "type": "string",
              "id": 2
            }
          }
        },
        "SyncSessionMessage": {
          "fields": {
            "syncGlobalHash": {
              "type": "SyncGlobalHash",
              "id": 1
            },
            "syncKeyHash": {
              "type": "SyncKeyHash",
              "id": 2
            },
            "syncKey": {
              "type": "SyncKey",
              "id": 3
            },
            "syncKeyResult": {
              "type": "SyncKeyResult",
              "id": 4
            }
          }
        },
        "SyncService": {
          "methods": {
            "SyncSession": {
              "requestType": "SyncSessionMessage",
              "requestStream": true,
              "responseType": "SyncSessionMessage",
              "responseStream": true
            }
          }
        }
      }
    }
  }
};
