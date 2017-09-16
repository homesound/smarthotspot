window.SimpleWebSocket = function() {
	function print(msg) {
		console.log('[ws]: ' + msg);
	}

	var constructor = function() {
		var socket = {};
		socket.connect = function(addr) {
			if(!addr) {
				addr = 'ws://' + window.location.host + "/ws";
			}

			var ws = new WebSocket(addr);
			// Set up binary type for msgpack ease-of-use
			ws.binaryType = 'arraybuffer';

			ws.onopen = function(evt) {
				print("OPEN");
				if(socket.onopen) {
					socket.onopen(evt);
				}
			}
			ws.onclose = function(evt) {
				print("CLOSE");
				if(socket.onclose) {
					socket.onclose(evt);
				}
			}
			ws.onmessage = function(evt) {
				var rawBinary = new Uint8Array(evt.data);
				var msg = msgpack.decode(rawBinary);
				print('Message: ' + msg.data);
				$(socket).trigger(msg.event, msg.data);
				if(socket.onmessage) {
					socket.onmessage(evt);
				}
			}
			ws.onerror = function(evt) {
				print("ERROR: " + evt.data);
				if(socket.onerror) {
					socket.onerror(evt);
				}
			}

			socket.on = function(evt, cb) {
				$(socket).on(evt, function(e, data) {
					cb(data);
				});
			}

			socket.emit = function(evt, data) {
				var msg = {
					event: evt,
					data: data,
				};
				var bytes = msgpack.encode(msg);
				socket._websocket.send(bytes);
			}

			socket._websocket = ws;
		}
		return socket;
	}
	console.log('SimpleWebSocket loaded');
	return constructor;
}();
