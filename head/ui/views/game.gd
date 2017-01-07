extends Spatial

const MODE_PLAY = 0
const MODE_PAUSED = 1

var frames = null
var last_frame = null
var start = 0
var time = 0
var buf_size = 300
var mode = MODE_PLAY
var time_speed = 1.0
var dirty = false
var user_sliding = false
var max_time = 0

var objects = {}

onready var time_slider = get_node("hud/time/slider")
onready var time_max_label = get_node("hud/time/max")
onready var time_curr_label = get_node("hud/time/current")
onready var play_speed = get_node("hud/time/play_speed")
onready var play_pause = get_node("hud/time/play_pause")

var _play_tex = load("res://ui/images/play_button.png")
var _play_hover_tex = load("res://ui/images/play_button_hover.png")
var _pause_tex = load("res://ui/images/pause_button.png")
var _pause_hover_tex = load("res://ui/images/pause_button_hover.png")

func _ready():
	_request_frames(0)
	get_node("hud/quit").connect("pressed", self, "_on_quit_pressed")
	get_node("hud/time/play_pause").connect("pressed", self, "_on_play_paused_pressed")
	play_speed.connect("pressed", self, "_on_play_speed_pressed")
	time_slider.connect("input_event", self, "_on_slider_input_event")
	time_slider.set_step(0)
	set_process(true)

func _request_frames(s):
	last_frame = null
	start = s
	if frames != null:
		frames.close()
		time = s / frames.fps
		remove_child(frames)
	var err = global.client.get_frames(global.game_id, start, start+buf_size)
	if !err.is_ok():
		global.fail(err)
		return
	frames = err.value()
	add_child(frames)

func _clear():
	for id in objects.keys():
		var obj = objects[id]
		remove_child(obj)
	objects = {}

func _process(delta):
	if !dirty and mode == MODE_PAUSED:
		return
	time += delta * time_speed
	if frames.eof_reached():
		_request_frames(frames.stop_frame+1)
	while frames.get_available_frames_count() > 0:
		if dirty:
			_clear()
		var frame = last_frame
		if frame == null:
			frame = frames.get_var()
			# Discard non dictionary vars
			if typeof(frame) != TYPE_DICTIONARY:
				continue
			 # Discard vars that are not frames
			if !frame.has('Time'):
				continue

		var t = frame['Time']
		if t > time:
			# wait till time catches up to render frame
			last_frame = frame
			break
		last_frame = null
		
		_update_slider(t, frames.max_time())

		var score_txt = "Scores:\n"
		for fleet in frame['Scores'].keys():
			score_txt += ("%-20s%4.02f\n" % [fleet, frame['Scores'][fleet]])
		score_txt += "\n\n\nGameID: %s" % global.game_id
		get_node("hud/scores").set_text(score_txt)
		for id in frame['DeletedObjects']:
			if objects.has(id):
				var obj = objects[id]
				remove_child(obj)
				objects.erase(id)
		for obj in frame['Objects']:
			if objects.has(obj['ID']):
				var objNode = objects[obj['ID']]
				objNode.set_translation(obj['Position'])
			else:
				var s = load("res://models/"+obj['Model'] +".tscn")
				var objNode
				if s:
					objNode = s.instance()
				else:
					objNode = preload("res://models/cube.tscn").instance()
				add_child(objNode)
				objects[obj['ID']] = objNode
				objNode.set_translation(obj['Position'])
				var r = obj['Radius']
				if obj['Model'] == 'projectile':
					r = r * 20
				objNode.set_scale(Vector3(r,r,r))
		if dirty:
			dirty = false
			break

func _update_slider(t, m):
	if t > max_time:
		max_time = t
	if global.is_live:
		m = max_time
	if !user_sliding:
		time_slider.set_unit_value(t/m)
		time_max_label.set_text(_format_m_s(m))
		time_curr_label.set_text(_format_m_s(t))

func _format_m_s(t):
	return "%02d:%02d" % [t/60.0, int(t)%60]
	
func _on_quit_pressed():
	# TODO close out resources
	global.goto_scene("res://ui/views/main.tscn")
	
func _on_play_paused_pressed():
	if mode == MODE_PLAY:
		mode = MODE_PAUSED
		play_pause.set_normal_texture(_play_tex)
		play_pause.set_hover_texture(_play_hover_tex)
	else:
		mode = MODE_PLAY
		play_pause.set_normal_texture(_pause_tex)
		play_pause.set_hover_texture(_pause_hover_tex)
		
func _on_slider_input_event(event):
	if event.type == InputEvent.MOUSE_BUTTON and event.button_index == 1 and event.pressed:
		user_sliding = true
	if user_sliding and InputEvent.MOUSE_MOTION:
		var v = time_slider.get_unit_value()
		var t = frames.last_frame*v/frames.fps
		time_curr_label.set_text(_format_m_s(t))
	if event.type == InputEvent.MOUSE_BUTTON and event.button_index == 1 and not event.pressed:
		var v = time_slider.get_unit_value()
		if frames != null:
			dirty = true
			_request_frames(int(frames.last_frame*v))
		user_sliding = false

func _on_play_speed_pressed():
	if time_speed == 1.0:
		time_speed = 2.0
	elif time_speed == 2.0:
		time_speed = 4.0
	else:
		time_speed = 1.0
	play_speed.set_text("%dx" % time_speed)