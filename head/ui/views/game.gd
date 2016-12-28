extends Spatial

var frames = null
var last_frame = null
var time = 0

var objects = {}

func _ready():
	time = 0
	frames = global.client.get_frames(global.game_id)
	get_node("hud/quit").connect("pressed", self, "_on_quit_pressed")
	set_process(true)


func _process(delta):
	time += delta
	while last_frame != null or frames.get_available_packet_count() > 0:
		var frame = last_frame
		if frame == null:
			frame = frames.get_var()
			# Discard non dictionary vars
			if typeof(frame) != TYPE_DICTIONARY:
				continue
		# wait till time catches up to render frame
		var t = frame['Time']
		if t > time:
			last_frame = frame
			break
		last_frame = null
		var score_txt = "Scores:\n"
		for fleet in frame['Scores'].keys():
			score_txt += ("%-20s%4d\n" % [fleet, frame['Scores'][fleet]])
		score_txt += "\n\nTime: %0.02fs" % t
		get_node("hud/scores").set_text(score_txt)
		for id in frame['DeletedObjects']:
			if objects.has(id):
				var obj = objects[id]
				remove_child(obj)
				objects.erase(id)
		for obj in frame['NewObjects']:
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
		for obj in frame['ObjectUpdates']:
			if objects.has(obj['ID']):
				var objNode = objects[obj['ID']]
				objNode.set_translation(obj['Position'])

func _on_quit_pressed():
	global.goto_scene("res://ui/views/main.tscn")