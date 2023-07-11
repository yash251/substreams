
declare namespace Substreams {
  interface Module {
    exports: any
  }
}

declare var module: Substreams.Module

class Buffer {
  static from(input: string, encoding: string): Uint8Array
}

declare namespace substreams_engine {
  function output(bytes: Uint8Array)
}

// export * from "./shims/bigInt"
