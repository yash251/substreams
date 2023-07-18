module.exports = {
  run() {
    let encoder = new TextEncoder();
    ensureEqual(encoder.encoding, "utf-8", "bad encoding");

    let value = encoder.encode("ABCD");
    ensureBytesEqual(value, new Uint8Array([0x41, 0x42, 0x43, 0x44]), "bad encoded value");
  },
};

function ensureEqual(left, right, msg) {
  ensure(left === right, `${msg}: ${left} !== ${right}`);
}

function ensureBytesEqual(left, right, msg) {
  const isEquals = () => {
    if (left.length != right.length) return false;

    for (var i = 0; i != left.length; i++) {
      if (left[i] != right[i]) return false;
    }

    return true;
  };

  ensure(isEquals(), `${msg}: ${left} !== ${right}`);
}

function ensure(cond, msg) {
  if (!cond) {
    throw new Error(msg);
  }
}
