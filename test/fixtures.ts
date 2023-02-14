
// This file was generated from the fixtures.yaml file.    
const fixture = {
  "ipfs": {
    "root": {
      "_cid": "Qmckhu9X5A4K6wNzQGSrDRoHPhSpbmUELTFRdYjeQZx1M3",
      "_data": "EisKIhIgIE7LR5inp4/eN2jGFANfjmr8mtK0Il7JmoGHItfZh4MSA2RpchhgCgIIAQ==",
      "dir": {
        "_cid": "QmQWmSaF8SdriHSSZUCwTTWwC3ti7WY1dEEX25Etve9m78",
        "_data": "EjEKIhIgD7PCI2Z6hZNKl2Kd7ViL4ByIVDe+K/ENWGXu17xNxHUSCWFzY2lpLnR4dBgpCgIIAQ==",
        "ascii.txt": {
          "_cid": "QmPPwobbE7eyUTZukZxWVu3etJEc7bk3b35bs6LxkLECRa",
          "_data": "CicIAhIhZ29vZGJ5ZSBhcHBsaWNhdGlvbi92bmQuaXBsZC5yYXcKGCE="
        }
      }
    }
  }
}

export const raw = (x: {_data: string}): Buffer => {
  return Buffer.from(x._data, "base64");
}

export const size = (x: { _data: string }): number => {
  return raw(x).length;
};

export const asString = (x: { _data: string }): string => {
  return raw(x).toString("utf-8");
};

export default fixture.ipfs
