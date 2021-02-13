
/**

Butly (Modified)

**/

var control_panel = process.env.CP_SERVER;
var link_redirect = process.env.LR_SERVER;
var heroku_app = "https://butly.herokuapp.com";

var cookieParser = require('cookie-parser');
var crypto = require('crypto');
var express = require('express');
var exphbs = require('express-handlebars');
var fs = require('fs');
var Client = require('node-rest-client').Client;

var app = express();

app.use(express.json());
app.use(express.urlencoded({ extended: true }));

app.use("/images", express.static(__dirname + '/images'));
app.engine('handlebars', exphbs());
app.set('view engine', 'handlebars');
var os = require("os");
var hostname = os.hostname();

var secretKey = process.env.SECRET_KEY ;

var get_hash = function( ts ) {

	// http://nodejs.org/api/crypto.html#crypto_crypto_createhmac_algorithm_key
	text = "make_shortlink" + "|" + ts + "|" + secretKey ;
	hmac = crypto.createHmac("sha256", secretKey);
	hmac.setEncoding('base64');
	hmac.write(text);
	hmac.end() ;
	hash = hmac.read();
	//console.log( "HASH: " + hash )
	return hash ;

}

var error = function( req, res, msg, ts ) {

	var result = new Object() ;
	state = "error" ;
	hash = get_hash ( state, ts ) ;

	result.msg = msg ;
	result.ts = ts ;
	result.hash = hash ;

	res.render('bitly', {
	    ts: result.ts,
	    hash: result.hash,
	    message: result.msg
	});

}

var page = function( req, res, ts, server, short_link ) {

	var result = new Object() ;
	hash = get_hash ( ts ) ;

	var client = new Client();
	var title = "";
	client.get( server + "/ping",
		function(data, response_raw){
			var data_json = JSON.parse(data) ;
			console.log(data);
			title = data_json.Test;
      console.log( "title = " + title ) ;
      var msg =   "\n\nButly, Inc.\n\nCloud-Enabled Link-Shortener\n" +
                  title + "\n" +
                  "Running on: " + hostname + "\n" ;
	    if ( short_link ) {
	      var msg2 = "Try out your new shortlink!" ;
			}
      result.msg = msg ;
      result.ts = ts ;
      result.hash = hash ;
			res.render('bitly', {
				ts: result.ts,
	      hash: result.hash,
	      message: result.msg,
				message2: msg2,
				link: short_link
			});
	  });
}

var handle_post = function (req, res, next) {

	console.log( "Post: " + "Action: " +  req.body.event + "URL: " + req.body.orig_url + "\n" ) ;
	console.log(req.body) ;
	var body_msg = req.body.message ;
	var hash1 = "" + req.body.hash ;
	var action = "" + req.body.event ;
	var orig_url = "" + req.body.orig_url ;
	var ts = parseInt(req.body.ts) ;
	var now = new Date().getTime() ;
	var diff = ((now - ts)/1000) ;
	hash2 = get_hash ( ts ) ;
	console.log( "DIFF:  " +  diff ) ;
	console.log( "HASH1: " + hash1 ) ;
	console.log( "HASH2: " + hash2 ) ;

	if ( orig_url.length == 0 ) {
		error( req, res, "*** INVALID URL ***", ts ) ;
	}
	else if ( diff > 120 || hash1 != hash2 || body_msg == undefined || body_msg.length == 0 ) {
		error( req, res, "*** SESSION INVALID ***\n\t*** REFRESH THE PAGE ***", ts ) ;
	}
	else if ( action == "Shorten URL" ) {
		var client = new Client();
		var args = {
			data: { OrigUrl: orig_url },
			headers: { "Content-Type": "application/json" }
		} ;
		var short_url ;
		client.post( control_panel + "/link_save", args,
			function(data, response_raw) {
				console.log(data);
				var data_json = JSON.parse(data) ;
				short_url = data_json.ShortUrl ;
	      console.log( "short_url = " + short_url ) ;
				// var link = link_redirect + "/r/" + short_url ;
				var link = heroku_app + "/" + short_url ;
				page( req, res, ts, link_redirect, link ) ;
	  	}
		);
	}
}

var handle_get = function (req, res, next) {
	console.log( "Get: ..." ) ;
	var sl = req.params.sl ;
	console.log( sl ) ;
	if ( sl == undefined || sl.length == 0 ) {
		ts = new Date().getTime()
		console.log( ts )
		page( req, res, ts, control_panel ) ;
		return ;
	}
	handle_redirect( req, res ) ;
}

var handle_redirect = function ( req, res ) {
	console.log( "Redirect..." ) ;
	var client = new Client();
	client.get( link_redirect + "/r/" + req.params.sl,
		function(data, response_raw) {
			console.log(data);
			// console.log(response_raw);
			if (response_raw.statusCode != 200 ) {
				ts = new Date().getTime()
				page( req, res, ts, control_panel ) ;
				return ;

			}
			var data_json = JSON.parse(data) ;
			var orig_url = data_json.OrigUrl ;
			console.log( "orig_url = " + orig_url ) ;
			res.redirect(301, orig_url) ;
		}
	);
}

app.get('/:sl?', handle_get ) ;
app.post('/', handle_post ) ;

port = process.env.PORT
if (port == undefined || port.length == 0) {
	port = "3000"
}

app.listen( port, () => console.log( "Server running on Port ..." + port ) ) ;

/**
process.on('SIGTERM', () => {
  server.close(() => {
	console.log('Process terminated')
  })
})
*/
