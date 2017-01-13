extends Node

const STATUS_DISCONNECTED = 0
const STATUS_CONNECTED = 1
const ping_interval = 5

var host = null
var port = null
var tcp = null
var http = null
var status = STATUS_DISCONNECTED
var time = 0

func _ready():
	set_process(true)
	
func _process(delta):
	if status == STATUS_CONNECTED:
		time += delta
		if time > ping_interval:
			_ping()
			time = 0
	
func connect(h, p):
	host = h
	port = p
	return _ping()

func _ping():
	var err = _do(HTTPClient.METHOD_GET, "/avi/ping", null)
	if err.is_ok():
		var r = err.value()
		if r.has('pong'):
			status = STATUS_CONNECTED
			return global.ok(OK)
	
	status = STATUS_DISCONNECTED
	return global.err("failed to connect to AVI server")

func is_connected():
	return status == STATUS_CONNECTED

func _connect_tcp():
	if true or (tcp == null or !tcp.is_connected()):
		tcp = StreamPeerTCP.new()
		var err = tcp.connect(host, port)
		if err != OK:
			return err
	return OK
	
func _connect_http():
	if http == null or http.get_status() != HTTPClient.STATUS_CONNECTED:
		http = HTTPClient.new()
		var err = http.connect(host, port)
		if err != OK:
			return global.err("failed to connect to HTTP server")
			
		# Wait until resolved and connected
		var itr = 0
		while itr < 5 and http.get_status() == HTTPClient.STATUS_CONNECTING or http.get_status() == HTTPClient.STATUS_RESOLVING:
			OS.delay_msec(5)
			http.poll()
			itr += 1
	
		if http.get_status() != HTTPClient.STATUS_CONNECTED:
			return global.err("timedout connecting to HTTP server")
	return global.ok(OK)

func _do(method, path, data):
	var err = _connect_http()
	if !err.is_ok():
		return err

	var body = ""
	var headers = []
	if data != null:
		body = data.to_json()
		headers = ["Content-Type: application/json", "Content-Length: " + str(body.length())]
	http.request(method, path, headers, body)
	while http.get_status() == HTTPClient.STATUS_REQUESTING:
		# Keep polling until the request is going on
		http.poll()
	
	# Make sure request finished well.
	if !(http.get_status() == HTTPClient.STATUS_BODY or http.get_status() == HTTPClient.STATUS_CONNECTED ):
		http.close()
		return global.err("failed to make HTTP request")

	var response = {}
	if (http.has_response()):
		# Array that will hold the data
		var rb = RawArray()
		while http.get_status() == HTTPClient.STATUS_BODY:
			# While there is body left to be read
			http.poll()
			var chunk = http.read_response_body_chunk() # Get a chunk
			if chunk.size() > 0:
				rb = rb + chunk # Append to read buffer
		response.parse_json(rb.get_string_from_utf8())
	return global.ok(response)

func get_maps():
	return _do(HTTPClient.METHOD_GET, "/avi/maps", null)

func get_part_sets():
	return _do(HTTPClient.METHOD_GET, "/avi/part_sets", null)
	
func get_fleets():
	return _do(HTTPClient.METHOD_GET, "/avi/fleets", null)
	
func get_games():
	return _do(HTTPClient.METHOD_GET, "/avi/games", null)
	
func start_game(map, part_set, fleets):
	var err = _do(HTTPClient.METHOD_POST, "/avi/games", {"map" : map, "part_set" : part_set, "fleets": fleets})
	if err.is_ok():
		var r = err.value()
		if r.has('id'):
			return global.ok(r['id'])
		err = global.err("missing ID in response")
	return global.wrap(err, "failed to start game")
	
const http_line       = 1
const header_key      = 2
const header_delim    = 3
const header_value    = 4
const first_carriage  = 5
const first_newline   = 6
const second_carriage = 7
const second_newline  = 8

func get_frames(game_id, start, stop):
	# Hand crafted HTTP client, so we can pass the tcp connection on to PacketPeerStream
	var err = _connect_tcp()
	if err != OK:
		return global.err("failed to connect to server")
		
	# Use of  HTTP 1.0 is explicit so that we do not get chunked encoded responses.
	var err = tcp.put_data(("GET /avi/games/%s?start=%d&stop=%d HTTP/1.0\r\nHost: %s\r\nUser-Agent: avi_head\r\nAccept: */*\r\n\r\n" % [game_id,start,stop, host]).to_ascii())
	if err != OK:
		return global.err("failed to make HTTP request")
		
	var state = http_line
	# Read and discard headers
	var code = 0
	var headers = {}
	var buf = RawArray()
	var k = ""
	var v = ""
	while state != second_newline:
		var b = tcp.get_u8()
		if state == http_line:
			if b == 10:
				var parts = buf.get_string_from_utf8().split(" ")
				if parts.size() >= 2:
					code = int(parts[1])
				buf.resize(0)
				state = header_key
			else:
				buf.append(b)
		elif state == header_key:
			if b == 58:
				state = header_delim
			else:
				buf.append(b)
		elif state == header_delim and b == 32:
				k = buf.get_string_from_utf8()
				buf.resize(0)
				state = header_value
		elif state == header_value:
			if b == 13:
				if k != "":
					v = buf.get_string_from_utf8()
					buf.resize(0)
					headers[k] = v
					k = ""
					v = ""
				state = first_carriage
			else:
				buf.append(b)
		elif state == first_carriage and b == 10:
			state = first_newline
		elif state == first_newline:
			if b == 13:
				state = second_carriage
			else:
				state = header_key
				buf.append(b)
		elif state == second_carriage and b == 10:
			state = second_newline
	if code != 200:
		# Read error message
		var res = tcp.get_data(int(headers['Content-Length']))
		var data = res[1].get_string_from_utf8()
		var r = {}
		r.parse_json(data)
		var err_msg = data
		if r.has('error'):
			err_msg = r['error']
		return global.err("non 200 repsonse code %d: %s" % [code, err_msg])
	if !headers.has('Frame-Count') or !headers.has('Frames-Per-Second') or !headers.has('Total-Frame-Count') or !headers.has('Stop-Frame'):
		return global.err("failed to understand server response")
	# The remaing data is variant encoded objects
	var frames = preload("res://frames.gd").new(tcp, int(headers['Content-Length']), float(headers['Frames-Per-Second']), int(headers['Stop-Frame']), int(headers['Total-Frame-Count']))
	return global.ok(frames)
