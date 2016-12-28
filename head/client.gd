extends Node

var host = null
var port = null

func _ready():
	pass
	
func connect(h, p):
	host = h
	port = p
	var r = _do(HTTPClient.METHOD_GET, "/avi/ping", null)
	if r.has('pong'):
		return OK
	return "failed to connect"


func _do(method, path, data):
	var http = HTTPClient.new()
	var err = http.connect(host, port)
	if err != OK:
		return {}
		
	# Wait until resolved and connected
	while http.get_status() == HTTPClient.STATUS_CONNECTING or http.get_status() == HTTPClient.STATUS_RESOLVING:
		http.poll()

	if http.get_status() != HTTPClient.STATUS_CONNECTED:
		return {}

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
		return {}

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
	http.close()
	return response

func get_maps():
	return _do(HTTPClient.METHOD_GET, "/avi/maps", null)

func get_part_sets():
	return _do(HTTPClient.METHOD_GET, "/avi/part_sets", null)
	
func get_fleets():
	return _do(HTTPClient.METHOD_GET, "/avi/fleets", null)
	
func get_games():
	return _do(HTTPClient.METHOD_GET, "/avi/games", null)
	
func start_game(map, part_set, fleets):
	var r = _do(HTTPClient.METHOD_POST, "/avi/games", {"map" : map, "part_set" : part_set, "fleets": fleets})
	if r.has('id'):
		return r['id']
	return null
	

func get_frames(game_id):
	# Hand crafted HTTP client, so we can pass the tcp connection on to PacketPeerStream
	var tcp = StreamPeerTCP.new()
	var err = tcp.connect(host, port)
	if err != OK:
		return null
	# Use of  HTTP 1.0 is explicit so that we do not get chunked encoded responses.
	var err = tcp.put_data(("GET /avi/games/%s HTTP/1.0\r\nHost: localhost:4242\r\nUser-Agent: avi_head\r\nAccept: */*\r\n\r\n" % game_id).to_ascii())
	if err != OK:
		return null
	var state = 0
	# Read and discard headers
	while state != 4:
		var b = tcp.get_u8()
		if state == 0 and b == 13:
			state = 1
		elif state == 1 and b == 10:
			state = 2
		elif state == 2 and b == 13:
			state = 3
		elif state == 3 and b == 10:
			state = 4
		else:
			state = 0
	# The remain data is variant encoded objects
	var frames = PacketPeerStream.new()
	frames.set_stream_peer(tcp)
	return frames
