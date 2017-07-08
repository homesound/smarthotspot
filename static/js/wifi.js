var socket;

window.onload = function() {
	$(".button-collapse").sideNav();


	socket = io.connect();

	// Check connectivity status
	// If we're running hotspot, then we show the wi-fi page
	socket.on('wifi-status', function(msg) {
		var json = JSON.parse(msg);
		if(json.hotspot) {
		}
	});

	socket.emit('wifi-status');
}

socket.on('wifi-scan-results', function(msg) {
	console.log("Received scan results")
	function __addScanResult(entry) {
		var html = '<li class="collection-item" data-interface="' + entry.interface + '" data-data-ssid="' + entry.SSID + '">' + entry.SSID + '</li>';
		var el = $($.parseHTML(html));
		el.on('click', function(e) {
			$('#wifi-ssid').val(entry.SSID);
		});
		$('#wifi-scan-results').append(el);
	}

	$('#wifi-scan-results').empty();

	var json = JSON.parse(msg);
	for(var i = 0; i < json.length; i++) {
		var entry = json[i];
		for(var j = 0; j < entry.scanResults.length; j++) {
			entry.scanResults[j].interface = entry.interface;
			__addScanResult(entry.scanResults[j]);
		}
	}
	$('#wifi-scan-trigger').toggleClass("disabled", false);
});

socket.emit('wifi-scan');

$('#wifi-scan-trigger').on('click', function() {
	$(this).toggleClass("disabled");
	socket.emit('wifi-scan', "{}")
});


$('#wifi-connect').on('click', function() {
	var payload = {};
	payload.SSID = $('#wifi-ssid').val();
	payload.password = $('#wifi-password').val();

	console.log("Issuing wifi-connect event")
	socket.emit('wifi-connect', JSON.stringify(payload));
});
