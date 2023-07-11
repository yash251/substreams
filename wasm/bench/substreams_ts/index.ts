// import {
//   TextEncoder as TextEncoderShim,
//   TextDecoder as TextDecoderShim,
// } from "fastestsmallesttextencoderdecoder"

import "./shims/textEncodeDecoder"
import bigInt from "./shims/bigInt"

import { Block, TransactionTraceStatus } from "./pb/sf/ethereum/type/v2/type_pb"
import {
  DatabaseChanges,
  Field,
  TableChange,
  TableChange_Operation,
} from "./pb/sf/substreams/sink/database/v1/database_pb"

const rocketAddress = bytesFromHex("0xae78736Cd615f374D3085123A210448E74Fc6393")
const approvalTopic = bytesFromHex(
  "0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925",
)
const transferTopic = bytesFromHex(
  "0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef",
)

export function map_noop() {}

export function map_decode_proto_only(data: Uint8Array) {
  const block = new Block()
  block.fromBinary(data)
}

export function map_block(data: Uint8Array) {
  const block = new Block()
  block.fromBinary(data)

  const changes = new DatabaseChanges()

  const blockNumberStr = block.header?.number.toString() ?? ""
  const blockTimestampStr = block.header?.timestamp?.seconds.toString() ?? ""

  block.transactionTraces.forEach((trace) => {
    if (trace.status !== TransactionTraceStatus.SUCCEEDED) {
      return
    }

    trace.calls.forEach((call) => {
      if (call.stateReverted) {
        return
      }

      call.logs.forEach((log) => {
        if (!bytesEqual(log.address, rocketAddress) || log.topics.length === 0) {
          return
        }

        if (bytesEqual(log.topics[0], approvalTopic)) {
          const change = new TableChange()
          change.table = "Approval"
          change.primaryKey = { case: "pk", value: `${bytesToHex(trace.hash)}-${log.index}` }
          change.operation = TableChange_Operation.CREATE
          // @ts-ignore
          change.ordinal = bigInt(0)
          change.fields = [
            new Field({ name: "timestamp", newValue: blockTimestampStr }),
            new Field({ name: "block_number", newValue: blockNumberStr }),
            new Field({ name: "log_index", newValue: log.index.toString() }),
            new Field({ name: "tx_hash", newValue: bytesToHex(trace.hash) }),
            new Field({ name: "spender", newValue: bytesToHex(log.topics[1].slice(12)) }),
            new Field({ name: "owner", newValue: bytesToHex(log.topics[2].slice(12)) }),
            new Field({ name: "amount", newValue: bytesToHex(stripZeroBytes(log.data)) }),
          ]

          changes.tableChanges.push(change)
          return
        }

        if (bytesEqual(log.topics[0], transferTopic)) {
          const change = new TableChange({})
          change.table = "Transfer"
          change.primaryKey = { case: "pk", value: `${bytesToHex(trace.hash)}-${log.index}` }
          change.operation = TableChange_Operation.CREATE
          // @ts-ignore
          change.ordinal = bigInt(0)
          change.fields = [
            new Field({ name: "timestamp", newValue: blockTimestampStr }),
            new Field({ name: "block_number", newValue: blockNumberStr }),
            new Field({ name: "log_index", newValue: log.index.toString() }),
            new Field({ name: "tx_hash", newValue: bytesToHex(trace.hash) }),
            new Field({ name: "sender", newValue: bytesToHex(log.topics[1].slice(12)) }),
            new Field({ name: "receiver", newValue: bytesToHex(log.topics[2].slice(12)) }),
            new Field({ name: "value", newValue: bytesToHex(stripZeroBytes(log.data)) }),
          ]

          changes.tableChanges.push(change)
          return
        }
      })
    })
  })

  substreams_engine.output(changes.toBinary())
}

function stripZeroBytes(input: Uint8Array): Uint8Array {
  for (let i = 0; i != input.length; i++) {
    if (input[i] != 0) {
      return input.slice(i)
    }
  }

  return input
}

function bytesToHex(input: Uint8Array): string {
  // @ts-ignore
  return Buffer.from(input).toString("hex")
}

function bytesFromHex(input: string): Uint8Array {
  if (input.match(/^0(x|X)/)) {
    input = input.slice(2)
  }

  return new Uint8Array(Buffer.from(input, "hex"))
}

function bytesEqual(left: Uint8Array, right: Uint8Array) {
  if (left.length != right.length) return false

  for (var i = 0; i != left.byteLength; i++) {
    if (left[i] != right[i]) return false
  }

  return true
}
