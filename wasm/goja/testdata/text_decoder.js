module.exports = {
    run() {
        let decoder = new TextDecoder();
        ensureEqual(decoder.encoding, "utf-8", "bad encoding");

        let value = decoder.decode(new Uint8Array([0x41, 0x42, 0x43, 0x44]));
        ensureEqual(value, "ABCD", "bad decoded value");
    }
}

function ensureEqual(left, right, msg) {
    ensure(left === right, `${msg}: ${left} !== ${right}`)
}

function ensure(cond, msg) {
    if (!cond) {
        throw new Error(msg)
    }
}
