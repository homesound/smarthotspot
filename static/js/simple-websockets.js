(function() {
	function print(msg) {
		console.log('[ws]: ' + msg);
	}

	function connect(addr) {
		if(!addr) {
			addr = 'ws://' + window.location.host + "/ws";
		}

		var ret = {};

		var ws = new WebSocket(addr);
		// Set up binary type for msgpack ease-of-use
		ws.binaryType = 'arraybuffer';

		ws.onopen = function(evt) {
			print("OPEN");
		}
		ws.onclose = function(evt) {
			print("CLOSE");
			ws = null;
		}
		ws.onmessage = function(evt) {
			debugger;
			var rawBinary = new Uint8Array(evt.data);
			var msg = msgpack.decode(rawBinary);
			print('Message: ' + msg.data);
			$(ret).trigger(msg.event, msg.data);
		}
		ws.onerror = function(evt) {
			print("ERROR: " + evt.data);
		}

		ret.on = function(evt, cb) {
			$(ret).on(evt, function(e, data) {
				cb(data);
			});
		}

		ret.emit = function(evt, data) {
			var msg = {
				event: evt,
				data: data,
			};
			var bytes = msgpack.encode(msg);
			ret._websocket.send(bytes);
		}

		ret._websocket = ws;
		return ret;
	}

	window.SimpleWebSocket = {
		connect: connect,
	}
})();
