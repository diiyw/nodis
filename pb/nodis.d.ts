import * as $protobuf from "protobufjs";
import Long = require("long");
/** Namespace pb. */
export namespace pb {

    /** OpType enum. */
    enum OpType {
        None = 0,
        Clear = 1,
        Del = 2,
        Expire = 3,
        ExpireAt = 4,
        HClear = 5,
        HDel = 6,
        HIncrBy = 7,
        HIncrByFloat = 8,
        HMSet = 9,
        HSet = 10,
        HSetNX = 11,
        LInsert = 12,
        LPop = 13,
        LPopRPush = 14,
        LPush = 15,
        LPushX = 16,
        LRem = 17,
        LSet = 18,
        LTrim = 19,
        RPop = 20,
        RPopLPush = 21,
        RPush = 22,
        RPushX = 23,
        SAdd = 24,
        SRem = 25,
        Set = 26,
        ZAdd = 27,
        ZClear = 28,
        ZIncrBy = 29,
        ZRem = 30,
        ZRemRangeByRank = 31,
        ZRemRangeByScore = 32,
        Rename = 33
    }

    /** Properties of an Operation. */
    interface IOperation {

        /** Operation Type */
        Type?: (pb.OpType|null);

        /** Operation Key */
        Key?: (string|null);

        /** Operation Member */
        Member?: (string|null);

        /** Operation Value */
        Value?: (Uint8Array|null);

        /** Operation Expiration */
        Expiration?: (number|Long|null);

        /** Operation Score */
        Score?: (number|null);

        /** Operation Values */
        Values?: (Uint8Array[]|null);

        /** Operation DstKey */
        DstKey?: (string|null);

        /** Operation Pivot */
        Pivot?: (Uint8Array|null);

        /** Operation Count */
        Count?: (number|Long|null);

        /** Operation Index */
        Index?: (number|Long|null);

        /** Operation Members */
        Members?: (string[]|null);

        /** Operation start */
        start?: (number|Long|null);

        /** Operation stop */
        stop?: (number|Long|null);

        /** Operation min */
        min?: (number|null);

        /** Operation max */
        max?: (number|null);

        /** Operation Field */
        Field?: (string|null);

        /** Operation IncrFloat */
        IncrFloat?: (number|null);

        /** Operation IncrInt */
        IncrInt?: (number|Long|null);

        /** Operation Before */
        Before?: (boolean|null);
    }

    /** Represents an Operation. */
    class Operation implements IOperation {

        /**
         * Constructs a new Operation.
         * @param [properties] Properties to set
         */
        constructor(properties?: pb.IOperation);

        /** Operation Type. */
        public Type: pb.OpType;

        /** Operation Key. */
        public Key: string;

        /** Operation Member. */
        public Member: string;

        /** Operation Value. */
        public Value: Uint8Array;

        /** Operation Expiration. */
        public Expiration: (number|Long);

        /** Operation Score. */
        public Score: number;

        /** Operation Values. */
        public Values: Uint8Array[];

        /** Operation DstKey. */
        public DstKey: string;

        /** Operation Pivot. */
        public Pivot: Uint8Array;

        /** Operation Count. */
        public Count: (number|Long);

        /** Operation Index. */
        public Index: (number|Long);

        /** Operation Members. */
        public Members: string[];

        /** Operation start. */
        public start: (number|Long);

        /** Operation stop. */
        public stop: (number|Long);

        /** Operation min. */
        public min: number;

        /** Operation max. */
        public max: number;

        /** Operation Field. */
        public Field: string;

        /** Operation IncrFloat. */
        public IncrFloat: number;

        /** Operation IncrInt. */
        public IncrInt: (number|Long);

        /** Operation Before. */
        public Before: boolean;

        /**
         * Creates a new Operation instance using the specified properties.
         * @param [properties] Properties to set
         * @returns Operation instance
         */
        public static create(properties?: pb.IOperation): pb.Operation;

        /**
         * Encodes the specified Operation message. Does not implicitly {@link pb.Operation.verify|verify} messages.
         * @param message Operation message or plain object to encode
         * @param [writer] Writer to encode to
         * @returns Writer
         */
        public static encode(message: pb.IOperation, writer?: $protobuf.Writer): $protobuf.Writer;

        /**
         * Encodes the specified Operation message, length delimited. Does not implicitly {@link pb.Operation.verify|verify} messages.
         * @param message Operation message or plain object to encode
         * @param [writer] Writer to encode to
         * @returns Writer
         */
        public static encodeDelimited(message: pb.IOperation, writer?: $protobuf.Writer): $protobuf.Writer;

        /**
         * Decodes an Operation message from the specified reader or buffer.
         * @param reader Reader or buffer to decode from
         * @param [length] Message length if known beforehand
         * @returns Operation
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): pb.Operation;

        /**
         * Decodes an Operation message from the specified reader or buffer, length delimited.
         * @param reader Reader or buffer to decode from
         * @returns Operation
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        public static decodeDelimited(reader: ($protobuf.Reader|Uint8Array)): pb.Operation;

        /**
         * Verifies an Operation message.
         * @param message Plain object to verify
         * @returns `null` if valid, otherwise the reason why it is not
         */
        public static verify(message: { [k: string]: any }): (string|null);

        /**
         * Creates an Operation message from a plain object. Also converts values to their respective internal types.
         * @param object Plain object
         * @returns Operation
         */
        public static fromObject(object: { [k: string]: any }): pb.Operation;

        /**
         * Creates a plain object from an Operation message. Also converts values to other types if specified.
         * @param message Operation
         * @param [options] Conversion options
         * @returns Plain object
         */
        public static toObject(message: pb.Operation, options?: $protobuf.IConversionOptions): { [k: string]: any };

        /**
         * Converts this Operation to JSON.
         * @returns JSON object
         */
        public toJSON(): { [k: string]: any };

        /**
         * Gets the default type url for Operation
         * @param [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns The default type url
         */
        public static getTypeUrl(typeUrlPrefix?: string): string;
    }

    /** Properties of a KeyScore. */
    interface IKeyScore {

        /** KeyScore Member */
        Member?: (string|null);

        /** KeyScore Score */
        Score?: (number|null);
    }

    /** Represents a KeyScore. */
    class KeyScore implements IKeyScore {

        /**
         * Constructs a new KeyScore.
         * @param [properties] Properties to set
         */
        constructor(properties?: pb.IKeyScore);

        /** KeyScore Member. */
        public Member: string;

        /** KeyScore Score. */
        public Score: number;

        /**
         * Creates a new KeyScore instance using the specified properties.
         * @param [properties] Properties to set
         * @returns KeyScore instance
         */
        public static create(properties?: pb.IKeyScore): pb.KeyScore;

        /**
         * Encodes the specified KeyScore message. Does not implicitly {@link pb.KeyScore.verify|verify} messages.
         * @param message KeyScore message or plain object to encode
         * @param [writer] Writer to encode to
         * @returns Writer
         */
        public static encode(message: pb.IKeyScore, writer?: $protobuf.Writer): $protobuf.Writer;

        /**
         * Encodes the specified KeyScore message, length delimited. Does not implicitly {@link pb.KeyScore.verify|verify} messages.
         * @param message KeyScore message or plain object to encode
         * @param [writer] Writer to encode to
         * @returns Writer
         */
        public static encodeDelimited(message: pb.IKeyScore, writer?: $protobuf.Writer): $protobuf.Writer;

        /**
         * Decodes a KeyScore message from the specified reader or buffer.
         * @param reader Reader or buffer to decode from
         * @param [length] Message length if known beforehand
         * @returns KeyScore
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): pb.KeyScore;

        /**
         * Decodes a KeyScore message from the specified reader or buffer, length delimited.
         * @param reader Reader or buffer to decode from
         * @returns KeyScore
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        public static decodeDelimited(reader: ($protobuf.Reader|Uint8Array)): pb.KeyScore;

        /**
         * Verifies a KeyScore message.
         * @param message Plain object to verify
         * @returns `null` if valid, otherwise the reason why it is not
         */
        public static verify(message: { [k: string]: any }): (string|null);

        /**
         * Creates a KeyScore message from a plain object. Also converts values to their respective internal types.
         * @param object Plain object
         * @returns KeyScore
         */
        public static fromObject(object: { [k: string]: any }): pb.KeyScore;

        /**
         * Creates a plain object from a KeyScore message. Also converts values to other types if specified.
         * @param message KeyScore
         * @param [options] Conversion options
         * @returns Plain object
         */
        public static toObject(message: pb.KeyScore, options?: $protobuf.IConversionOptions): { [k: string]: any };

        /**
         * Converts this KeyScore to JSON.
         * @returns JSON object
         */
        public toJSON(): { [k: string]: any };

        /**
         * Gets the default type url for KeyScore
         * @param [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns The default type url
         */
        public static getTypeUrl(typeUrlPrefix?: string): string;
    }

    /** Properties of a ZSetValue. */
    interface IZSetValue {

        /** ZSetValue Values */
        Values?: (pb.IKeyScore[]|null);
    }

    /** Represents a ZSetValue. */
    class ZSetValue implements IZSetValue {

        /**
         * Constructs a new ZSetValue.
         * @param [properties] Properties to set
         */
        constructor(properties?: pb.IZSetValue);

        /** ZSetValue Values. */
        public Values: pb.IKeyScore[];

        /**
         * Creates a new ZSetValue instance using the specified properties.
         * @param [properties] Properties to set
         * @returns ZSetValue instance
         */
        public static create(properties?: pb.IZSetValue): pb.ZSetValue;

        /**
         * Encodes the specified ZSetValue message. Does not implicitly {@link pb.ZSetValue.verify|verify} messages.
         * @param message ZSetValue message or plain object to encode
         * @param [writer] Writer to encode to
         * @returns Writer
         */
        public static encode(message: pb.IZSetValue, writer?: $protobuf.Writer): $protobuf.Writer;

        /**
         * Encodes the specified ZSetValue message, length delimited. Does not implicitly {@link pb.ZSetValue.verify|verify} messages.
         * @param message ZSetValue message or plain object to encode
         * @param [writer] Writer to encode to
         * @returns Writer
         */
        public static encodeDelimited(message: pb.IZSetValue, writer?: $protobuf.Writer): $protobuf.Writer;

        /**
         * Decodes a ZSetValue message from the specified reader or buffer.
         * @param reader Reader or buffer to decode from
         * @param [length] Message length if known beforehand
         * @returns ZSetValue
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): pb.ZSetValue;

        /**
         * Decodes a ZSetValue message from the specified reader or buffer, length delimited.
         * @param reader Reader or buffer to decode from
         * @returns ZSetValue
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        public static decodeDelimited(reader: ($protobuf.Reader|Uint8Array)): pb.ZSetValue;

        /**
         * Verifies a ZSetValue message.
         * @param message Plain object to verify
         * @returns `null` if valid, otherwise the reason why it is not
         */
        public static verify(message: { [k: string]: any }): (string|null);

        /**
         * Creates a ZSetValue message from a plain object. Also converts values to their respective internal types.
         * @param object Plain object
         * @returns ZSetValue
         */
        public static fromObject(object: { [k: string]: any }): pb.ZSetValue;

        /**
         * Creates a plain object from a ZSetValue message. Also converts values to other types if specified.
         * @param message ZSetValue
         * @param [options] Conversion options
         * @returns Plain object
         */
        public static toObject(message: pb.ZSetValue, options?: $protobuf.IConversionOptions): { [k: string]: any };

        /**
         * Converts this ZSetValue to JSON.
         * @returns JSON object
         */
        public toJSON(): { [k: string]: any };

        /**
         * Gets the default type url for ZSetValue
         * @param [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns The default type url
         */
        public static getTypeUrl(typeUrlPrefix?: string): string;
    }

    /** Properties of a ListValue. */
    interface IListValue {

        /** ListValue Values */
        Values?: (Uint8Array[]|null);
    }

    /** Represents a ListValue. */
    class ListValue implements IListValue {

        /**
         * Constructs a new ListValue.
         * @param [properties] Properties to set
         */
        constructor(properties?: pb.IListValue);

        /** ListValue Values. */
        public Values: Uint8Array[];

        /**
         * Creates a new ListValue instance using the specified properties.
         * @param [properties] Properties to set
         * @returns ListValue instance
         */
        public static create(properties?: pb.IListValue): pb.ListValue;

        /**
         * Encodes the specified ListValue message. Does not implicitly {@link pb.ListValue.verify|verify} messages.
         * @param message ListValue message or plain object to encode
         * @param [writer] Writer to encode to
         * @returns Writer
         */
        public static encode(message: pb.IListValue, writer?: $protobuf.Writer): $protobuf.Writer;

        /**
         * Encodes the specified ListValue message, length delimited. Does not implicitly {@link pb.ListValue.verify|verify} messages.
         * @param message ListValue message or plain object to encode
         * @param [writer] Writer to encode to
         * @returns Writer
         */
        public static encodeDelimited(message: pb.IListValue, writer?: $protobuf.Writer): $protobuf.Writer;

        /**
         * Decodes a ListValue message from the specified reader or buffer.
         * @param reader Reader or buffer to decode from
         * @param [length] Message length if known beforehand
         * @returns ListValue
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): pb.ListValue;

        /**
         * Decodes a ListValue message from the specified reader or buffer, length delimited.
         * @param reader Reader or buffer to decode from
         * @returns ListValue
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        public static decodeDelimited(reader: ($protobuf.Reader|Uint8Array)): pb.ListValue;

        /**
         * Verifies a ListValue message.
         * @param message Plain object to verify
         * @returns `null` if valid, otherwise the reason why it is not
         */
        public static verify(message: { [k: string]: any }): (string|null);

        /**
         * Creates a ListValue message from a plain object. Also converts values to their respective internal types.
         * @param object Plain object
         * @returns ListValue
         */
        public static fromObject(object: { [k: string]: any }): pb.ListValue;

        /**
         * Creates a plain object from a ListValue message. Also converts values to other types if specified.
         * @param message ListValue
         * @param [options] Conversion options
         * @returns Plain object
         */
        public static toObject(message: pb.ListValue, options?: $protobuf.IConversionOptions): { [k: string]: any };

        /**
         * Converts this ListValue to JSON.
         * @returns JSON object
         */
        public toJSON(): { [k: string]: any };

        /**
         * Gets the default type url for ListValue
         * @param [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns The default type url
         */
        public static getTypeUrl(typeUrlPrefix?: string): string;
    }

    /** Properties of a StringValue. */
    interface IStringValue {

        /** StringValue Value */
        Value?: (Uint8Array|null);
    }

    /** Represents a StringValue. */
    class StringValue implements IStringValue {

        /**
         * Constructs a new StringValue.
         * @param [properties] Properties to set
         */
        constructor(properties?: pb.IStringValue);

        /** StringValue Value. */
        public Value: Uint8Array;

        /**
         * Creates a new StringValue instance using the specified properties.
         * @param [properties] Properties to set
         * @returns StringValue instance
         */
        public static create(properties?: pb.IStringValue): pb.StringValue;

        /**
         * Encodes the specified StringValue message. Does not implicitly {@link pb.StringValue.verify|verify} messages.
         * @param message StringValue message or plain object to encode
         * @param [writer] Writer to encode to
         * @returns Writer
         */
        public static encode(message: pb.IStringValue, writer?: $protobuf.Writer): $protobuf.Writer;

        /**
         * Encodes the specified StringValue message, length delimited. Does not implicitly {@link pb.StringValue.verify|verify} messages.
         * @param message StringValue message or plain object to encode
         * @param [writer] Writer to encode to
         * @returns Writer
         */
        public static encodeDelimited(message: pb.IStringValue, writer?: $protobuf.Writer): $protobuf.Writer;

        /**
         * Decodes a StringValue message from the specified reader or buffer.
         * @param reader Reader or buffer to decode from
         * @param [length] Message length if known beforehand
         * @returns StringValue
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): pb.StringValue;

        /**
         * Decodes a StringValue message from the specified reader or buffer, length delimited.
         * @param reader Reader or buffer to decode from
         * @returns StringValue
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        public static decodeDelimited(reader: ($protobuf.Reader|Uint8Array)): pb.StringValue;

        /**
         * Verifies a StringValue message.
         * @param message Plain object to verify
         * @returns `null` if valid, otherwise the reason why it is not
         */
        public static verify(message: { [k: string]: any }): (string|null);

        /**
         * Creates a StringValue message from a plain object. Also converts values to their respective internal types.
         * @param object Plain object
         * @returns StringValue
         */
        public static fromObject(object: { [k: string]: any }): pb.StringValue;

        /**
         * Creates a plain object from a StringValue message. Also converts values to other types if specified.
         * @param message StringValue
         * @param [options] Conversion options
         * @returns Plain object
         */
        public static toObject(message: pb.StringValue, options?: $protobuf.IConversionOptions): { [k: string]: any };

        /**
         * Converts this StringValue to JSON.
         * @returns JSON object
         */
        public toJSON(): { [k: string]: any };

        /**
         * Gets the default type url for StringValue
         * @param [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns The default type url
         */
        public static getTypeUrl(typeUrlPrefix?: string): string;
    }

    /** Properties of a MemberBytes. */
    interface IMemberBytes {

        /** MemberBytes Member */
        Member?: (string|null);

        /** MemberBytes Value */
        Value?: (Uint8Array|null);
    }

    /** Represents a MemberBytes. */
    class MemberBytes implements IMemberBytes {

        /**
         * Constructs a new MemberBytes.
         * @param [properties] Properties to set
         */
        constructor(properties?: pb.IMemberBytes);

        /** MemberBytes Member. */
        public Member: string;

        /** MemberBytes Value. */
        public Value: Uint8Array;

        /**
         * Creates a new MemberBytes instance using the specified properties.
         * @param [properties] Properties to set
         * @returns MemberBytes instance
         */
        public static create(properties?: pb.IMemberBytes): pb.MemberBytes;

        /**
         * Encodes the specified MemberBytes message. Does not implicitly {@link pb.MemberBytes.verify|verify} messages.
         * @param message MemberBytes message or plain object to encode
         * @param [writer] Writer to encode to
         * @returns Writer
         */
        public static encode(message: pb.IMemberBytes, writer?: $protobuf.Writer): $protobuf.Writer;

        /**
         * Encodes the specified MemberBytes message, length delimited. Does not implicitly {@link pb.MemberBytes.verify|verify} messages.
         * @param message MemberBytes message or plain object to encode
         * @param [writer] Writer to encode to
         * @returns Writer
         */
        public static encodeDelimited(message: pb.IMemberBytes, writer?: $protobuf.Writer): $protobuf.Writer;

        /**
         * Decodes a MemberBytes message from the specified reader or buffer.
         * @param reader Reader or buffer to decode from
         * @param [length] Message length if known beforehand
         * @returns MemberBytes
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): pb.MemberBytes;

        /**
         * Decodes a MemberBytes message from the specified reader or buffer, length delimited.
         * @param reader Reader or buffer to decode from
         * @returns MemberBytes
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        public static decodeDelimited(reader: ($protobuf.Reader|Uint8Array)): pb.MemberBytes;

        /**
         * Verifies a MemberBytes message.
         * @param message Plain object to verify
         * @returns `null` if valid, otherwise the reason why it is not
         */
        public static verify(message: { [k: string]: any }): (string|null);

        /**
         * Creates a MemberBytes message from a plain object. Also converts values to their respective internal types.
         * @param object Plain object
         * @returns MemberBytes
         */
        public static fromObject(object: { [k: string]: any }): pb.MemberBytes;

        /**
         * Creates a plain object from a MemberBytes message. Also converts values to other types if specified.
         * @param message MemberBytes
         * @param [options] Conversion options
         * @returns Plain object
         */
        public static toObject(message: pb.MemberBytes, options?: $protobuf.IConversionOptions): { [k: string]: any };

        /**
         * Converts this MemberBytes to JSON.
         * @returns JSON object
         */
        public toJSON(): { [k: string]: any };

        /**
         * Gets the default type url for MemberBytes
         * @param [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns The default type url
         */
        public static getTypeUrl(typeUrlPrefix?: string): string;
    }

    /** Properties of a SetValue. */
    interface ISetValue {

        /** SetValue Values */
        Values?: (string[]|null);
    }

    /** Represents a SetValue. */
    class SetValue implements ISetValue {

        /**
         * Constructs a new SetValue.
         * @param [properties] Properties to set
         */
        constructor(properties?: pb.ISetValue);

        /** SetValue Values. */
        public Values: string[];

        /**
         * Creates a new SetValue instance using the specified properties.
         * @param [properties] Properties to set
         * @returns SetValue instance
         */
        public static create(properties?: pb.ISetValue): pb.SetValue;

        /**
         * Encodes the specified SetValue message. Does not implicitly {@link pb.SetValue.verify|verify} messages.
         * @param message SetValue message or plain object to encode
         * @param [writer] Writer to encode to
         * @returns Writer
         */
        public static encode(message: pb.ISetValue, writer?: $protobuf.Writer): $protobuf.Writer;

        /**
         * Encodes the specified SetValue message, length delimited. Does not implicitly {@link pb.SetValue.verify|verify} messages.
         * @param message SetValue message or plain object to encode
         * @param [writer] Writer to encode to
         * @returns Writer
         */
        public static encodeDelimited(message: pb.ISetValue, writer?: $protobuf.Writer): $protobuf.Writer;

        /**
         * Decodes a SetValue message from the specified reader or buffer.
         * @param reader Reader or buffer to decode from
         * @param [length] Message length if known beforehand
         * @returns SetValue
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): pb.SetValue;

        /**
         * Decodes a SetValue message from the specified reader or buffer, length delimited.
         * @param reader Reader or buffer to decode from
         * @returns SetValue
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        public static decodeDelimited(reader: ($protobuf.Reader|Uint8Array)): pb.SetValue;

        /**
         * Verifies a SetValue message.
         * @param message Plain object to verify
         * @returns `null` if valid, otherwise the reason why it is not
         */
        public static verify(message: { [k: string]: any }): (string|null);

        /**
         * Creates a SetValue message from a plain object. Also converts values to their respective internal types.
         * @param object Plain object
         * @returns SetValue
         */
        public static fromObject(object: { [k: string]: any }): pb.SetValue;

        /**
         * Creates a plain object from a SetValue message. Also converts values to other types if specified.
         * @param message SetValue
         * @param [options] Conversion options
         * @returns Plain object
         */
        public static toObject(message: pb.SetValue, options?: $protobuf.IConversionOptions): { [k: string]: any };

        /**
         * Converts this SetValue to JSON.
         * @returns JSON object
         */
        public toJSON(): { [k: string]: any };

        /**
         * Gets the default type url for SetValue
         * @param [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns The default type url
         */
        public static getTypeUrl(typeUrlPrefix?: string): string;
    }

    /** Properties of a HashValue. */
    interface IHashValue {

        /** HashValue Values */
        Values?: (pb.IMemberBytes[]|null);
    }

    /** Represents a HashValue. */
    class HashValue implements IHashValue {

        /**
         * Constructs a new HashValue.
         * @param [properties] Properties to set
         */
        constructor(properties?: pb.IHashValue);

        /** HashValue Values. */
        public Values: pb.IMemberBytes[];

        /**
         * Creates a new HashValue instance using the specified properties.
         * @param [properties] Properties to set
         * @returns HashValue instance
         */
        public static create(properties?: pb.IHashValue): pb.HashValue;

        /**
         * Encodes the specified HashValue message. Does not implicitly {@link pb.HashValue.verify|verify} messages.
         * @param message HashValue message or plain object to encode
         * @param [writer] Writer to encode to
         * @returns Writer
         */
        public static encode(message: pb.IHashValue, writer?: $protobuf.Writer): $protobuf.Writer;

        /**
         * Encodes the specified HashValue message, length delimited. Does not implicitly {@link pb.HashValue.verify|verify} messages.
         * @param message HashValue message or plain object to encode
         * @param [writer] Writer to encode to
         * @returns Writer
         */
        public static encodeDelimited(message: pb.IHashValue, writer?: $protobuf.Writer): $protobuf.Writer;

        /**
         * Decodes a HashValue message from the specified reader or buffer.
         * @param reader Reader or buffer to decode from
         * @param [length] Message length if known beforehand
         * @returns HashValue
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): pb.HashValue;

        /**
         * Decodes a HashValue message from the specified reader or buffer, length delimited.
         * @param reader Reader or buffer to decode from
         * @returns HashValue
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        public static decodeDelimited(reader: ($protobuf.Reader|Uint8Array)): pb.HashValue;

        /**
         * Verifies a HashValue message.
         * @param message Plain object to verify
         * @returns `null` if valid, otherwise the reason why it is not
         */
        public static verify(message: { [k: string]: any }): (string|null);

        /**
         * Creates a HashValue message from a plain object. Also converts values to their respective internal types.
         * @param object Plain object
         * @returns HashValue
         */
        public static fromObject(object: { [k: string]: any }): pb.HashValue;

        /**
         * Creates a plain object from a HashValue message. Also converts values to other types if specified.
         * @param message HashValue
         * @param [options] Conversion options
         * @returns Plain object
         */
        public static toObject(message: pb.HashValue, options?: $protobuf.IConversionOptions): { [k: string]: any };

        /**
         * Converts this HashValue to JSON.
         * @returns JSON object
         */
        public toJSON(): { [k: string]: any };

        /**
         * Gets the default type url for HashValue
         * @param [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns The default type url
         */
        public static getTypeUrl(typeUrlPrefix?: string): string;
    }

    /** Properties of an Entity. */
    interface IEntity {

        /** Entity Type */
        Type?: (number|null);

        /** Entity Key */
        Key?: (string|null);

        /** Entity StringValue */
        StringValue?: (pb.IStringValue|null);

        /** Entity ListValue */
        ListValue?: (pb.IListValue|null);

        /** Entity SetValue */
        SetValue?: (pb.ISetValue|null);

        /** Entity HashValue */
        HashValue?: (pb.IHashValue|null);

        /** Entity ZSetValue */
        ZSetValue?: (pb.IZSetValue|null);

        /** Entity Expiration */
        Expiration?: (number|Long|null);
    }

    /** Represents an Entity. */
    class Entity implements IEntity {

        /**
         * Constructs a new Entity.
         * @param [properties] Properties to set
         */
        constructor(properties?: pb.IEntity);

        /** Entity Type. */
        public Type: number;

        /** Entity Key. */
        public Key: string;

        /** Entity StringValue. */
        public StringValue?: (pb.IStringValue|null);

        /** Entity ListValue. */
        public ListValue?: (pb.IListValue|null);

        /** Entity SetValue. */
        public SetValue?: (pb.ISetValue|null);

        /** Entity HashValue. */
        public HashValue?: (pb.IHashValue|null);

        /** Entity ZSetValue. */
        public ZSetValue?: (pb.IZSetValue|null);

        /** Entity Expiration. */
        public Expiration: (number|Long);

        /** Entity Value. */
        public Value?: ("StringValue"|"ListValue"|"SetValue"|"HashValue"|"ZSetValue");

        /**
         * Creates a new Entity instance using the specified properties.
         * @param [properties] Properties to set
         * @returns Entity instance
         */
        public static create(properties?: pb.IEntity): pb.Entity;

        /**
         * Encodes the specified Entity message. Does not implicitly {@link pb.Entity.verify|verify} messages.
         * @param message Entity message or plain object to encode
         * @param [writer] Writer to encode to
         * @returns Writer
         */
        public static encode(message: pb.IEntity, writer?: $protobuf.Writer): $protobuf.Writer;

        /**
         * Encodes the specified Entity message, length delimited. Does not implicitly {@link pb.Entity.verify|verify} messages.
         * @param message Entity message or plain object to encode
         * @param [writer] Writer to encode to
         * @returns Writer
         */
        public static encodeDelimited(message: pb.IEntity, writer?: $protobuf.Writer): $protobuf.Writer;

        /**
         * Decodes an Entity message from the specified reader or buffer.
         * @param reader Reader or buffer to decode from
         * @param [length] Message length if known beforehand
         * @returns Entity
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): pb.Entity;

        /**
         * Decodes an Entity message from the specified reader or buffer, length delimited.
         * @param reader Reader or buffer to decode from
         * @returns Entity
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        public static decodeDelimited(reader: ($protobuf.Reader|Uint8Array)): pb.Entity;

        /**
         * Verifies an Entity message.
         * @param message Plain object to verify
         * @returns `null` if valid, otherwise the reason why it is not
         */
        public static verify(message: { [k: string]: any }): (string|null);

        /**
         * Creates an Entity message from a plain object. Also converts values to their respective internal types.
         * @param object Plain object
         * @returns Entity
         */
        public static fromObject(object: { [k: string]: any }): pb.Entity;

        /**
         * Creates a plain object from an Entity message. Also converts values to other types if specified.
         * @param message Entity
         * @param [options] Conversion options
         * @returns Plain object
         */
        public static toObject(message: pb.Entity, options?: $protobuf.IConversionOptions): { [k: string]: any };

        /**
         * Converts this Entity to JSON.
         * @returns JSON object
         */
        public toJSON(): { [k: string]: any };

        /**
         * Gets the default type url for Entity
         * @param [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns The default type url
         */
        public static getTypeUrl(typeUrlPrefix?: string): string;
    }

    /** Properties of an Index. */
    interface IIndex {

        /** Index Items */
        Items?: (pb.Index.IItem[]|null);
    }

    /** Represents an Index. */
    class Index implements IIndex {

        /**
         * Constructs a new Index.
         * @param [properties] Properties to set
         */
        constructor(properties?: pb.IIndex);

        /** Index Items. */
        public Items: pb.Index.IItem[];

        /**
         * Creates a new Index instance using the specified properties.
         * @param [properties] Properties to set
         * @returns Index instance
         */
        public static create(properties?: pb.IIndex): pb.Index;

        /**
         * Encodes the specified Index message. Does not implicitly {@link pb.Index.verify|verify} messages.
         * @param message Index message or plain object to encode
         * @param [writer] Writer to encode to
         * @returns Writer
         */
        public static encode(message: pb.IIndex, writer?: $protobuf.Writer): $protobuf.Writer;

        /**
         * Encodes the specified Index message, length delimited. Does not implicitly {@link pb.Index.verify|verify} messages.
         * @param message Index message or plain object to encode
         * @param [writer] Writer to encode to
         * @returns Writer
         */
        public static encodeDelimited(message: pb.IIndex, writer?: $protobuf.Writer): $protobuf.Writer;

        /**
         * Decodes an Index message from the specified reader or buffer.
         * @param reader Reader or buffer to decode from
         * @param [length] Message length if known beforehand
         * @returns Index
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): pb.Index;

        /**
         * Decodes an Index message from the specified reader or buffer, length delimited.
         * @param reader Reader or buffer to decode from
         * @returns Index
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        public static decodeDelimited(reader: ($protobuf.Reader|Uint8Array)): pb.Index;

        /**
         * Verifies an Index message.
         * @param message Plain object to verify
         * @returns `null` if valid, otherwise the reason why it is not
         */
        public static verify(message: { [k: string]: any }): (string|null);

        /**
         * Creates an Index message from a plain object. Also converts values to their respective internal types.
         * @param object Plain object
         * @returns Index
         */
        public static fromObject(object: { [k: string]: any }): pb.Index;

        /**
         * Creates a plain object from an Index message. Also converts values to other types if specified.
         * @param message Index
         * @param [options] Conversion options
         * @returns Plain object
         */
        public static toObject(message: pb.Index, options?: $protobuf.IConversionOptions): { [k: string]: any };

        /**
         * Converts this Index to JSON.
         * @returns JSON object
         */
        public toJSON(): { [k: string]: any };

        /**
         * Gets the default type url for Index
         * @param [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns The default type url
         */
        public static getTypeUrl(typeUrlPrefix?: string): string;
    }

    namespace Index {

        /** Properties of an Item. */
        interface IItem {

            /** Item Key */
            Key?: (string|null);

            /** Item Data */
            Data?: (Uint8Array|null);
        }

        /** Represents an Item. */
        class Item implements IItem {

            /**
             * Constructs a new Item.
             * @param [properties] Properties to set
             */
            constructor(properties?: pb.Index.IItem);

            /** Item Key. */
            public Key: string;

            /** Item Data. */
            public Data: Uint8Array;

            /**
             * Creates a new Item instance using the specified properties.
             * @param [properties] Properties to set
             * @returns Item instance
             */
            public static create(properties?: pb.Index.IItem): pb.Index.Item;

            /**
             * Encodes the specified Item message. Does not implicitly {@link pb.Index.Item.verify|verify} messages.
             * @param message Item message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: pb.Index.IItem, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Encodes the specified Item message, length delimited. Does not implicitly {@link pb.Index.Item.verify|verify} messages.
             * @param message Item message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encodeDelimited(message: pb.Index.IItem, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes an Item message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns Item
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): pb.Index.Item;

            /**
             * Decodes an Item message from the specified reader or buffer, length delimited.
             * @param reader Reader or buffer to decode from
             * @returns Item
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decodeDelimited(reader: ($protobuf.Reader|Uint8Array)): pb.Index.Item;

            /**
             * Verifies an Item message.
             * @param message Plain object to verify
             * @returns `null` if valid, otherwise the reason why it is not
             */
            public static verify(message: { [k: string]: any }): (string|null);

            /**
             * Creates an Item message from a plain object. Also converts values to their respective internal types.
             * @param object Plain object
             * @returns Item
             */
            public static fromObject(object: { [k: string]: any }): pb.Index.Item;

            /**
             * Creates a plain object from an Item message. Also converts values to other types if specified.
             * @param message Item
             * @param [options] Conversion options
             * @returns Plain object
             */
            public static toObject(message: pb.Index.Item, options?: $protobuf.IConversionOptions): { [k: string]: any };

            /**
             * Converts this Item to JSON.
             * @returns JSON object
             */
            public toJSON(): { [k: string]: any };

            /**
             * Gets the default type url for Item
             * @param [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
             * @returns The default type url
             */
            public static getTypeUrl(typeUrlPrefix?: string): string;
        }
    }
}
