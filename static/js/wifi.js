var socket;

window.onload = function() {
	$(".button-collapse").sideNav();


	socket = new SimpleWebSocket();
	socket.onopen = function() {
		// Check connectivity status
		// If we're running hotspot, then we show the wi-fi page
		socket.on('wifi-status', function(msg) {
			var json = JSON.parse(msg);
			if(json.hotspot) {
			}
		});

		socket.on('wifi-scan-results', function(msg) {
			console.log("Received scan results: " + JSON.stringify(msg))
			function __addScanResult(entry) {
				var html = '<li class="collection-item" data-interface="' + entry.interface + '" data-data-ssid="' + entry.SSID + '">' + entry.SSID + '</li>';
				var el = $($.parseHTML(html));
				el.on('click', function(e) {
					$('#wifi-ssid').val(entry.SSID);
				});
				$('#wifi-scan-results').append(el);
			}

			$('#wifi-scan-results').empty();

			var json = JSON.parse(JSON.stringify(msg));
			for(var i = 0; i < json.scanResults.length; i++) {
				var entry = json.scanResults[i];
				entry.interface = json.interface;
				__addScanResult(entry);
			}
			$('#wifi-scan-trigger').toggleClass("disabled", false);
		});

		$('#wifi-scan-trigger').on('click', function() {
			$(this).toggleClass("disabled");
			socket.emit('wifi-scan', "{}")
		});


		$('#wifi-connect').on('click', function() {
			var payload = {};
			payload.SSID = $('#wifi-ssid').val();
			payload.password = $('#wifi-password').val();

			console.log("Issuing wifi-connect event")
			socket.emit('wifi-connect', payload);
		});

		socket.emit('wifi-scan');
	}
	socket.connect();
}


