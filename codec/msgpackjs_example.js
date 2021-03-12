// js消息支持 msgpack-lite 库
// https://github.com/kawanet/msgpack-lite#extension-types
// 0x1B	Buffer

// js发送消息
// type(String): 消息类型
// data(Object): 消息内容
function send(type, data) {
    // _.values依赖lodash库
    var data = msgpack.encode(_.values(data), codec)
    data = msgpack.encode([type, data], codec)
    ws.send(data)
}

// 支持time格式，go处理时间为int类型
function getUint64(bytes, littleEndian) {
    var low = 4,
        high = 0

    if (littleEndian) {
        low = 0
        high = 4
    }

    var dv = new DataView(Uint8Array.from(bytes).buffer);

    return (dv.getUint32(high, littleEndian) << 32) |
        dv.getUint32(low, littleEndian)
}

var codecWithGoTime = msgpack.createCodec()
codecWithGoTime.addExtUnpacker(0xFF, goTimeUnpacker)
function goTimeUnpacker(data) {
    return getUint64(data, false)
}

// 接收消息处理
ws.onmessage = function (event) {
    // event.data是ArrayBuffer类型
    if (typeof event.data == 'object') {
        // 解开包装
        array = msgpack.decode(new Uint8Array(event.data))
        if (array.length != 2) {
            console.error("msg warpper not 2")
            return
        }

        msg = {
            type: array[0],
            data: msgpack.decode(array[1], {
                codec: codecWithGoTime
            })
        }
    }
}

/*
 格式化消息，方便使用

例如：go发送
    type Custom struct {
        None bool
    }
    type foo struct {
        Int int
        Array []int
        String string
        Custom Custom
        Customs []Custom
    }
    data = foo{Int: 11, Array: []int{1, 2, 3}, String: "bala", Custom: Custom{None: true}, Customs: []Customs{
        Custom{None: false}, Custom{None: true}
    }}

    js格式化消息
    var msg = formatMessage(["Int", "Array", "String", ["Custom", ["None"]], ["Customs", ["None"], "array"]], data)
    console.log(msg.Int, msg.Array, msg.String, msg.Custom, msg.Customs)
    输出：11 [1,2,3] bala {None: true} [{None: false}, {None: true}]
*/


function formatMessage(struct, data) {
    if (struct.length != data.length) {
        console.error("length not equal", struct.length, data.length)
        console.log("data", data)
        console.log("struct", struct)
        return
    }

    var msg = {}
    _.forEach(struct, function (item, index) {
        if (typeof item == 'string') {
            msg[item] = data[index]
        } else if (item instanceof Array) {
            if (item.length < 2 || item.length > 3) {
                console.error("struct format wrong:\n\tformat1: ['ArrayName', ['ArrayItemName' ...]]\n\tformat2: ['ArrayName', ['ArrayItemName' ...], 'array']")
                return
            }
            if (item.length == 3 && item[2] !== 'array') {
                console.error("struct format wrong: only support array\n\t['ArrayName', ['ArrayItemName' ...], 'array']")
                return
            }

            if (item.length == 3) {
                var array = []
                _.forEach(data[index], function(dataItem) {
                    var tmp = formatMessage(item[1], dataItem)
                    array.push(tmp)
                })
                msg[item[0]] = array
            } else {
                msg[item[0]] = formatMessage(item[1], data[index])
            }
        }
    })
    return msg
}