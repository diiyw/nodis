/*eslint-disable block-scoped-var, id-length, no-control-regex, no-magic-numbers, no-prototype-builtins, no-redeclare, no-shadow, no-var, sort-vars*/
"use strict";

var $protobuf = require("protobufjs/minimal");

// Common aliases
var $Reader = $protobuf.Reader, $Writer = $protobuf.Writer, $util = $protobuf.util;

// Exported root namespace
var $root = $protobuf.roots["default"] || ($protobuf.roots["default"] = {});

$root.pb = (function() {

    /**
     * Namespace pb.
     * @exports pb
     * @namespace
     */
    var pb = {};

    /**
     * OpType enum.
     * @name pb.OpType
     * @enum {number}
     * @property {number} None=0 None value
     * @property {number} Clear=1 Clear value
     * @property {number} Del=2 Del value
     * @property {number} Expire=3 Expire value
     * @property {number} ExpireAt=4 ExpireAt value
     * @property {number} HClear=5 HClear value
     * @property {number} HDel=6 HDel value
     * @property {number} HIncrBy=7 HIncrBy value
     * @property {number} HIncrByFloat=8 HIncrByFloat value
     * @property {number} HMSet=9 HMSet value
     * @property {number} HSet=10 HSet value
     * @property {number} HSetNX=11 HSetNX value
     * @property {number} LInsert=12 LInsert value
     * @property {number} LPop=13 LPop value
     * @property {number} LPopRPush=14 LPopRPush value
     * @property {number} LPush=15 LPush value
     * @property {number} LPushX=16 LPushX value
     * @property {number} LRem=17 LRem value
     * @property {number} LSet=18 LSet value
     * @property {number} LTrim=19 LTrim value
     * @property {number} RPop=20 RPop value
     * @property {number} RPopLPush=21 RPopLPush value
     * @property {number} RPush=22 RPush value
     * @property {number} RPushX=23 RPushX value
     * @property {number} SAdd=24 SAdd value
     * @property {number} SRem=25 SRem value
     * @property {number} Set=26 Set value
     * @property {number} ZAdd=27 ZAdd value
     * @property {number} ZClear=28 ZClear value
     * @property {number} ZIncrBy=29 ZIncrBy value
     * @property {number} ZRem=30 ZRem value
     * @property {number} ZRemRangeByRank=31 ZRemRangeByRank value
     * @property {number} ZRemRangeByScore=32 ZRemRangeByScore value
     * @property {number} Rename=33 Rename value
     */
    pb.OpType = (function() {
        var valuesById = {}, values = Object.create(valuesById);
        values[valuesById[0] = "None"] = 0;
        values[valuesById[1] = "Clear"] = 1;
        values[valuesById[2] = "Del"] = 2;
        values[valuesById[3] = "Expire"] = 3;
        values[valuesById[4] = "ExpireAt"] = 4;
        values[valuesById[5] = "HClear"] = 5;
        values[valuesById[6] = "HDel"] = 6;
        values[valuesById[7] = "HIncrBy"] = 7;
        values[valuesById[8] = "HIncrByFloat"] = 8;
        values[valuesById[9] = "HMSet"] = 9;
        values[valuesById[10] = "HSet"] = 10;
        values[valuesById[11] = "HSetNX"] = 11;
        values[valuesById[12] = "LInsert"] = 12;
        values[valuesById[13] = "LPop"] = 13;
        values[valuesById[14] = "LPopRPush"] = 14;
        values[valuesById[15] = "LPush"] = 15;
        values[valuesById[16] = "LPushX"] = 16;
        values[valuesById[17] = "LRem"] = 17;
        values[valuesById[18] = "LSet"] = 18;
        values[valuesById[19] = "LTrim"] = 19;
        values[valuesById[20] = "RPop"] = 20;
        values[valuesById[21] = "RPopLPush"] = 21;
        values[valuesById[22] = "RPush"] = 22;
        values[valuesById[23] = "RPushX"] = 23;
        values[valuesById[24] = "SAdd"] = 24;
        values[valuesById[25] = "SRem"] = 25;
        values[valuesById[26] = "Set"] = 26;
        values[valuesById[27] = "ZAdd"] = 27;
        values[valuesById[28] = "ZClear"] = 28;
        values[valuesById[29] = "ZIncrBy"] = 29;
        values[valuesById[30] = "ZRem"] = 30;
        values[valuesById[31] = "ZRemRangeByRank"] = 31;
        values[valuesById[32] = "ZRemRangeByScore"] = 32;
        values[valuesById[33] = "Rename"] = 33;
        return values;
    })();

    pb.Operation = (function() {

        /**
         * Properties of an Operation.
         * @memberof pb
         * @interface IOperation
         * @property {pb.OpType|null} [Type] Operation Type
         * @property {string|null} [Key] Operation Key
         * @property {string|null} [Member] Operation Member
         * @property {Uint8Array|null} [Value] Operation Value
         * @property {number|Long|null} [Expiration] Operation Expiration
         * @property {number|null} [Score] Operation Score
         * @property {Array.<Uint8Array>|null} [Values] Operation Values
         * @property {string|null} [DstKey] Operation DstKey
         * @property {Uint8Array|null} [Pivot] Operation Pivot
         * @property {number|Long|null} [Count] Operation Count
         * @property {number|Long|null} [Index] Operation Index
         * @property {Array.<string>|null} [Members] Operation Members
         * @property {number|Long|null} [start] Operation start
         * @property {number|Long|null} [stop] Operation stop
         * @property {number|null} [min] Operation min
         * @property {number|null} [max] Operation max
         * @property {string|null} [Field] Operation Field
         * @property {number|null} [IncrFloat] Operation IncrFloat
         * @property {number|Long|null} [IncrInt] Operation IncrInt
         * @property {boolean|null} [Before] Operation Before
         */

        /**
         * Constructs a new Operation.
         * @memberof pb
         * @classdesc Represents an Operation.
         * @implements IOperation
         * @constructor
         * @param {pb.IOperation=} [properties] Properties to set
         */
        function Operation(properties) {
            this.Values = [];
            this.Members = [];
            if (properties)
                for (var keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null)
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * Operation Type.
         * @member {pb.OpType} Type
         * @memberof pb.Operation
         * @instance
         */
        Operation.prototype.Type = 0;

        /**
         * Operation Key.
         * @member {string} Key
         * @memberof pb.Operation
         * @instance
         */
        Operation.prototype.Key = "";

        /**
         * Operation Member.
         * @member {string} Member
         * @memberof pb.Operation
         * @instance
         */
        Operation.prototype.Member = "";

        /**
         * Operation Value.
         * @member {Uint8Array} Value
         * @memberof pb.Operation
         * @instance
         */
        Operation.prototype.Value = $util.newBuffer([]);

        /**
         * Operation Expiration.
         * @member {number|Long} Expiration
         * @memberof pb.Operation
         * @instance
         */
        Operation.prototype.Expiration = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

        /**
         * Operation Score.
         * @member {number} Score
         * @memberof pb.Operation
         * @instance
         */
        Operation.prototype.Score = 0;

        /**
         * Operation Values.
         * @member {Array.<Uint8Array>} Values
         * @memberof pb.Operation
         * @instance
         */
        Operation.prototype.Values = $util.emptyArray;

        /**
         * Operation DstKey.
         * @member {string} DstKey
         * @memberof pb.Operation
         * @instance
         */
        Operation.prototype.DstKey = "";

        /**
         * Operation Pivot.
         * @member {Uint8Array} Pivot
         * @memberof pb.Operation
         * @instance
         */
        Operation.prototype.Pivot = $util.newBuffer([]);

        /**
         * Operation Count.
         * @member {number|Long} Count
         * @memberof pb.Operation
         * @instance
         */
        Operation.prototype.Count = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

        /**
         * Operation Index.
         * @member {number|Long} Index
         * @memberof pb.Operation
         * @instance
         */
        Operation.prototype.Index = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

        /**
         * Operation Members.
         * @member {Array.<string>} Members
         * @memberof pb.Operation
         * @instance
         */
        Operation.prototype.Members = $util.emptyArray;

        /**
         * Operation start.
         * @member {number|Long} start
         * @memberof pb.Operation
         * @instance
         */
        Operation.prototype.start = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

        /**
         * Operation stop.
         * @member {number|Long} stop
         * @memberof pb.Operation
         * @instance
         */
        Operation.prototype.stop = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

        /**
         * Operation min.
         * @member {number} min
         * @memberof pb.Operation
         * @instance
         */
        Operation.prototype.min = 0;

        /**
         * Operation max.
         * @member {number} max
         * @memberof pb.Operation
         * @instance
         */
        Operation.prototype.max = 0;

        /**
         * Operation Field.
         * @member {string} Field
         * @memberof pb.Operation
         * @instance
         */
        Operation.prototype.Field = "";

        /**
         * Operation IncrFloat.
         * @member {number} IncrFloat
         * @memberof pb.Operation
         * @instance
         */
        Operation.prototype.IncrFloat = 0;

        /**
         * Operation IncrInt.
         * @member {number|Long} IncrInt
         * @memberof pb.Operation
         * @instance
         */
        Operation.prototype.IncrInt = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

        /**
         * Operation Before.
         * @member {boolean} Before
         * @memberof pb.Operation
         * @instance
         */
        Operation.prototype.Before = false;

        /**
         * Creates a new Operation instance using the specified properties.
         * @function create
         * @memberof pb.Operation
         * @static
         * @param {pb.IOperation=} [properties] Properties to set
         * @returns {pb.Operation} Operation instance
         */
        Operation.create = function create(properties) {
            return new Operation(properties);
        };

        /**
         * Encodes the specified Operation message. Does not implicitly {@link pb.Operation.verify|verify} messages.
         * @function encode
         * @memberof pb.Operation
         * @static
         * @param {pb.IOperation} message Operation message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        Operation.encode = function encode(message, writer) {
            if (!writer)
                writer = $Writer.create();
            if (message.Type != null && Object.hasOwnProperty.call(message, "Type"))
                writer.uint32(/* id 1, wireType 0 =*/8).int32(message.Type);
            if (message.Key != null && Object.hasOwnProperty.call(message, "Key"))
                writer.uint32(/* id 2, wireType 2 =*/18).string(message.Key);
            if (message.Member != null && Object.hasOwnProperty.call(message, "Member"))
                writer.uint32(/* id 3, wireType 2 =*/26).string(message.Member);
            if (message.Value != null && Object.hasOwnProperty.call(message, "Value"))
                writer.uint32(/* id 4, wireType 2 =*/34).bytes(message.Value);
            if (message.Expiration != null && Object.hasOwnProperty.call(message, "Expiration"))
                writer.uint32(/* id 5, wireType 0 =*/40).int64(message.Expiration);
            if (message.Score != null && Object.hasOwnProperty.call(message, "Score"))
                writer.uint32(/* id 6, wireType 1 =*/49).double(message.Score);
            if (message.Values != null && message.Values.length)
                for (var i = 0; i < message.Values.length; ++i)
                    writer.uint32(/* id 7, wireType 2 =*/58).bytes(message.Values[i]);
            if (message.DstKey != null && Object.hasOwnProperty.call(message, "DstKey"))
                writer.uint32(/* id 8, wireType 2 =*/66).string(message.DstKey);
            if (message.Pivot != null && Object.hasOwnProperty.call(message, "Pivot"))
                writer.uint32(/* id 9, wireType 2 =*/74).bytes(message.Pivot);
            if (message.Count != null && Object.hasOwnProperty.call(message, "Count"))
                writer.uint32(/* id 10, wireType 0 =*/80).int64(message.Count);
            if (message.Index != null && Object.hasOwnProperty.call(message, "Index"))
                writer.uint32(/* id 11, wireType 0 =*/88).int64(message.Index);
            if (message.Members != null && message.Members.length)
                for (var i = 0; i < message.Members.length; ++i)
                    writer.uint32(/* id 12, wireType 2 =*/98).string(message.Members[i]);
            if (message.start != null && Object.hasOwnProperty.call(message, "start"))
                writer.uint32(/* id 13, wireType 0 =*/104).int64(message.start);
            if (message.stop != null && Object.hasOwnProperty.call(message, "stop"))
                writer.uint32(/* id 14, wireType 0 =*/112).int64(message.stop);
            if (message.min != null && Object.hasOwnProperty.call(message, "min"))
                writer.uint32(/* id 15, wireType 1 =*/121).double(message.min);
            if (message.max != null && Object.hasOwnProperty.call(message, "max"))
                writer.uint32(/* id 16, wireType 1 =*/129).double(message.max);
            if (message.Field != null && Object.hasOwnProperty.call(message, "Field"))
                writer.uint32(/* id 17, wireType 2 =*/138).string(message.Field);
            if (message.IncrFloat != null && Object.hasOwnProperty.call(message, "IncrFloat"))
                writer.uint32(/* id 18, wireType 1 =*/145).double(message.IncrFloat);
            if (message.IncrInt != null && Object.hasOwnProperty.call(message, "IncrInt"))
                writer.uint32(/* id 19, wireType 0 =*/152).int64(message.IncrInt);
            if (message.Before != null && Object.hasOwnProperty.call(message, "Before"))
                writer.uint32(/* id 20, wireType 0 =*/160).bool(message.Before);
            return writer;
        };

        /**
         * Encodes the specified Operation message, length delimited. Does not implicitly {@link pb.Operation.verify|verify} messages.
         * @function encodeDelimited
         * @memberof pb.Operation
         * @static
         * @param {pb.IOperation} message Operation message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        Operation.encodeDelimited = function encodeDelimited(message, writer) {
            return this.encode(message, writer).ldelim();
        };

        /**
         * Decodes an Operation message from the specified reader or buffer.
         * @function decode
         * @memberof pb.Operation
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {pb.Operation} Operation
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        Operation.decode = function decode(reader, length) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            var end = length === undefined ? reader.len : reader.pos + length, message = new $root.pb.Operation();
            while (reader.pos < end) {
                var tag = reader.uint32();
                switch (tag >>> 3) {
                case 1: {
                        message.Type = reader.int32();
                        break;
                    }
                case 2: {
                        message.Key = reader.string();
                        break;
                    }
                case 3: {
                        message.Member = reader.string();
                        break;
                    }
                case 4: {
                        message.Value = reader.bytes();
                        break;
                    }
                case 5: {
                        message.Expiration = reader.int64();
                        break;
                    }
                case 6: {
                        message.Score = reader.double();
                        break;
                    }
                case 7: {
                        if (!(message.Values && message.Values.length))
                            message.Values = [];
                        message.Values.push(reader.bytes());
                        break;
                    }
                case 8: {
                        message.DstKey = reader.string();
                        break;
                    }
                case 9: {
                        message.Pivot = reader.bytes();
                        break;
                    }
                case 10: {
                        message.Count = reader.int64();
                        break;
                    }
                case 11: {
                        message.Index = reader.int64();
                        break;
                    }
                case 12: {
                        if (!(message.Members && message.Members.length))
                            message.Members = [];
                        message.Members.push(reader.string());
                        break;
                    }
                case 13: {
                        message.start = reader.int64();
                        break;
                    }
                case 14: {
                        message.stop = reader.int64();
                        break;
                    }
                case 15: {
                        message.min = reader.double();
                        break;
                    }
                case 16: {
                        message.max = reader.double();
                        break;
                    }
                case 17: {
                        message.Field = reader.string();
                        break;
                    }
                case 18: {
                        message.IncrFloat = reader.double();
                        break;
                    }
                case 19: {
                        message.IncrInt = reader.int64();
                        break;
                    }
                case 20: {
                        message.Before = reader.bool();
                        break;
                    }
                default:
                    reader.skipType(tag & 7);
                    break;
                }
            }
            return message;
        };

        /**
         * Decodes an Operation message from the specified reader or buffer, length delimited.
         * @function decodeDelimited
         * @memberof pb.Operation
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @returns {pb.Operation} Operation
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        Operation.decodeDelimited = function decodeDelimited(reader) {
            if (!(reader instanceof $Reader))
                reader = new $Reader(reader);
            return this.decode(reader, reader.uint32());
        };

        /**
         * Verifies an Operation message.
         * @function verify
         * @memberof pb.Operation
         * @static
         * @param {Object.<string,*>} message Plain object to verify
         * @returns {string|null} `null` if valid, otherwise the reason why it is not
         */
        Operation.verify = function verify(message) {
            if (typeof message !== "object" || message === null)
                return "object expected";
            if (message.Type != null && message.hasOwnProperty("Type"))
                switch (message.Type) {
                default:
                    return "Type: enum value expected";
                case 0:
                case 1:
                case 2:
                case 3:
                case 4:
                case 5:
                case 6:
                case 7:
                case 8:
                case 9:
                case 10:
                case 11:
                case 12:
                case 13:
                case 14:
                case 15:
                case 16:
                case 17:
                case 18:
                case 19:
                case 20:
                case 21:
                case 22:
                case 23:
                case 24:
                case 25:
                case 26:
                case 27:
                case 28:
                case 29:
                case 30:
                case 31:
                case 32:
                case 33:
                    break;
                }
            if (message.Key != null && message.hasOwnProperty("Key"))
                if (!$util.isString(message.Key))
                    return "Key: string expected";
            if (message.Member != null && message.hasOwnProperty("Member"))
                if (!$util.isString(message.Member))
                    return "Member: string expected";
            if (message.Value != null && message.hasOwnProperty("Value"))
                if (!(message.Value && typeof message.Value.length === "number" || $util.isString(message.Value)))
                    return "Value: buffer expected";
            if (message.Expiration != null && message.hasOwnProperty("Expiration"))
                if (!$util.isInteger(message.Expiration) && !(message.Expiration && $util.isInteger(message.Expiration.low) && $util.isInteger(message.Expiration.high)))
                    return "Expiration: integer|Long expected";
            if (message.Score != null && message.hasOwnProperty("Score"))
                if (typeof message.Score !== "number")
                    return "Score: number expected";
            if (message.Values != null && message.hasOwnProperty("Values")) {
                if (!Array.isArray(message.Values))
                    return "Values: array expected";
                for (var i = 0; i < message.Values.length; ++i)
                    if (!(message.Values[i] && typeof message.Values[i].length === "number" || $util.isString(message.Values[i])))
                        return "Values: buffer[] expected";
            }
            if (message.DstKey != null && message.hasOwnProperty("DstKey"))
                if (!$util.isString(message.DstKey))
                    return "DstKey: string expected";
            if (message.Pivot != null && message.hasOwnProperty("Pivot"))
                if (!(message.Pivot && typeof message.Pivot.length === "number" || $util.isString(message.Pivot)))
                    return "Pivot: buffer expected";
            if (message.Count != null && message.hasOwnProperty("Count"))
                if (!$util.isInteger(message.Count) && !(message.Count && $util.isInteger(message.Count.low) && $util.isInteger(message.Count.high)))
                    return "Count: integer|Long expected";
            if (message.Index != null && message.hasOwnProperty("Index"))
                if (!$util.isInteger(message.Index) && !(message.Index && $util.isInteger(message.Index.low) && $util.isInteger(message.Index.high)))
                    return "Index: integer|Long expected";
            if (message.Members != null && message.hasOwnProperty("Members")) {
                if (!Array.isArray(message.Members))
                    return "Members: array expected";
                for (var i = 0; i < message.Members.length; ++i)
                    if (!$util.isString(message.Members[i]))
                        return "Members: string[] expected";
            }
            if (message.start != null && message.hasOwnProperty("start"))
                if (!$util.isInteger(message.start) && !(message.start && $util.isInteger(message.start.low) && $util.isInteger(message.start.high)))
                    return "start: integer|Long expected";
            if (message.stop != null && message.hasOwnProperty("stop"))
                if (!$util.isInteger(message.stop) && !(message.stop && $util.isInteger(message.stop.low) && $util.isInteger(message.stop.high)))
                    return "stop: integer|Long expected";
            if (message.min != null && message.hasOwnProperty("min"))
                if (typeof message.min !== "number")
                    return "min: number expected";
            if (message.max != null && message.hasOwnProperty("max"))
                if (typeof message.max !== "number")
                    return "max: number expected";
            if (message.Field != null && message.hasOwnProperty("Field"))
                if (!$util.isString(message.Field))
                    return "Field: string expected";
            if (message.IncrFloat != null && message.hasOwnProperty("IncrFloat"))
                if (typeof message.IncrFloat !== "number")
                    return "IncrFloat: number expected";
            if (message.IncrInt != null && message.hasOwnProperty("IncrInt"))
                if (!$util.isInteger(message.IncrInt) && !(message.IncrInt && $util.isInteger(message.IncrInt.low) && $util.isInteger(message.IncrInt.high)))
                    return "IncrInt: integer|Long expected";
            if (message.Before != null && message.hasOwnProperty("Before"))
                if (typeof message.Before !== "boolean")
                    return "Before: boolean expected";
            return null;
        };

        /**
         * Creates an Operation message from a plain object. Also converts values to their respective internal types.
         * @function fromObject
         * @memberof pb.Operation
         * @static
         * @param {Object.<string,*>} object Plain object
         * @returns {pb.Operation} Operation
         */
        Operation.fromObject = function fromObject(object) {
            if (object instanceof $root.pb.Operation)
                return object;
            var message = new $root.pb.Operation();
            switch (object.Type) {
            default:
                if (typeof object.Type === "number") {
                    message.Type = object.Type;
                    break;
                }
                break;
            case "None":
            case 0:
                message.Type = 0;
                break;
            case "Clear":
            case 1:
                message.Type = 1;
                break;
            case "Del":
            case 2:
                message.Type = 2;
                break;
            case "Expire":
            case 3:
                message.Type = 3;
                break;
            case "ExpireAt":
            case 4:
                message.Type = 4;
                break;
            case "HClear":
            case 5:
                message.Type = 5;
                break;
            case "HDel":
            case 6:
                message.Type = 6;
                break;
            case "HIncrBy":
            case 7:
                message.Type = 7;
                break;
            case "HIncrByFloat":
            case 8:
                message.Type = 8;
                break;
            case "HMSet":
            case 9:
                message.Type = 9;
                break;
            case "HSet":
            case 10:
                message.Type = 10;
                break;
            case "HSetNX":
            case 11:
                message.Type = 11;
                break;
            case "LInsert":
            case 12:
                message.Type = 12;
                break;
            case "LPop":
            case 13:
                message.Type = 13;
                break;
            case "LPopRPush":
            case 14:
                message.Type = 14;
                break;
            case "LPush":
            case 15:
                message.Type = 15;
                break;
            case "LPushX":
            case 16:
                message.Type = 16;
                break;
            case "LRem":
            case 17:
                message.Type = 17;
                break;
            case "LSet":
            case 18:
                message.Type = 18;
                break;
            case "LTrim":
            case 19:
                message.Type = 19;
                break;
            case "RPop":
            case 20:
                message.Type = 20;
                break;
            case "RPopLPush":
            case 21:
                message.Type = 21;
                break;
            case "RPush":
            case 22:
                message.Type = 22;
                break;
            case "RPushX":
            case 23:
                message.Type = 23;
                break;
            case "SAdd":
            case 24:
                message.Type = 24;
                break;
            case "SRem":
            case 25:
                message.Type = 25;
                break;
            case "Set":
            case 26:
                message.Type = 26;
                break;
            case "ZAdd":
            case 27:
                message.Type = 27;
                break;
            case "ZClear":
            case 28:
                message.Type = 28;
                break;
            case "ZIncrBy":
            case 29:
                message.Type = 29;
                break;
            case "ZRem":
            case 30:
                message.Type = 30;
                break;
            case "ZRemRangeByRank":
            case 31:
                message.Type = 31;
                break;
            case "ZRemRangeByScore":
            case 32:
                message.Type = 32;
                break;
            case "Rename":
            case 33:
                message.Type = 33;
                break;
            }
            if (object.Key != null)
                message.Key = String(object.Key);
            if (object.Member != null)
                message.Member = String(object.Member);
            if (object.Value != null)
                if (typeof object.Value === "string")
                    $util.base64.decode(object.Value, message.Value = $util.newBuffer($util.base64.length(object.Value)), 0);
                else if (object.Value.length >= 0)
                    message.Value = object.Value;
            if (object.Expiration != null)
                if ($util.Long)
                    (message.Expiration = $util.Long.fromValue(object.Expiration)).unsigned = false;
                else if (typeof object.Expiration === "string")
                    message.Expiration = parseInt(object.Expiration, 10);
                else if (typeof object.Expiration === "number")
                    message.Expiration = object.Expiration;
                else if (typeof object.Expiration === "object")
                    message.Expiration = new $util.LongBits(object.Expiration.low >>> 0, object.Expiration.high >>> 0).toNumber();
            if (object.Score != null)
                message.Score = Number(object.Score);
            if (object.Values) {
                if (!Array.isArray(object.Values))
                    throw TypeError(".pb.Operation.Values: array expected");
                message.Values = [];
                for (var i = 0; i < object.Values.length; ++i)
                    if (typeof object.Values[i] === "string")
                        $util.base64.decode(object.Values[i], message.Values[i] = $util.newBuffer($util.base64.length(object.Values[i])), 0);
                    else if (object.Values[i].length >= 0)
                        message.Values[i] = object.Values[i];
            }
            if (object.DstKey != null)
                message.DstKey = String(object.DstKey);
            if (object.Pivot != null)
                if (typeof object.Pivot === "string")
                    $util.base64.decode(object.Pivot, message.Pivot = $util.newBuffer($util.base64.length(object.Pivot)), 0);
                else if (object.Pivot.length >= 0)
                    message.Pivot = object.Pivot;
            if (object.Count != null)
                if ($util.Long)
                    (message.Count = $util.Long.fromValue(object.Count)).unsigned = false;
                else if (typeof object.Count === "string")
                    message.Count = parseInt(object.Count, 10);
                else if (typeof object.Count === "number")
                    message.Count = object.Count;
                else if (typeof object.Count === "object")
                    message.Count = new $util.LongBits(object.Count.low >>> 0, object.Count.high >>> 0).toNumber();
            if (object.Index != null)
                if ($util.Long)
                    (message.Index = $util.Long.fromValue(object.Index)).unsigned = false;
                else if (typeof object.Index === "string")
                    message.Index = parseInt(object.Index, 10);
                else if (typeof object.Index === "number")
                    message.Index = object.Index;
                else if (typeof object.Index === "object")
                    message.Index = new $util.LongBits(object.Index.low >>> 0, object.Index.high >>> 0).toNumber();
            if (object.Members) {
                if (!Array.isArray(object.Members))
                    throw TypeError(".pb.Operation.Members: array expected");
                message.Members = [];
                for (var i = 0; i < object.Members.length; ++i)
                    message.Members[i] = String(object.Members[i]);
            }
            if (object.start != null)
                if ($util.Long)
                    (message.start = $util.Long.fromValue(object.start)).unsigned = false;
                else if (typeof object.start === "string")
                    message.start = parseInt(object.start, 10);
                else if (typeof object.start === "number")
                    message.start = object.start;
                else if (typeof object.start === "object")
                    message.start = new $util.LongBits(object.start.low >>> 0, object.start.high >>> 0).toNumber();
            if (object.stop != null)
                if ($util.Long)
                    (message.stop = $util.Long.fromValue(object.stop)).unsigned = false;
                else if (typeof object.stop === "string")
                    message.stop = parseInt(object.stop, 10);
                else if (typeof object.stop === "number")
                    message.stop = object.stop;
                else if (typeof object.stop === "object")
                    message.stop = new $util.LongBits(object.stop.low >>> 0, object.stop.high >>> 0).toNumber();
            if (object.min != null)
                message.min = Number(object.min);
            if (object.max != null)
                message.max = Number(object.max);
            if (object.Field != null)
                message.Field = String(object.Field);
            if (object.IncrFloat != null)
                message.IncrFloat = Number(object.IncrFloat);
            if (object.IncrInt != null)
                if ($util.Long)
                    (message.IncrInt = $util.Long.fromValue(object.IncrInt)).unsigned = false;
                else if (typeof object.IncrInt === "string")
                    message.IncrInt = parseInt(object.IncrInt, 10);
                else if (typeof object.IncrInt === "number")
                    message.IncrInt = object.IncrInt;
                else if (typeof object.IncrInt === "object")
                    message.IncrInt = new $util.LongBits(object.IncrInt.low >>> 0, object.IncrInt.high >>> 0).toNumber();
            if (object.Before != null)
                message.Before = Boolean(object.Before);
            return message;
        };

        /**
         * Creates a plain object from an Operation message. Also converts values to other types if specified.
         * @function toObject
         * @memberof pb.Operation
         * @static
         * @param {pb.Operation} message Operation
         * @param {$protobuf.IConversionOptions} [options] Conversion options
         * @returns {Object.<string,*>} Plain object
         */
        Operation.toObject = function toObject(message, options) {
            if (!options)
                options = {};
            var object = {};
            if (options.arrays || options.defaults) {
                object.Values = [];
                object.Members = [];
            }
            if (options.defaults) {
                object.Type = options.enums === String ? "None" : 0;
                object.Key = "";
                object.Member = "";
                if (options.bytes === String)
                    object.Value = "";
                else {
                    object.Value = [];
                    if (options.bytes !== Array)
                        object.Value = $util.newBuffer(object.Value);
                }
                if ($util.Long) {
                    var long = new $util.Long(0, 0, false);
                    object.Expiration = options.longs === String ? long.toString() : options.longs === Number ? long.toNumber() : long;
                } else
                    object.Expiration = options.longs === String ? "0" : 0;
                object.Score = 0;
                object.DstKey = "";
                if (options.bytes === String)
                    object.Pivot = "";
                else {
                    object.Pivot = [];
                    if (options.bytes !== Array)
                        object.Pivot = $util.newBuffer(object.Pivot);
                }
                if ($util.Long) {
                    var long = new $util.Long(0, 0, false);
                    object.Count = options.longs === String ? long.toString() : options.longs === Number ? long.toNumber() : long;
                } else
                    object.Count = options.longs === String ? "0" : 0;
                if ($util.Long) {
                    var long = new $util.Long(0, 0, false);
                    object.Index = options.longs === String ? long.toString() : options.longs === Number ? long.toNumber() : long;
                } else
                    object.Index = options.longs === String ? "0" : 0;
                if ($util.Long) {
                    var long = new $util.Long(0, 0, false);
                    object.start = options.longs === String ? long.toString() : options.longs === Number ? long.toNumber() : long;
                } else
                    object.start = options.longs === String ? "0" : 0;
                if ($util.Long) {
                    var long = new $util.Long(0, 0, false);
                    object.stop = options.longs === String ? long.toString() : options.longs === Number ? long.toNumber() : long;
                } else
                    object.stop = options.longs === String ? "0" : 0;
                object.min = 0;
                object.max = 0;
                object.Field = "";
                object.IncrFloat = 0;
                if ($util.Long) {
                    var long = new $util.Long(0, 0, false);
                    object.IncrInt = options.longs === String ? long.toString() : options.longs === Number ? long.toNumber() : long;
                } else
                    object.IncrInt = options.longs === String ? "0" : 0;
                object.Before = false;
            }
            if (message.Type != null && message.hasOwnProperty("Type"))
                object.Type = options.enums === String ? $root.pb.OpType[message.Type] === undefined ? message.Type : $root.pb.OpType[message.Type] : message.Type;
            if (message.Key != null && message.hasOwnProperty("Key"))
                object.Key = message.Key;
            if (message.Member != null && message.hasOwnProperty("Member"))
                object.Member = message.Member;
            if (message.Value != null && message.hasOwnProperty("Value"))
                object.Value = options.bytes === String ? $util.base64.encode(message.Value, 0, message.Value.length) : options.bytes === Array ? Array.prototype.slice.call(message.Value) : message.Value;
            if (message.Expiration != null && message.hasOwnProperty("Expiration"))
                if (typeof message.Expiration === "number")
                    object.Expiration = options.longs === String ? String(message.Expiration) : message.Expiration;
                else
                    object.Expiration = options.longs === String ? $util.Long.prototype.toString.call(message.Expiration) : options.longs === Number ? new $util.LongBits(message.Expiration.low >>> 0, message.Expiration.high >>> 0).toNumber() : message.Expiration;
            if (message.Score != null && message.hasOwnProperty("Score"))
                object.Score = options.json && !isFinite(message.Score) ? String(message.Score) : message.Score;
            if (message.Values && message.Values.length) {
                object.Values = [];
                for (var j = 0; j < message.Values.length; ++j)
                    object.Values[j] = options.bytes === String ? $util.base64.encode(message.Values[j], 0, message.Values[j].length) : options.bytes === Array ? Array.prototype.slice.call(message.Values[j]) : message.Values[j];
            }
            if (message.DstKey != null && message.hasOwnProperty("DstKey"))
                object.DstKey = message.DstKey;
            if (message.Pivot != null && message.hasOwnProperty("Pivot"))
                object.Pivot = options.bytes === String ? $util.base64.encode(message.Pivot, 0, message.Pivot.length) : options.bytes === Array ? Array.prototype.slice.call(message.Pivot) : message.Pivot;
            if (message.Count != null && message.hasOwnProperty("Count"))
                if (typeof message.Count === "number")
                    object.Count = options.longs === String ? String(message.Count) : message.Count;
                else
                    object.Count = options.longs === String ? $util.Long.prototype.toString.call(message.Count) : options.longs === Number ? new $util.LongBits(message.Count.low >>> 0, message.Count.high >>> 0).toNumber() : message.Count;
            if (message.Index != null && message.hasOwnProperty("Index"))
                if (typeof message.Index === "number")
                    object.Index = options.longs === String ? String(message.Index) : message.Index;
                else
                    object.Index = options.longs === String ? $util.Long.prototype.toString.call(message.Index) : options.longs === Number ? new $util.LongBits(message.Index.low >>> 0, message.Index.high >>> 0).toNumber() : message.Index;
            if (message.Members && message.Members.length) {
                object.Members = [];
                for (var j = 0; j < message.Members.length; ++j)
                    object.Members[j] = message.Members[j];
            }
            if (message.start != null && message.hasOwnProperty("start"))
                if (typeof message.start === "number")
                    object.start = options.longs === String ? String(message.start) : message.start;
                else
                    object.start = options.longs === String ? $util.Long.prototype.toString.call(message.start) : options.longs === Number ? new $util.LongBits(message.start.low >>> 0, message.start.high >>> 0).toNumber() : message.start;
            if (message.stop != null && message.hasOwnProperty("stop"))
                if (typeof message.stop === "number")
                    object.stop = options.longs === String ? String(message.stop) : message.stop;
                else
                    object.stop = options.longs === String ? $util.Long.prototype.toString.call(message.stop) : options.longs === Number ? new $util.LongBits(message.stop.low >>> 0, message.stop.high >>> 0).toNumber() : message.stop;
            if (message.min != null && message.hasOwnProperty("min"))
                object.min = options.json && !isFinite(message.min) ? String(message.min) : message.min;
            if (message.max != null && message.hasOwnProperty("max"))
                object.max = options.json && !isFinite(message.max) ? String(message.max) : message.max;
            if (message.Field != null && message.hasOwnProperty("Field"))
                object.Field = message.Field;
            if (message.IncrFloat != null && message.hasOwnProperty("IncrFloat"))
                object.IncrFloat = options.json && !isFinite(message.IncrFloat) ? String(message.IncrFloat) : message.IncrFloat;
            if (message.IncrInt != null && message.hasOwnProperty("IncrInt"))
                if (typeof message.IncrInt === "number")
                    object.IncrInt = options.longs === String ? String(message.IncrInt) : message.IncrInt;
                else
                    object.IncrInt = options.longs === String ? $util.Long.prototype.toString.call(message.IncrInt) : options.longs === Number ? new $util.LongBits(message.IncrInt.low >>> 0, message.IncrInt.high >>> 0).toNumber() : message.IncrInt;
            if (message.Before != null && message.hasOwnProperty("Before"))
                object.Before = message.Before;
            return object;
        };

        /**
         * Converts this Operation to JSON.
         * @function toJSON
         * @memberof pb.Operation
         * @instance
         * @returns {Object.<string,*>} JSON object
         */
        Operation.prototype.toJSON = function toJSON() {
            return this.constructor.toObject(this, $protobuf.util.toJSONOptions);
        };

        /**
         * Gets the default type url for Operation
         * @function getTypeUrl
         * @memberof pb.Operation
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        Operation.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/pb.Operation";
        };

        return Operation;
    })();

    pb.KeyScore = (function() {

        /**
         * Properties of a KeyScore.
         * @memberof pb
         * @interface IKeyScore
         * @property {string|null} [Member] KeyScore Member
         * @property {number|null} [Score] KeyScore Score
         */

        /**
         * Constructs a new KeyScore.
         * @memberof pb
         * @classdesc Represents a KeyScore.
         * @implements IKeyScore
         * @constructor
         * @param {pb.IKeyScore=} [properties] Properties to set
         */
        function KeyScore(properties) {
            if (properties)
                for (var keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null)
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * KeyScore Member.
         * @member {string} Member
         * @memberof pb.KeyScore
         * @instance
         */
        KeyScore.prototype.Member = "";

        /**
         * KeyScore Score.
         * @member {number} Score
         * @memberof pb.KeyScore
         * @instance
         */
        KeyScore.prototype.Score = 0;

        /**
         * Creates a new KeyScore instance using the specified properties.
         * @function create
         * @memberof pb.KeyScore
         * @static
         * @param {pb.IKeyScore=} [properties] Properties to set
         * @returns {pb.KeyScore} KeyScore instance
         */
        KeyScore.create = function create(properties) {
            return new KeyScore(properties);
        };

        /**
         * Encodes the specified KeyScore message. Does not implicitly {@link pb.KeyScore.verify|verify} messages.
         * @function encode
         * @memberof pb.KeyScore
         * @static
         * @param {pb.IKeyScore} message KeyScore message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        KeyScore.encode = function encode(message, writer) {
            if (!writer)
                writer = $Writer.create();
            if (message.Member != null && Object.hasOwnProperty.call(message, "Member"))
                writer.uint32(/* id 1, wireType 2 =*/10).string(message.Member);
            if (message.Score != null && Object.hasOwnProperty.call(message, "Score"))
                writer.uint32(/* id 2, wireType 1 =*/17).double(message.Score);
            return writer;
        };

        /**
         * Encodes the specified KeyScore message, length delimited. Does not implicitly {@link pb.KeyScore.verify|verify} messages.
         * @function encodeDelimited
         * @memberof pb.KeyScore
         * @static
         * @param {pb.IKeyScore} message KeyScore message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        KeyScore.encodeDelimited = function encodeDelimited(message, writer) {
            return this.encode(message, writer).ldelim();
        };

        /**
         * Decodes a KeyScore message from the specified reader or buffer.
         * @function decode
         * @memberof pb.KeyScore
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {pb.KeyScore} KeyScore
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        KeyScore.decode = function decode(reader, length) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            var end = length === undefined ? reader.len : reader.pos + length, message = new $root.pb.KeyScore();
            while (reader.pos < end) {
                var tag = reader.uint32();
                switch (tag >>> 3) {
                case 1: {
                        message.Member = reader.string();
                        break;
                    }
                case 2: {
                        message.Score = reader.double();
                        break;
                    }
                default:
                    reader.skipType(tag & 7);
                    break;
                }
            }
            return message;
        };

        /**
         * Decodes a KeyScore message from the specified reader or buffer, length delimited.
         * @function decodeDelimited
         * @memberof pb.KeyScore
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @returns {pb.KeyScore} KeyScore
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        KeyScore.decodeDelimited = function decodeDelimited(reader) {
            if (!(reader instanceof $Reader))
                reader = new $Reader(reader);
            return this.decode(reader, reader.uint32());
        };

        /**
         * Verifies a KeyScore message.
         * @function verify
         * @memberof pb.KeyScore
         * @static
         * @param {Object.<string,*>} message Plain object to verify
         * @returns {string|null} `null` if valid, otherwise the reason why it is not
         */
        KeyScore.verify = function verify(message) {
            if (typeof message !== "object" || message === null)
                return "object expected";
            if (message.Member != null && message.hasOwnProperty("Member"))
                if (!$util.isString(message.Member))
                    return "Member: string expected";
            if (message.Score != null && message.hasOwnProperty("Score"))
                if (typeof message.Score !== "number")
                    return "Score: number expected";
            return null;
        };

        /**
         * Creates a KeyScore message from a plain object. Also converts values to their respective internal types.
         * @function fromObject
         * @memberof pb.KeyScore
         * @static
         * @param {Object.<string,*>} object Plain object
         * @returns {pb.KeyScore} KeyScore
         */
        KeyScore.fromObject = function fromObject(object) {
            if (object instanceof $root.pb.KeyScore)
                return object;
            var message = new $root.pb.KeyScore();
            if (object.Member != null)
                message.Member = String(object.Member);
            if (object.Score != null)
                message.Score = Number(object.Score);
            return message;
        };

        /**
         * Creates a plain object from a KeyScore message. Also converts values to other types if specified.
         * @function toObject
         * @memberof pb.KeyScore
         * @static
         * @param {pb.KeyScore} message KeyScore
         * @param {$protobuf.IConversionOptions} [options] Conversion options
         * @returns {Object.<string,*>} Plain object
         */
        KeyScore.toObject = function toObject(message, options) {
            if (!options)
                options = {};
            var object = {};
            if (options.defaults) {
                object.Member = "";
                object.Score = 0;
            }
            if (message.Member != null && message.hasOwnProperty("Member"))
                object.Member = message.Member;
            if (message.Score != null && message.hasOwnProperty("Score"))
                object.Score = options.json && !isFinite(message.Score) ? String(message.Score) : message.Score;
            return object;
        };

        /**
         * Converts this KeyScore to JSON.
         * @function toJSON
         * @memberof pb.KeyScore
         * @instance
         * @returns {Object.<string,*>} JSON object
         */
        KeyScore.prototype.toJSON = function toJSON() {
            return this.constructor.toObject(this, $protobuf.util.toJSONOptions);
        };

        /**
         * Gets the default type url for KeyScore
         * @function getTypeUrl
         * @memberof pb.KeyScore
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        KeyScore.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/pb.KeyScore";
        };

        return KeyScore;
    })();

    pb.ZSetValue = (function() {

        /**
         * Properties of a ZSetValue.
         * @memberof pb
         * @interface IZSetValue
         * @property {Array.<pb.IKeyScore>|null} [Values] ZSetValue Values
         */

        /**
         * Constructs a new ZSetValue.
         * @memberof pb
         * @classdesc Represents a ZSetValue.
         * @implements IZSetValue
         * @constructor
         * @param {pb.IZSetValue=} [properties] Properties to set
         */
        function ZSetValue(properties) {
            this.Values = [];
            if (properties)
                for (var keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null)
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * ZSetValue Values.
         * @member {Array.<pb.IKeyScore>} Values
         * @memberof pb.ZSetValue
         * @instance
         */
        ZSetValue.prototype.Values = $util.emptyArray;

        /**
         * Creates a new ZSetValue instance using the specified properties.
         * @function create
         * @memberof pb.ZSetValue
         * @static
         * @param {pb.IZSetValue=} [properties] Properties to set
         * @returns {pb.ZSetValue} ZSetValue instance
         */
        ZSetValue.create = function create(properties) {
            return new ZSetValue(properties);
        };

        /**
         * Encodes the specified ZSetValue message. Does not implicitly {@link pb.ZSetValue.verify|verify} messages.
         * @function encode
         * @memberof pb.ZSetValue
         * @static
         * @param {pb.IZSetValue} message ZSetValue message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        ZSetValue.encode = function encode(message, writer) {
            if (!writer)
                writer = $Writer.create();
            if (message.Values != null && message.Values.length)
                for (var i = 0; i < message.Values.length; ++i)
                    $root.pb.KeyScore.encode(message.Values[i], writer.uint32(/* id 2, wireType 2 =*/18).fork()).ldelim();
            return writer;
        };

        /**
         * Encodes the specified ZSetValue message, length delimited. Does not implicitly {@link pb.ZSetValue.verify|verify} messages.
         * @function encodeDelimited
         * @memberof pb.ZSetValue
         * @static
         * @param {pb.IZSetValue} message ZSetValue message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        ZSetValue.encodeDelimited = function encodeDelimited(message, writer) {
            return this.encode(message, writer).ldelim();
        };

        /**
         * Decodes a ZSetValue message from the specified reader or buffer.
         * @function decode
         * @memberof pb.ZSetValue
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {pb.ZSetValue} ZSetValue
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        ZSetValue.decode = function decode(reader, length) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            var end = length === undefined ? reader.len : reader.pos + length, message = new $root.pb.ZSetValue();
            while (reader.pos < end) {
                var tag = reader.uint32();
                switch (tag >>> 3) {
                case 2: {
                        if (!(message.Values && message.Values.length))
                            message.Values = [];
                        message.Values.push($root.pb.KeyScore.decode(reader, reader.uint32()));
                        break;
                    }
                default:
                    reader.skipType(tag & 7);
                    break;
                }
            }
            return message;
        };

        /**
         * Decodes a ZSetValue message from the specified reader or buffer, length delimited.
         * @function decodeDelimited
         * @memberof pb.ZSetValue
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @returns {pb.ZSetValue} ZSetValue
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        ZSetValue.decodeDelimited = function decodeDelimited(reader) {
            if (!(reader instanceof $Reader))
                reader = new $Reader(reader);
            return this.decode(reader, reader.uint32());
        };

        /**
         * Verifies a ZSetValue message.
         * @function verify
         * @memberof pb.ZSetValue
         * @static
         * @param {Object.<string,*>} message Plain object to verify
         * @returns {string|null} `null` if valid, otherwise the reason why it is not
         */
        ZSetValue.verify = function verify(message) {
            if (typeof message !== "object" || message === null)
                return "object expected";
            if (message.Values != null && message.hasOwnProperty("Values")) {
                if (!Array.isArray(message.Values))
                    return "Values: array expected";
                for (var i = 0; i < message.Values.length; ++i) {
                    var error = $root.pb.KeyScore.verify(message.Values[i]);
                    if (error)
                        return "Values." + error;
                }
            }
            return null;
        };

        /**
         * Creates a ZSetValue message from a plain object. Also converts values to their respective internal types.
         * @function fromObject
         * @memberof pb.ZSetValue
         * @static
         * @param {Object.<string,*>} object Plain object
         * @returns {pb.ZSetValue} ZSetValue
         */
        ZSetValue.fromObject = function fromObject(object) {
            if (object instanceof $root.pb.ZSetValue)
                return object;
            var message = new $root.pb.ZSetValue();
            if (object.Values) {
                if (!Array.isArray(object.Values))
                    throw TypeError(".pb.ZSetValue.Values: array expected");
                message.Values = [];
                for (var i = 0; i < object.Values.length; ++i) {
                    if (typeof object.Values[i] !== "object")
                        throw TypeError(".pb.ZSetValue.Values: object expected");
                    message.Values[i] = $root.pb.KeyScore.fromObject(object.Values[i]);
                }
            }
            return message;
        };

        /**
         * Creates a plain object from a ZSetValue message. Also converts values to other types if specified.
         * @function toObject
         * @memberof pb.ZSetValue
         * @static
         * @param {pb.ZSetValue} message ZSetValue
         * @param {$protobuf.IConversionOptions} [options] Conversion options
         * @returns {Object.<string,*>} Plain object
         */
        ZSetValue.toObject = function toObject(message, options) {
            if (!options)
                options = {};
            var object = {};
            if (options.arrays || options.defaults)
                object.Values = [];
            if (message.Values && message.Values.length) {
                object.Values = [];
                for (var j = 0; j < message.Values.length; ++j)
                    object.Values[j] = $root.pb.KeyScore.toObject(message.Values[j], options);
            }
            return object;
        };

        /**
         * Converts this ZSetValue to JSON.
         * @function toJSON
         * @memberof pb.ZSetValue
         * @instance
         * @returns {Object.<string,*>} JSON object
         */
        ZSetValue.prototype.toJSON = function toJSON() {
            return this.constructor.toObject(this, $protobuf.util.toJSONOptions);
        };

        /**
         * Gets the default type url for ZSetValue
         * @function getTypeUrl
         * @memberof pb.ZSetValue
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        ZSetValue.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/pb.ZSetValue";
        };

        return ZSetValue;
    })();

    pb.ListValue = (function() {

        /**
         * Properties of a ListValue.
         * @memberof pb
         * @interface IListValue
         * @property {Array.<Uint8Array>|null} [Values] ListValue Values
         */

        /**
         * Constructs a new ListValue.
         * @memberof pb
         * @classdesc Represents a ListValue.
         * @implements IListValue
         * @constructor
         * @param {pb.IListValue=} [properties] Properties to set
         */
        function ListValue(properties) {
            this.Values = [];
            if (properties)
                for (var keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null)
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * ListValue Values.
         * @member {Array.<Uint8Array>} Values
         * @memberof pb.ListValue
         * @instance
         */
        ListValue.prototype.Values = $util.emptyArray;

        /**
         * Creates a new ListValue instance using the specified properties.
         * @function create
         * @memberof pb.ListValue
         * @static
         * @param {pb.IListValue=} [properties] Properties to set
         * @returns {pb.ListValue} ListValue instance
         */
        ListValue.create = function create(properties) {
            return new ListValue(properties);
        };

        /**
         * Encodes the specified ListValue message. Does not implicitly {@link pb.ListValue.verify|verify} messages.
         * @function encode
         * @memberof pb.ListValue
         * @static
         * @param {pb.IListValue} message ListValue message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        ListValue.encode = function encode(message, writer) {
            if (!writer)
                writer = $Writer.create();
            if (message.Values != null && message.Values.length)
                for (var i = 0; i < message.Values.length; ++i)
                    writer.uint32(/* id 2, wireType 2 =*/18).bytes(message.Values[i]);
            return writer;
        };

        /**
         * Encodes the specified ListValue message, length delimited. Does not implicitly {@link pb.ListValue.verify|verify} messages.
         * @function encodeDelimited
         * @memberof pb.ListValue
         * @static
         * @param {pb.IListValue} message ListValue message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        ListValue.encodeDelimited = function encodeDelimited(message, writer) {
            return this.encode(message, writer).ldelim();
        };

        /**
         * Decodes a ListValue message from the specified reader or buffer.
         * @function decode
         * @memberof pb.ListValue
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {pb.ListValue} ListValue
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        ListValue.decode = function decode(reader, length) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            var end = length === undefined ? reader.len : reader.pos + length, message = new $root.pb.ListValue();
            while (reader.pos < end) {
                var tag = reader.uint32();
                switch (tag >>> 3) {
                case 2: {
                        if (!(message.Values && message.Values.length))
                            message.Values = [];
                        message.Values.push(reader.bytes());
                        break;
                    }
                default:
                    reader.skipType(tag & 7);
                    break;
                }
            }
            return message;
        };

        /**
         * Decodes a ListValue message from the specified reader or buffer, length delimited.
         * @function decodeDelimited
         * @memberof pb.ListValue
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @returns {pb.ListValue} ListValue
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        ListValue.decodeDelimited = function decodeDelimited(reader) {
            if (!(reader instanceof $Reader))
                reader = new $Reader(reader);
            return this.decode(reader, reader.uint32());
        };

        /**
         * Verifies a ListValue message.
         * @function verify
         * @memberof pb.ListValue
         * @static
         * @param {Object.<string,*>} message Plain object to verify
         * @returns {string|null} `null` if valid, otherwise the reason why it is not
         */
        ListValue.verify = function verify(message) {
            if (typeof message !== "object" || message === null)
                return "object expected";
            if (message.Values != null && message.hasOwnProperty("Values")) {
                if (!Array.isArray(message.Values))
                    return "Values: array expected";
                for (var i = 0; i < message.Values.length; ++i)
                    if (!(message.Values[i] && typeof message.Values[i].length === "number" || $util.isString(message.Values[i])))
                        return "Values: buffer[] expected";
            }
            return null;
        };

        /**
         * Creates a ListValue message from a plain object. Also converts values to their respective internal types.
         * @function fromObject
         * @memberof pb.ListValue
         * @static
         * @param {Object.<string,*>} object Plain object
         * @returns {pb.ListValue} ListValue
         */
        ListValue.fromObject = function fromObject(object) {
            if (object instanceof $root.pb.ListValue)
                return object;
            var message = new $root.pb.ListValue();
            if (object.Values) {
                if (!Array.isArray(object.Values))
                    throw TypeError(".pb.ListValue.Values: array expected");
                message.Values = [];
                for (var i = 0; i < object.Values.length; ++i)
                    if (typeof object.Values[i] === "string")
                        $util.base64.decode(object.Values[i], message.Values[i] = $util.newBuffer($util.base64.length(object.Values[i])), 0);
                    else if (object.Values[i].length >= 0)
                        message.Values[i] = object.Values[i];
            }
            return message;
        };

        /**
         * Creates a plain object from a ListValue message. Also converts values to other types if specified.
         * @function toObject
         * @memberof pb.ListValue
         * @static
         * @param {pb.ListValue} message ListValue
         * @param {$protobuf.IConversionOptions} [options] Conversion options
         * @returns {Object.<string,*>} Plain object
         */
        ListValue.toObject = function toObject(message, options) {
            if (!options)
                options = {};
            var object = {};
            if (options.arrays || options.defaults)
                object.Values = [];
            if (message.Values && message.Values.length) {
                object.Values = [];
                for (var j = 0; j < message.Values.length; ++j)
                    object.Values[j] = options.bytes === String ? $util.base64.encode(message.Values[j], 0, message.Values[j].length) : options.bytes === Array ? Array.prototype.slice.call(message.Values[j]) : message.Values[j];
            }
            return object;
        };

        /**
         * Converts this ListValue to JSON.
         * @function toJSON
         * @memberof pb.ListValue
         * @instance
         * @returns {Object.<string,*>} JSON object
         */
        ListValue.prototype.toJSON = function toJSON() {
            return this.constructor.toObject(this, $protobuf.util.toJSONOptions);
        };

        /**
         * Gets the default type url for ListValue
         * @function getTypeUrl
         * @memberof pb.ListValue
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        ListValue.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/pb.ListValue";
        };

        return ListValue;
    })();

    pb.StringValue = (function() {

        /**
         * Properties of a StringValue.
         * @memberof pb
         * @interface IStringValue
         * @property {Uint8Array|null} [Value] StringValue Value
         */

        /**
         * Constructs a new StringValue.
         * @memberof pb
         * @classdesc Represents a StringValue.
         * @implements IStringValue
         * @constructor
         * @param {pb.IStringValue=} [properties] Properties to set
         */
        function StringValue(properties) {
            if (properties)
                for (var keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null)
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * StringValue Value.
         * @member {Uint8Array} Value
         * @memberof pb.StringValue
         * @instance
         */
        StringValue.prototype.Value = $util.newBuffer([]);

        /**
         * Creates a new StringValue instance using the specified properties.
         * @function create
         * @memberof pb.StringValue
         * @static
         * @param {pb.IStringValue=} [properties] Properties to set
         * @returns {pb.StringValue} StringValue instance
         */
        StringValue.create = function create(properties) {
            return new StringValue(properties);
        };

        /**
         * Encodes the specified StringValue message. Does not implicitly {@link pb.StringValue.verify|verify} messages.
         * @function encode
         * @memberof pb.StringValue
         * @static
         * @param {pb.IStringValue} message StringValue message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        StringValue.encode = function encode(message, writer) {
            if (!writer)
                writer = $Writer.create();
            if (message.Value != null && Object.hasOwnProperty.call(message, "Value"))
                writer.uint32(/* id 2, wireType 2 =*/18).bytes(message.Value);
            return writer;
        };

        /**
         * Encodes the specified StringValue message, length delimited. Does not implicitly {@link pb.StringValue.verify|verify} messages.
         * @function encodeDelimited
         * @memberof pb.StringValue
         * @static
         * @param {pb.IStringValue} message StringValue message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        StringValue.encodeDelimited = function encodeDelimited(message, writer) {
            return this.encode(message, writer).ldelim();
        };

        /**
         * Decodes a StringValue message from the specified reader or buffer.
         * @function decode
         * @memberof pb.StringValue
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {pb.StringValue} StringValue
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        StringValue.decode = function decode(reader, length) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            var end = length === undefined ? reader.len : reader.pos + length, message = new $root.pb.StringValue();
            while (reader.pos < end) {
                var tag = reader.uint32();
                switch (tag >>> 3) {
                case 2: {
                        message.Value = reader.bytes();
                        break;
                    }
                default:
                    reader.skipType(tag & 7);
                    break;
                }
            }
            return message;
        };

        /**
         * Decodes a StringValue message from the specified reader or buffer, length delimited.
         * @function decodeDelimited
         * @memberof pb.StringValue
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @returns {pb.StringValue} StringValue
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        StringValue.decodeDelimited = function decodeDelimited(reader) {
            if (!(reader instanceof $Reader))
                reader = new $Reader(reader);
            return this.decode(reader, reader.uint32());
        };

        /**
         * Verifies a StringValue message.
         * @function verify
         * @memberof pb.StringValue
         * @static
         * @param {Object.<string,*>} message Plain object to verify
         * @returns {string|null} `null` if valid, otherwise the reason why it is not
         */
        StringValue.verify = function verify(message) {
            if (typeof message !== "object" || message === null)
                return "object expected";
            if (message.Value != null && message.hasOwnProperty("Value"))
                if (!(message.Value && typeof message.Value.length === "number" || $util.isString(message.Value)))
                    return "Value: buffer expected";
            return null;
        };

        /**
         * Creates a StringValue message from a plain object. Also converts values to their respective internal types.
         * @function fromObject
         * @memberof pb.StringValue
         * @static
         * @param {Object.<string,*>} object Plain object
         * @returns {pb.StringValue} StringValue
         */
        StringValue.fromObject = function fromObject(object) {
            if (object instanceof $root.pb.StringValue)
                return object;
            var message = new $root.pb.StringValue();
            if (object.Value != null)
                if (typeof object.Value === "string")
                    $util.base64.decode(object.Value, message.Value = $util.newBuffer($util.base64.length(object.Value)), 0);
                else if (object.Value.length >= 0)
                    message.Value = object.Value;
            return message;
        };

        /**
         * Creates a plain object from a StringValue message. Also converts values to other types if specified.
         * @function toObject
         * @memberof pb.StringValue
         * @static
         * @param {pb.StringValue} message StringValue
         * @param {$protobuf.IConversionOptions} [options] Conversion options
         * @returns {Object.<string,*>} Plain object
         */
        StringValue.toObject = function toObject(message, options) {
            if (!options)
                options = {};
            var object = {};
            if (options.defaults)
                if (options.bytes === String)
                    object.Value = "";
                else {
                    object.Value = [];
                    if (options.bytes !== Array)
                        object.Value = $util.newBuffer(object.Value);
                }
            if (message.Value != null && message.hasOwnProperty("Value"))
                object.Value = options.bytes === String ? $util.base64.encode(message.Value, 0, message.Value.length) : options.bytes === Array ? Array.prototype.slice.call(message.Value) : message.Value;
            return object;
        };

        /**
         * Converts this StringValue to JSON.
         * @function toJSON
         * @memberof pb.StringValue
         * @instance
         * @returns {Object.<string,*>} JSON object
         */
        StringValue.prototype.toJSON = function toJSON() {
            return this.constructor.toObject(this, $protobuf.util.toJSONOptions);
        };

        /**
         * Gets the default type url for StringValue
         * @function getTypeUrl
         * @memberof pb.StringValue
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        StringValue.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/pb.StringValue";
        };

        return StringValue;
    })();

    pb.MemberBytes = (function() {

        /**
         * Properties of a MemberBytes.
         * @memberof pb
         * @interface IMemberBytes
         * @property {string|null} [Member] MemberBytes Member
         * @property {Uint8Array|null} [Value] MemberBytes Value
         */

        /**
         * Constructs a new MemberBytes.
         * @memberof pb
         * @classdesc Represents a MemberBytes.
         * @implements IMemberBytes
         * @constructor
         * @param {pb.IMemberBytes=} [properties] Properties to set
         */
        function MemberBytes(properties) {
            if (properties)
                for (var keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null)
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * MemberBytes Member.
         * @member {string} Member
         * @memberof pb.MemberBytes
         * @instance
         */
        MemberBytes.prototype.Member = "";

        /**
         * MemberBytes Value.
         * @member {Uint8Array} Value
         * @memberof pb.MemberBytes
         * @instance
         */
        MemberBytes.prototype.Value = $util.newBuffer([]);

        /**
         * Creates a new MemberBytes instance using the specified properties.
         * @function create
         * @memberof pb.MemberBytes
         * @static
         * @param {pb.IMemberBytes=} [properties] Properties to set
         * @returns {pb.MemberBytes} MemberBytes instance
         */
        MemberBytes.create = function create(properties) {
            return new MemberBytes(properties);
        };

        /**
         * Encodes the specified MemberBytes message. Does not implicitly {@link pb.MemberBytes.verify|verify} messages.
         * @function encode
         * @memberof pb.MemberBytes
         * @static
         * @param {pb.IMemberBytes} message MemberBytes message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        MemberBytes.encode = function encode(message, writer) {
            if (!writer)
                writer = $Writer.create();
            if (message.Member != null && Object.hasOwnProperty.call(message, "Member"))
                writer.uint32(/* id 1, wireType 2 =*/10).string(message.Member);
            if (message.Value != null && Object.hasOwnProperty.call(message, "Value"))
                writer.uint32(/* id 2, wireType 2 =*/18).bytes(message.Value);
            return writer;
        };

        /**
         * Encodes the specified MemberBytes message, length delimited. Does not implicitly {@link pb.MemberBytes.verify|verify} messages.
         * @function encodeDelimited
         * @memberof pb.MemberBytes
         * @static
         * @param {pb.IMemberBytes} message MemberBytes message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        MemberBytes.encodeDelimited = function encodeDelimited(message, writer) {
            return this.encode(message, writer).ldelim();
        };

        /**
         * Decodes a MemberBytes message from the specified reader or buffer.
         * @function decode
         * @memberof pb.MemberBytes
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {pb.MemberBytes} MemberBytes
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        MemberBytes.decode = function decode(reader, length) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            var end = length === undefined ? reader.len : reader.pos + length, message = new $root.pb.MemberBytes();
            while (reader.pos < end) {
                var tag = reader.uint32();
                switch (tag >>> 3) {
                case 1: {
                        message.Member = reader.string();
                        break;
                    }
                case 2: {
                        message.Value = reader.bytes();
                        break;
                    }
                default:
                    reader.skipType(tag & 7);
                    break;
                }
            }
            return message;
        };

        /**
         * Decodes a MemberBytes message from the specified reader or buffer, length delimited.
         * @function decodeDelimited
         * @memberof pb.MemberBytes
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @returns {pb.MemberBytes} MemberBytes
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        MemberBytes.decodeDelimited = function decodeDelimited(reader) {
            if (!(reader instanceof $Reader))
                reader = new $Reader(reader);
            return this.decode(reader, reader.uint32());
        };

        /**
         * Verifies a MemberBytes message.
         * @function verify
         * @memberof pb.MemberBytes
         * @static
         * @param {Object.<string,*>} message Plain object to verify
         * @returns {string|null} `null` if valid, otherwise the reason why it is not
         */
        MemberBytes.verify = function verify(message) {
            if (typeof message !== "object" || message === null)
                return "object expected";
            if (message.Member != null && message.hasOwnProperty("Member"))
                if (!$util.isString(message.Member))
                    return "Member: string expected";
            if (message.Value != null && message.hasOwnProperty("Value"))
                if (!(message.Value && typeof message.Value.length === "number" || $util.isString(message.Value)))
                    return "Value: buffer expected";
            return null;
        };

        /**
         * Creates a MemberBytes message from a plain object. Also converts values to their respective internal types.
         * @function fromObject
         * @memberof pb.MemberBytes
         * @static
         * @param {Object.<string,*>} object Plain object
         * @returns {pb.MemberBytes} MemberBytes
         */
        MemberBytes.fromObject = function fromObject(object) {
            if (object instanceof $root.pb.MemberBytes)
                return object;
            var message = new $root.pb.MemberBytes();
            if (object.Member != null)
                message.Member = String(object.Member);
            if (object.Value != null)
                if (typeof object.Value === "string")
                    $util.base64.decode(object.Value, message.Value = $util.newBuffer($util.base64.length(object.Value)), 0);
                else if (object.Value.length >= 0)
                    message.Value = object.Value;
            return message;
        };

        /**
         * Creates a plain object from a MemberBytes message. Also converts values to other types if specified.
         * @function toObject
         * @memberof pb.MemberBytes
         * @static
         * @param {pb.MemberBytes} message MemberBytes
         * @param {$protobuf.IConversionOptions} [options] Conversion options
         * @returns {Object.<string,*>} Plain object
         */
        MemberBytes.toObject = function toObject(message, options) {
            if (!options)
                options = {};
            var object = {};
            if (options.defaults) {
                object.Member = "";
                if (options.bytes === String)
                    object.Value = "";
                else {
                    object.Value = [];
                    if (options.bytes !== Array)
                        object.Value = $util.newBuffer(object.Value);
                }
            }
            if (message.Member != null && message.hasOwnProperty("Member"))
                object.Member = message.Member;
            if (message.Value != null && message.hasOwnProperty("Value"))
                object.Value = options.bytes === String ? $util.base64.encode(message.Value, 0, message.Value.length) : options.bytes === Array ? Array.prototype.slice.call(message.Value) : message.Value;
            return object;
        };

        /**
         * Converts this MemberBytes to JSON.
         * @function toJSON
         * @memberof pb.MemberBytes
         * @instance
         * @returns {Object.<string,*>} JSON object
         */
        MemberBytes.prototype.toJSON = function toJSON() {
            return this.constructor.toObject(this, $protobuf.util.toJSONOptions);
        };

        /**
         * Gets the default type url for MemberBytes
         * @function getTypeUrl
         * @memberof pb.MemberBytes
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        MemberBytes.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/pb.MemberBytes";
        };

        return MemberBytes;
    })();

    pb.SetValue = (function() {

        /**
         * Properties of a SetValue.
         * @memberof pb
         * @interface ISetValue
         * @property {Array.<string>|null} [Values] SetValue Values
         */

        /**
         * Constructs a new SetValue.
         * @memberof pb
         * @classdesc Represents a SetValue.
         * @implements ISetValue
         * @constructor
         * @param {pb.ISetValue=} [properties] Properties to set
         */
        function SetValue(properties) {
            this.Values = [];
            if (properties)
                for (var keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null)
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * SetValue Values.
         * @member {Array.<string>} Values
         * @memberof pb.SetValue
         * @instance
         */
        SetValue.prototype.Values = $util.emptyArray;

        /**
         * Creates a new SetValue instance using the specified properties.
         * @function create
         * @memberof pb.SetValue
         * @static
         * @param {pb.ISetValue=} [properties] Properties to set
         * @returns {pb.SetValue} SetValue instance
         */
        SetValue.create = function create(properties) {
            return new SetValue(properties);
        };

        /**
         * Encodes the specified SetValue message. Does not implicitly {@link pb.SetValue.verify|verify} messages.
         * @function encode
         * @memberof pb.SetValue
         * @static
         * @param {pb.ISetValue} message SetValue message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        SetValue.encode = function encode(message, writer) {
            if (!writer)
                writer = $Writer.create();
            if (message.Values != null && message.Values.length)
                for (var i = 0; i < message.Values.length; ++i)
                    writer.uint32(/* id 2, wireType 2 =*/18).string(message.Values[i]);
            return writer;
        };

        /**
         * Encodes the specified SetValue message, length delimited. Does not implicitly {@link pb.SetValue.verify|verify} messages.
         * @function encodeDelimited
         * @memberof pb.SetValue
         * @static
         * @param {pb.ISetValue} message SetValue message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        SetValue.encodeDelimited = function encodeDelimited(message, writer) {
            return this.encode(message, writer).ldelim();
        };

        /**
         * Decodes a SetValue message from the specified reader or buffer.
         * @function decode
         * @memberof pb.SetValue
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {pb.SetValue} SetValue
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        SetValue.decode = function decode(reader, length) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            var end = length === undefined ? reader.len : reader.pos + length, message = new $root.pb.SetValue();
            while (reader.pos < end) {
                var tag = reader.uint32();
                switch (tag >>> 3) {
                case 2: {
                        if (!(message.Values && message.Values.length))
                            message.Values = [];
                        message.Values.push(reader.string());
                        break;
                    }
                default:
                    reader.skipType(tag & 7);
                    break;
                }
            }
            return message;
        };

        /**
         * Decodes a SetValue message from the specified reader or buffer, length delimited.
         * @function decodeDelimited
         * @memberof pb.SetValue
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @returns {pb.SetValue} SetValue
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        SetValue.decodeDelimited = function decodeDelimited(reader) {
            if (!(reader instanceof $Reader))
                reader = new $Reader(reader);
            return this.decode(reader, reader.uint32());
        };

        /**
         * Verifies a SetValue message.
         * @function verify
         * @memberof pb.SetValue
         * @static
         * @param {Object.<string,*>} message Plain object to verify
         * @returns {string|null} `null` if valid, otherwise the reason why it is not
         */
        SetValue.verify = function verify(message) {
            if (typeof message !== "object" || message === null)
                return "object expected";
            if (message.Values != null && message.hasOwnProperty("Values")) {
                if (!Array.isArray(message.Values))
                    return "Values: array expected";
                for (var i = 0; i < message.Values.length; ++i)
                    if (!$util.isString(message.Values[i]))
                        return "Values: string[] expected";
            }
            return null;
        };

        /**
         * Creates a SetValue message from a plain object. Also converts values to their respective internal types.
         * @function fromObject
         * @memberof pb.SetValue
         * @static
         * @param {Object.<string,*>} object Plain object
         * @returns {pb.SetValue} SetValue
         */
        SetValue.fromObject = function fromObject(object) {
            if (object instanceof $root.pb.SetValue)
                return object;
            var message = new $root.pb.SetValue();
            if (object.Values) {
                if (!Array.isArray(object.Values))
                    throw TypeError(".pb.SetValue.Values: array expected");
                message.Values = [];
                for (var i = 0; i < object.Values.length; ++i)
                    message.Values[i] = String(object.Values[i]);
            }
            return message;
        };

        /**
         * Creates a plain object from a SetValue message. Also converts values to other types if specified.
         * @function toObject
         * @memberof pb.SetValue
         * @static
         * @param {pb.SetValue} message SetValue
         * @param {$protobuf.IConversionOptions} [options] Conversion options
         * @returns {Object.<string,*>} Plain object
         */
        SetValue.toObject = function toObject(message, options) {
            if (!options)
                options = {};
            var object = {};
            if (options.arrays || options.defaults)
                object.Values = [];
            if (message.Values && message.Values.length) {
                object.Values = [];
                for (var j = 0; j < message.Values.length; ++j)
                    object.Values[j] = message.Values[j];
            }
            return object;
        };

        /**
         * Converts this SetValue to JSON.
         * @function toJSON
         * @memberof pb.SetValue
         * @instance
         * @returns {Object.<string,*>} JSON object
         */
        SetValue.prototype.toJSON = function toJSON() {
            return this.constructor.toObject(this, $protobuf.util.toJSONOptions);
        };

        /**
         * Gets the default type url for SetValue
         * @function getTypeUrl
         * @memberof pb.SetValue
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        SetValue.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/pb.SetValue";
        };

        return SetValue;
    })();

    pb.HashValue = (function() {

        /**
         * Properties of a HashValue.
         * @memberof pb
         * @interface IHashValue
         * @property {Array.<pb.IMemberBytes>|null} [Values] HashValue Values
         */

        /**
         * Constructs a new HashValue.
         * @memberof pb
         * @classdesc Represents a HashValue.
         * @implements IHashValue
         * @constructor
         * @param {pb.IHashValue=} [properties] Properties to set
         */
        function HashValue(properties) {
            this.Values = [];
            if (properties)
                for (var keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null)
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * HashValue Values.
         * @member {Array.<pb.IMemberBytes>} Values
         * @memberof pb.HashValue
         * @instance
         */
        HashValue.prototype.Values = $util.emptyArray;

        /**
         * Creates a new HashValue instance using the specified properties.
         * @function create
         * @memberof pb.HashValue
         * @static
         * @param {pb.IHashValue=} [properties] Properties to set
         * @returns {pb.HashValue} HashValue instance
         */
        HashValue.create = function create(properties) {
            return new HashValue(properties);
        };

        /**
         * Encodes the specified HashValue message. Does not implicitly {@link pb.HashValue.verify|verify} messages.
         * @function encode
         * @memberof pb.HashValue
         * @static
         * @param {pb.IHashValue} message HashValue message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        HashValue.encode = function encode(message, writer) {
            if (!writer)
                writer = $Writer.create();
            if (message.Values != null && message.Values.length)
                for (var i = 0; i < message.Values.length; ++i)
                    $root.pb.MemberBytes.encode(message.Values[i], writer.uint32(/* id 2, wireType 2 =*/18).fork()).ldelim();
            return writer;
        };

        /**
         * Encodes the specified HashValue message, length delimited. Does not implicitly {@link pb.HashValue.verify|verify} messages.
         * @function encodeDelimited
         * @memberof pb.HashValue
         * @static
         * @param {pb.IHashValue} message HashValue message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        HashValue.encodeDelimited = function encodeDelimited(message, writer) {
            return this.encode(message, writer).ldelim();
        };

        /**
         * Decodes a HashValue message from the specified reader or buffer.
         * @function decode
         * @memberof pb.HashValue
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {pb.HashValue} HashValue
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        HashValue.decode = function decode(reader, length) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            var end = length === undefined ? reader.len : reader.pos + length, message = new $root.pb.HashValue();
            while (reader.pos < end) {
                var tag = reader.uint32();
                switch (tag >>> 3) {
                case 2: {
                        if (!(message.Values && message.Values.length))
                            message.Values = [];
                        message.Values.push($root.pb.MemberBytes.decode(reader, reader.uint32()));
                        break;
                    }
                default:
                    reader.skipType(tag & 7);
                    break;
                }
            }
            return message;
        };

        /**
         * Decodes a HashValue message from the specified reader or buffer, length delimited.
         * @function decodeDelimited
         * @memberof pb.HashValue
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @returns {pb.HashValue} HashValue
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        HashValue.decodeDelimited = function decodeDelimited(reader) {
            if (!(reader instanceof $Reader))
                reader = new $Reader(reader);
            return this.decode(reader, reader.uint32());
        };

        /**
         * Verifies a HashValue message.
         * @function verify
         * @memberof pb.HashValue
         * @static
         * @param {Object.<string,*>} message Plain object to verify
         * @returns {string|null} `null` if valid, otherwise the reason why it is not
         */
        HashValue.verify = function verify(message) {
            if (typeof message !== "object" || message === null)
                return "object expected";
            if (message.Values != null && message.hasOwnProperty("Values")) {
                if (!Array.isArray(message.Values))
                    return "Values: array expected";
                for (var i = 0; i < message.Values.length; ++i) {
                    var error = $root.pb.MemberBytes.verify(message.Values[i]);
                    if (error)
                        return "Values." + error;
                }
            }
            return null;
        };

        /**
         * Creates a HashValue message from a plain object. Also converts values to their respective internal types.
         * @function fromObject
         * @memberof pb.HashValue
         * @static
         * @param {Object.<string,*>} object Plain object
         * @returns {pb.HashValue} HashValue
         */
        HashValue.fromObject = function fromObject(object) {
            if (object instanceof $root.pb.HashValue)
                return object;
            var message = new $root.pb.HashValue();
            if (object.Values) {
                if (!Array.isArray(object.Values))
                    throw TypeError(".pb.HashValue.Values: array expected");
                message.Values = [];
                for (var i = 0; i < object.Values.length; ++i) {
                    if (typeof object.Values[i] !== "object")
                        throw TypeError(".pb.HashValue.Values: object expected");
                    message.Values[i] = $root.pb.MemberBytes.fromObject(object.Values[i]);
                }
            }
            return message;
        };

        /**
         * Creates a plain object from a HashValue message. Also converts values to other types if specified.
         * @function toObject
         * @memberof pb.HashValue
         * @static
         * @param {pb.HashValue} message HashValue
         * @param {$protobuf.IConversionOptions} [options] Conversion options
         * @returns {Object.<string,*>} Plain object
         */
        HashValue.toObject = function toObject(message, options) {
            if (!options)
                options = {};
            var object = {};
            if (options.arrays || options.defaults)
                object.Values = [];
            if (message.Values && message.Values.length) {
                object.Values = [];
                for (var j = 0; j < message.Values.length; ++j)
                    object.Values[j] = $root.pb.MemberBytes.toObject(message.Values[j], options);
            }
            return object;
        };

        /**
         * Converts this HashValue to JSON.
         * @function toJSON
         * @memberof pb.HashValue
         * @instance
         * @returns {Object.<string,*>} JSON object
         */
        HashValue.prototype.toJSON = function toJSON() {
            return this.constructor.toObject(this, $protobuf.util.toJSONOptions);
        };

        /**
         * Gets the default type url for HashValue
         * @function getTypeUrl
         * @memberof pb.HashValue
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        HashValue.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/pb.HashValue";
        };

        return HashValue;
    })();

    pb.Entity = (function() {

        /**
         * Properties of an Entity.
         * @memberof pb
         * @interface IEntity
         * @property {number|null} [Type] Entity Type
         * @property {string|null} [Key] Entity Key
         * @property {pb.IStringValue|null} [StringValue] Entity StringValue
         * @property {pb.IListValue|null} [ListValue] Entity ListValue
         * @property {pb.ISetValue|null} [SetValue] Entity SetValue
         * @property {pb.IHashValue|null} [HashValue] Entity HashValue
         * @property {pb.IZSetValue|null} [ZSetValue] Entity ZSetValue
         * @property {number|Long|null} [Expiration] Entity Expiration
         */

        /**
         * Constructs a new Entity.
         * @memberof pb
         * @classdesc Represents an Entity.
         * @implements IEntity
         * @constructor
         * @param {pb.IEntity=} [properties] Properties to set
         */
        function Entity(properties) {
            if (properties)
                for (var keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null)
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * Entity Type.
         * @member {number} Type
         * @memberof pb.Entity
         * @instance
         */
        Entity.prototype.Type = 0;

        /**
         * Entity Key.
         * @member {string} Key
         * @memberof pb.Entity
         * @instance
         */
        Entity.prototype.Key = "";

        /**
         * Entity StringValue.
         * @member {pb.IStringValue|null|undefined} StringValue
         * @memberof pb.Entity
         * @instance
         */
        Entity.prototype.StringValue = null;

        /**
         * Entity ListValue.
         * @member {pb.IListValue|null|undefined} ListValue
         * @memberof pb.Entity
         * @instance
         */
        Entity.prototype.ListValue = null;

        /**
         * Entity SetValue.
         * @member {pb.ISetValue|null|undefined} SetValue
         * @memberof pb.Entity
         * @instance
         */
        Entity.prototype.SetValue = null;

        /**
         * Entity HashValue.
         * @member {pb.IHashValue|null|undefined} HashValue
         * @memberof pb.Entity
         * @instance
         */
        Entity.prototype.HashValue = null;

        /**
         * Entity ZSetValue.
         * @member {pb.IZSetValue|null|undefined} ZSetValue
         * @memberof pb.Entity
         * @instance
         */
        Entity.prototype.ZSetValue = null;

        /**
         * Entity Expiration.
         * @member {number|Long} Expiration
         * @memberof pb.Entity
         * @instance
         */
        Entity.prototype.Expiration = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

        // OneOf field names bound to virtual getters and setters
        var $oneOfFields;

        /**
         * Entity Value.
         * @member {"StringValue"|"ListValue"|"SetValue"|"HashValue"|"ZSetValue"|undefined} Value
         * @memberof pb.Entity
         * @instance
         */
        Object.defineProperty(Entity.prototype, "Value", {
            get: $util.oneOfGetter($oneOfFields = ["StringValue", "ListValue", "SetValue", "HashValue", "ZSetValue"]),
            set: $util.oneOfSetter($oneOfFields)
        });

        /**
         * Creates a new Entity instance using the specified properties.
         * @function create
         * @memberof pb.Entity
         * @static
         * @param {pb.IEntity=} [properties] Properties to set
         * @returns {pb.Entity} Entity instance
         */
        Entity.create = function create(properties) {
            return new Entity(properties);
        };

        /**
         * Encodes the specified Entity message. Does not implicitly {@link pb.Entity.verify|verify} messages.
         * @function encode
         * @memberof pb.Entity
         * @static
         * @param {pb.IEntity} message Entity message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        Entity.encode = function encode(message, writer) {
            if (!writer)
                writer = $Writer.create();
            if (message.Type != null && Object.hasOwnProperty.call(message, "Type"))
                writer.uint32(/* id 1, wireType 0 =*/8).uint32(message.Type);
            if (message.Key != null && Object.hasOwnProperty.call(message, "Key"))
                writer.uint32(/* id 2, wireType 2 =*/18).string(message.Key);
            if (message.StringValue != null && Object.hasOwnProperty.call(message, "StringValue"))
                $root.pb.StringValue.encode(message.StringValue, writer.uint32(/* id 3, wireType 2 =*/26).fork()).ldelim();
            if (message.ListValue != null && Object.hasOwnProperty.call(message, "ListValue"))
                $root.pb.ListValue.encode(message.ListValue, writer.uint32(/* id 4, wireType 2 =*/34).fork()).ldelim();
            if (message.SetValue != null && Object.hasOwnProperty.call(message, "SetValue"))
                $root.pb.SetValue.encode(message.SetValue, writer.uint32(/* id 5, wireType 2 =*/42).fork()).ldelim();
            if (message.HashValue != null && Object.hasOwnProperty.call(message, "HashValue"))
                $root.pb.HashValue.encode(message.HashValue, writer.uint32(/* id 6, wireType 2 =*/50).fork()).ldelim();
            if (message.ZSetValue != null && Object.hasOwnProperty.call(message, "ZSetValue"))
                $root.pb.ZSetValue.encode(message.ZSetValue, writer.uint32(/* id 7, wireType 2 =*/58).fork()).ldelim();
            if (message.Expiration != null && Object.hasOwnProperty.call(message, "Expiration"))
                writer.uint32(/* id 8, wireType 0 =*/64).int64(message.Expiration);
            return writer;
        };

        /**
         * Encodes the specified Entity message, length delimited. Does not implicitly {@link pb.Entity.verify|verify} messages.
         * @function encodeDelimited
         * @memberof pb.Entity
         * @static
         * @param {pb.IEntity} message Entity message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        Entity.encodeDelimited = function encodeDelimited(message, writer) {
            return this.encode(message, writer).ldelim();
        };

        /**
         * Decodes an Entity message from the specified reader or buffer.
         * @function decode
         * @memberof pb.Entity
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {pb.Entity} Entity
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        Entity.decode = function decode(reader, length) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            var end = length === undefined ? reader.len : reader.pos + length, message = new $root.pb.Entity();
            while (reader.pos < end) {
                var tag = reader.uint32();
                switch (tag >>> 3) {
                case 1: {
                        message.Type = reader.uint32();
                        break;
                    }
                case 2: {
                        message.Key = reader.string();
                        break;
                    }
                case 3: {
                        message.StringValue = $root.pb.StringValue.decode(reader, reader.uint32());
                        break;
                    }
                case 4: {
                        message.ListValue = $root.pb.ListValue.decode(reader, reader.uint32());
                        break;
                    }
                case 5: {
                        message.SetValue = $root.pb.SetValue.decode(reader, reader.uint32());
                        break;
                    }
                case 6: {
                        message.HashValue = $root.pb.HashValue.decode(reader, reader.uint32());
                        break;
                    }
                case 7: {
                        message.ZSetValue = $root.pb.ZSetValue.decode(reader, reader.uint32());
                        break;
                    }
                case 8: {
                        message.Expiration = reader.int64();
                        break;
                    }
                default:
                    reader.skipType(tag & 7);
                    break;
                }
            }
            return message;
        };

        /**
         * Decodes an Entity message from the specified reader or buffer, length delimited.
         * @function decodeDelimited
         * @memberof pb.Entity
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @returns {pb.Entity} Entity
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        Entity.decodeDelimited = function decodeDelimited(reader) {
            if (!(reader instanceof $Reader))
                reader = new $Reader(reader);
            return this.decode(reader, reader.uint32());
        };

        /**
         * Verifies an Entity message.
         * @function verify
         * @memberof pb.Entity
         * @static
         * @param {Object.<string,*>} message Plain object to verify
         * @returns {string|null} `null` if valid, otherwise the reason why it is not
         */
        Entity.verify = function verify(message) {
            if (typeof message !== "object" || message === null)
                return "object expected";
            var properties = {};
            if (message.Type != null && message.hasOwnProperty("Type"))
                if (!$util.isInteger(message.Type))
                    return "Type: integer expected";
            if (message.Key != null && message.hasOwnProperty("Key"))
                if (!$util.isString(message.Key))
                    return "Key: string expected";
            if (message.StringValue != null && message.hasOwnProperty("StringValue")) {
                properties.Value = 1;
                {
                    var error = $root.pb.StringValue.verify(message.StringValue);
                    if (error)
                        return "StringValue." + error;
                }
            }
            if (message.ListValue != null && message.hasOwnProperty("ListValue")) {
                if (properties.Value === 1)
                    return "Value: multiple values";
                properties.Value = 1;
                {
                    var error = $root.pb.ListValue.verify(message.ListValue);
                    if (error)
                        return "ListValue." + error;
                }
            }
            if (message.SetValue != null && message.hasOwnProperty("SetValue")) {
                if (properties.Value === 1)
                    return "Value: multiple values";
                properties.Value = 1;
                {
                    var error = $root.pb.SetValue.verify(message.SetValue);
                    if (error)
                        return "SetValue." + error;
                }
            }
            if (message.HashValue != null && message.hasOwnProperty("HashValue")) {
                if (properties.Value === 1)
                    return "Value: multiple values";
                properties.Value = 1;
                {
                    var error = $root.pb.HashValue.verify(message.HashValue);
                    if (error)
                        return "HashValue." + error;
                }
            }
            if (message.ZSetValue != null && message.hasOwnProperty("ZSetValue")) {
                if (properties.Value === 1)
                    return "Value: multiple values";
                properties.Value = 1;
                {
                    var error = $root.pb.ZSetValue.verify(message.ZSetValue);
                    if (error)
                        return "ZSetValue." + error;
                }
            }
            if (message.Expiration != null && message.hasOwnProperty("Expiration"))
                if (!$util.isInteger(message.Expiration) && !(message.Expiration && $util.isInteger(message.Expiration.low) && $util.isInteger(message.Expiration.high)))
                    return "Expiration: integer|Long expected";
            return null;
        };

        /**
         * Creates an Entity message from a plain object. Also converts values to their respective internal types.
         * @function fromObject
         * @memberof pb.Entity
         * @static
         * @param {Object.<string,*>} object Plain object
         * @returns {pb.Entity} Entity
         */
        Entity.fromObject = function fromObject(object) {
            if (object instanceof $root.pb.Entity)
                return object;
            var message = new $root.pb.Entity();
            if (object.Type != null)
                message.Type = object.Type >>> 0;
            if (object.Key != null)
                message.Key = String(object.Key);
            if (object.StringValue != null) {
                if (typeof object.StringValue !== "object")
                    throw TypeError(".pb.Entity.StringValue: object expected");
                message.StringValue = $root.pb.StringValue.fromObject(object.StringValue);
            }
            if (object.ListValue != null) {
                if (typeof object.ListValue !== "object")
                    throw TypeError(".pb.Entity.ListValue: object expected");
                message.ListValue = $root.pb.ListValue.fromObject(object.ListValue);
            }
            if (object.SetValue != null) {
                if (typeof object.SetValue !== "object")
                    throw TypeError(".pb.Entity.SetValue: object expected");
                message.SetValue = $root.pb.SetValue.fromObject(object.SetValue);
            }
            if (object.HashValue != null) {
                if (typeof object.HashValue !== "object")
                    throw TypeError(".pb.Entity.HashValue: object expected");
                message.HashValue = $root.pb.HashValue.fromObject(object.HashValue);
            }
            if (object.ZSetValue != null) {
                if (typeof object.ZSetValue !== "object")
                    throw TypeError(".pb.Entity.ZSetValue: object expected");
                message.ZSetValue = $root.pb.ZSetValue.fromObject(object.ZSetValue);
            }
            if (object.Expiration != null)
                if ($util.Long)
                    (message.Expiration = $util.Long.fromValue(object.Expiration)).unsigned = false;
                else if (typeof object.Expiration === "string")
                    message.Expiration = parseInt(object.Expiration, 10);
                else if (typeof object.Expiration === "number")
                    message.Expiration = object.Expiration;
                else if (typeof object.Expiration === "object")
                    message.Expiration = new $util.LongBits(object.Expiration.low >>> 0, object.Expiration.high >>> 0).toNumber();
            return message;
        };

        /**
         * Creates a plain object from an Entity message. Also converts values to other types if specified.
         * @function toObject
         * @memberof pb.Entity
         * @static
         * @param {pb.Entity} message Entity
         * @param {$protobuf.IConversionOptions} [options] Conversion options
         * @returns {Object.<string,*>} Plain object
         */
        Entity.toObject = function toObject(message, options) {
            if (!options)
                options = {};
            var object = {};
            if (options.defaults) {
                object.Type = 0;
                object.Key = "";
                if ($util.Long) {
                    var long = new $util.Long(0, 0, false);
                    object.Expiration = options.longs === String ? long.toString() : options.longs === Number ? long.toNumber() : long;
                } else
                    object.Expiration = options.longs === String ? "0" : 0;
            }
            if (message.Type != null && message.hasOwnProperty("Type"))
                object.Type = message.Type;
            if (message.Key != null && message.hasOwnProperty("Key"))
                object.Key = message.Key;
            if (message.StringValue != null && message.hasOwnProperty("StringValue")) {
                object.StringValue = $root.pb.StringValue.toObject(message.StringValue, options);
                if (options.oneofs)
                    object.Value = "StringValue";
            }
            if (message.ListValue != null && message.hasOwnProperty("ListValue")) {
                object.ListValue = $root.pb.ListValue.toObject(message.ListValue, options);
                if (options.oneofs)
                    object.Value = "ListValue";
            }
            if (message.SetValue != null && message.hasOwnProperty("SetValue")) {
                object.SetValue = $root.pb.SetValue.toObject(message.SetValue, options);
                if (options.oneofs)
                    object.Value = "SetValue";
            }
            if (message.HashValue != null && message.hasOwnProperty("HashValue")) {
                object.HashValue = $root.pb.HashValue.toObject(message.HashValue, options);
                if (options.oneofs)
                    object.Value = "HashValue";
            }
            if (message.ZSetValue != null && message.hasOwnProperty("ZSetValue")) {
                object.ZSetValue = $root.pb.ZSetValue.toObject(message.ZSetValue, options);
                if (options.oneofs)
                    object.Value = "ZSetValue";
            }
            if (message.Expiration != null && message.hasOwnProperty("Expiration"))
                if (typeof message.Expiration === "number")
                    object.Expiration = options.longs === String ? String(message.Expiration) : message.Expiration;
                else
                    object.Expiration = options.longs === String ? $util.Long.prototype.toString.call(message.Expiration) : options.longs === Number ? new $util.LongBits(message.Expiration.low >>> 0, message.Expiration.high >>> 0).toNumber() : message.Expiration;
            return object;
        };

        /**
         * Converts this Entity to JSON.
         * @function toJSON
         * @memberof pb.Entity
         * @instance
         * @returns {Object.<string,*>} JSON object
         */
        Entity.prototype.toJSON = function toJSON() {
            return this.constructor.toObject(this, $protobuf.util.toJSONOptions);
        };

        /**
         * Gets the default type url for Entity
         * @function getTypeUrl
         * @memberof pb.Entity
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        Entity.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/pb.Entity";
        };

        return Entity;
    })();

    pb.Index = (function() {

        /**
         * Properties of an Index.
         * @memberof pb
         * @interface IIndex
         * @property {Array.<pb.Index.IItem>|null} [Items] Index Items
         */

        /**
         * Constructs a new Index.
         * @memberof pb
         * @classdesc Represents an Index.
         * @implements IIndex
         * @constructor
         * @param {pb.IIndex=} [properties] Properties to set
         */
        function Index(properties) {
            this.Items = [];
            if (properties)
                for (var keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null)
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * Index Items.
         * @member {Array.<pb.Index.IItem>} Items
         * @memberof pb.Index
         * @instance
         */
        Index.prototype.Items = $util.emptyArray;

        /**
         * Creates a new Index instance using the specified properties.
         * @function create
         * @memberof pb.Index
         * @static
         * @param {pb.IIndex=} [properties] Properties to set
         * @returns {pb.Index} Index instance
         */
        Index.create = function create(properties) {
            return new Index(properties);
        };

        /**
         * Encodes the specified Index message. Does not implicitly {@link pb.Index.verify|verify} messages.
         * @function encode
         * @memberof pb.Index
         * @static
         * @param {pb.IIndex} message Index message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        Index.encode = function encode(message, writer) {
            if (!writer)
                writer = $Writer.create();
            if (message.Items != null && message.Items.length)
                for (var i = 0; i < message.Items.length; ++i)
                    $root.pb.Index.Item.encode(message.Items[i], writer.uint32(/* id 1, wireType 2 =*/10).fork()).ldelim();
            return writer;
        };

        /**
         * Encodes the specified Index message, length delimited. Does not implicitly {@link pb.Index.verify|verify} messages.
         * @function encodeDelimited
         * @memberof pb.Index
         * @static
         * @param {pb.IIndex} message Index message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        Index.encodeDelimited = function encodeDelimited(message, writer) {
            return this.encode(message, writer).ldelim();
        };

        /**
         * Decodes an Index message from the specified reader or buffer.
         * @function decode
         * @memberof pb.Index
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {pb.Index} Index
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        Index.decode = function decode(reader, length) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            var end = length === undefined ? reader.len : reader.pos + length, message = new $root.pb.Index();
            while (reader.pos < end) {
                var tag = reader.uint32();
                switch (tag >>> 3) {
                case 1: {
                        if (!(message.Items && message.Items.length))
                            message.Items = [];
                        message.Items.push($root.pb.Index.Item.decode(reader, reader.uint32()));
                        break;
                    }
                default:
                    reader.skipType(tag & 7);
                    break;
                }
            }
            return message;
        };

        /**
         * Decodes an Index message from the specified reader or buffer, length delimited.
         * @function decodeDelimited
         * @memberof pb.Index
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @returns {pb.Index} Index
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        Index.decodeDelimited = function decodeDelimited(reader) {
            if (!(reader instanceof $Reader))
                reader = new $Reader(reader);
            return this.decode(reader, reader.uint32());
        };

        /**
         * Verifies an Index message.
         * @function verify
         * @memberof pb.Index
         * @static
         * @param {Object.<string,*>} message Plain object to verify
         * @returns {string|null} `null` if valid, otherwise the reason why it is not
         */
        Index.verify = function verify(message) {
            if (typeof message !== "object" || message === null)
                return "object expected";
            if (message.Items != null && message.hasOwnProperty("Items")) {
                if (!Array.isArray(message.Items))
                    return "Items: array expected";
                for (var i = 0; i < message.Items.length; ++i) {
                    var error = $root.pb.Index.Item.verify(message.Items[i]);
                    if (error)
                        return "Items." + error;
                }
            }
            return null;
        };

        /**
         * Creates an Index message from a plain object. Also converts values to their respective internal types.
         * @function fromObject
         * @memberof pb.Index
         * @static
         * @param {Object.<string,*>} object Plain object
         * @returns {pb.Index} Index
         */
        Index.fromObject = function fromObject(object) {
            if (object instanceof $root.pb.Index)
                return object;
            var message = new $root.pb.Index();
            if (object.Items) {
                if (!Array.isArray(object.Items))
                    throw TypeError(".pb.Index.Items: array expected");
                message.Items = [];
                for (var i = 0; i < object.Items.length; ++i) {
                    if (typeof object.Items[i] !== "object")
                        throw TypeError(".pb.Index.Items: object expected");
                    message.Items[i] = $root.pb.Index.Item.fromObject(object.Items[i]);
                }
            }
            return message;
        };

        /**
         * Creates a plain object from an Index message. Also converts values to other types if specified.
         * @function toObject
         * @memberof pb.Index
         * @static
         * @param {pb.Index} message Index
         * @param {$protobuf.IConversionOptions} [options] Conversion options
         * @returns {Object.<string,*>} Plain object
         */
        Index.toObject = function toObject(message, options) {
            if (!options)
                options = {};
            var object = {};
            if (options.arrays || options.defaults)
                object.Items = [];
            if (message.Items && message.Items.length) {
                object.Items = [];
                for (var j = 0; j < message.Items.length; ++j)
                    object.Items[j] = $root.pb.Index.Item.toObject(message.Items[j], options);
            }
            return object;
        };

        /**
         * Converts this Index to JSON.
         * @function toJSON
         * @memberof pb.Index
         * @instance
         * @returns {Object.<string,*>} JSON object
         */
        Index.prototype.toJSON = function toJSON() {
            return this.constructor.toObject(this, $protobuf.util.toJSONOptions);
        };

        /**
         * Gets the default type url for Index
         * @function getTypeUrl
         * @memberof pb.Index
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        Index.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/pb.Index";
        };

        Index.Item = (function() {

            /**
             * Properties of an Item.
             * @memberof pb.Index
             * @interface IItem
             * @property {string|null} [Key] Item Key
             * @property {Uint8Array|null} [Data] Item Data
             */

            /**
             * Constructs a new Item.
             * @memberof pb.Index
             * @classdesc Represents an Item.
             * @implements IItem
             * @constructor
             * @param {pb.Index.IItem=} [properties] Properties to set
             */
            function Item(properties) {
                if (properties)
                    for (var keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                        if (properties[keys[i]] != null)
                            this[keys[i]] = properties[keys[i]];
            }

            /**
             * Item Key.
             * @member {string} Key
             * @memberof pb.Index.Item
             * @instance
             */
            Item.prototype.Key = "";

            /**
             * Item Data.
             * @member {Uint8Array} Data
             * @memberof pb.Index.Item
             * @instance
             */
            Item.prototype.Data = $util.newBuffer([]);

            /**
             * Creates a new Item instance using the specified properties.
             * @function create
             * @memberof pb.Index.Item
             * @static
             * @param {pb.Index.IItem=} [properties] Properties to set
             * @returns {pb.Index.Item} Item instance
             */
            Item.create = function create(properties) {
                return new Item(properties);
            };

            /**
             * Encodes the specified Item message. Does not implicitly {@link pb.Index.Item.verify|verify} messages.
             * @function encode
             * @memberof pb.Index.Item
             * @static
             * @param {pb.Index.IItem} message Item message or plain object to encode
             * @param {$protobuf.Writer} [writer] Writer to encode to
             * @returns {$protobuf.Writer} Writer
             */
            Item.encode = function encode(message, writer) {
                if (!writer)
                    writer = $Writer.create();
                if (message.Key != null && Object.hasOwnProperty.call(message, "Key"))
                    writer.uint32(/* id 1, wireType 2 =*/10).string(message.Key);
                if (message.Data != null && Object.hasOwnProperty.call(message, "Data"))
                    writer.uint32(/* id 2, wireType 2 =*/18).bytes(message.Data);
                return writer;
            };

            /**
             * Encodes the specified Item message, length delimited. Does not implicitly {@link pb.Index.Item.verify|verify} messages.
             * @function encodeDelimited
             * @memberof pb.Index.Item
             * @static
             * @param {pb.Index.IItem} message Item message or plain object to encode
             * @param {$protobuf.Writer} [writer] Writer to encode to
             * @returns {$protobuf.Writer} Writer
             */
            Item.encodeDelimited = function encodeDelimited(message, writer) {
                return this.encode(message, writer).ldelim();
            };

            /**
             * Decodes an Item message from the specified reader or buffer.
             * @function decode
             * @memberof pb.Index.Item
             * @static
             * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
             * @param {number} [length] Message length if known beforehand
             * @returns {pb.Index.Item} Item
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            Item.decode = function decode(reader, length) {
                if (!(reader instanceof $Reader))
                    reader = $Reader.create(reader);
                var end = length === undefined ? reader.len : reader.pos + length, message = new $root.pb.Index.Item();
                while (reader.pos < end) {
                    var tag = reader.uint32();
                    switch (tag >>> 3) {
                    case 1: {
                            message.Key = reader.string();
                            break;
                        }
                    case 2: {
                            message.Data = reader.bytes();
                            break;
                        }
                    default:
                        reader.skipType(tag & 7);
                        break;
                    }
                }
                return message;
            };

            /**
             * Decodes an Item message from the specified reader or buffer, length delimited.
             * @function decodeDelimited
             * @memberof pb.Index.Item
             * @static
             * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
             * @returns {pb.Index.Item} Item
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            Item.decodeDelimited = function decodeDelimited(reader) {
                if (!(reader instanceof $Reader))
                    reader = new $Reader(reader);
                return this.decode(reader, reader.uint32());
            };

            /**
             * Verifies an Item message.
             * @function verify
             * @memberof pb.Index.Item
             * @static
             * @param {Object.<string,*>} message Plain object to verify
             * @returns {string|null} `null` if valid, otherwise the reason why it is not
             */
            Item.verify = function verify(message) {
                if (typeof message !== "object" || message === null)
                    return "object expected";
                if (message.Key != null && message.hasOwnProperty("Key"))
                    if (!$util.isString(message.Key))
                        return "Key: string expected";
                if (message.Data != null && message.hasOwnProperty("Data"))
                    if (!(message.Data && typeof message.Data.length === "number" || $util.isString(message.Data)))
                        return "Data: buffer expected";
                return null;
            };

            /**
             * Creates an Item message from a plain object. Also converts values to their respective internal types.
             * @function fromObject
             * @memberof pb.Index.Item
             * @static
             * @param {Object.<string,*>} object Plain object
             * @returns {pb.Index.Item} Item
             */
            Item.fromObject = function fromObject(object) {
                if (object instanceof $root.pb.Index.Item)
                    return object;
                var message = new $root.pb.Index.Item();
                if (object.Key != null)
                    message.Key = String(object.Key);
                if (object.Data != null)
                    if (typeof object.Data === "string")
                        $util.base64.decode(object.Data, message.Data = $util.newBuffer($util.base64.length(object.Data)), 0);
                    else if (object.Data.length >= 0)
                        message.Data = object.Data;
                return message;
            };

            /**
             * Creates a plain object from an Item message. Also converts values to other types if specified.
             * @function toObject
             * @memberof pb.Index.Item
             * @static
             * @param {pb.Index.Item} message Item
             * @param {$protobuf.IConversionOptions} [options] Conversion options
             * @returns {Object.<string,*>} Plain object
             */
            Item.toObject = function toObject(message, options) {
                if (!options)
                    options = {};
                var object = {};
                if (options.defaults) {
                    object.Key = "";
                    if (options.bytes === String)
                        object.Data = "";
                    else {
                        object.Data = [];
                        if (options.bytes !== Array)
                            object.Data = $util.newBuffer(object.Data);
                    }
                }
                if (message.Key != null && message.hasOwnProperty("Key"))
                    object.Key = message.Key;
                if (message.Data != null && message.hasOwnProperty("Data"))
                    object.Data = options.bytes === String ? $util.base64.encode(message.Data, 0, message.Data.length) : options.bytes === Array ? Array.prototype.slice.call(message.Data) : message.Data;
                return object;
            };

            /**
             * Converts this Item to JSON.
             * @function toJSON
             * @memberof pb.Index.Item
             * @instance
             * @returns {Object.<string,*>} JSON object
             */
            Item.prototype.toJSON = function toJSON() {
                return this.constructor.toObject(this, $protobuf.util.toJSONOptions);
            };

            /**
             * Gets the default type url for Item
             * @function getTypeUrl
             * @memberof pb.Index.Item
             * @static
             * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
             * @returns {string} The default type url
             */
            Item.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
                if (typeUrlPrefix === undefined) {
                    typeUrlPrefix = "type.googleapis.com";
                }
                return typeUrlPrefix + "/pb.Index.Item";
            };

            return Item;
        })();

        return Index;
    })();

    return pb;
})();

module.exports = $root;
