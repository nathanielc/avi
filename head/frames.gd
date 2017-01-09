extends Node

var stream = null
var length = 0
var i = 0
var fps = 0
var stop_frame = 0
var last_frame = 0
var buf = RawArray()
var size = 0
var pos = 0
var _frames = Array()

func _init(s, len, f, sf, l):
	stream = s
	length = len
	fps = f
	stop_frame = sf
	last_frame = l
	
func _ready():
	set_process(true)

func _process(delta):
	if eof_reached():
		set_process(false)
		return
	var n = 0
	if buf.size() > 0:
		n = size - buf.size()
	else:
		size = stream.get_u32()
		pos += 4
		n = size
		buf.resize(0)
	
	var r = stream.get_partial_data(n)
	if r[0] != OK:
		return
	pos += r[1].size()
	buf.append_array(r[1])
	if buf.size() == size:
		var v = bytes2var(buf)
		i += 1
		buf.resize(0)
		_frames.push_back(v)

func get_var():
	if get_available_frames_count() == 0:
		return null
	var v = _frames[0]
	_frames.pop_front()
	return v

func get_available_frames_count():
	return _frames.size()

func eof_reached():
	return pos == length and _frames.size() == 0

func close():
	var n = length  - pos
	if n > 0:
		stream.get_data(n)
		
func max_time():
	return last_frame / fps