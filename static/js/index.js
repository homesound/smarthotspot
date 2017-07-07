var socket;

setupAngular();

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

function setupAngular() {
	var app = angular.module("homesound", ["ngRoute"]);
	app.config(function($routeProvider) {
		$routeProvider
		.when('/', {
			templateUrl: '/static/html/blank.html',
		})
		.when("/wifi", {
				templateUrl : "/static/html/wifi-configuration.html"
		})
	});
}
