extends Control

var active_node

func _ready():
	# Called every time the node is added to the scene.
	# Initialization here
	get_node("new_battle").connect("pressed", self, "_on_new_battle_pressed")
	get_node("watch_game").connect("pressed", self, "_on_watch_game_pressed")
	get_node("fullscreen").connect("pressed", self, "_on_fullscreen_pressed")
	get_node("quit").connect("pressed", self, "_on_quit_pressed")
	set_process(true)
	var fullscreen = global.conf.get_value("ui", "fullscreen", false)
	OS.set_window_fullscreen(fullscreen)

func _clear_nodes():
	global.clear_error()
	var parent = get_node("../")
	for name in ["../new_battle_menu","../watch_replay_menu"]:
		if has_node(name):
			parent.remove_child(get_node(name))

func _on_new_battle_pressed():
	var nb = preload("res://ui/elements/new_battle.tscn").instance()
	nb.set_name("new_battle_menu")
	nb.set_pos(Vector2(200, 0))
	_clear_nodes()
	get_node("../").add_child(nb)

func _on_watch_game_pressed():
	var join = preload("res://ui/elements/watch_game.tscn").instance()
	join.set_name("watch_replay_menu")
	join.set_pos(Vector2(200, 0))
	_clear_nodes()
	get_node("../").add_child(join)

func _on_quit_pressed():
	get_tree().quit()
	
func _on_fullscreen_pressed():
	var fullscreen = !OS.is_window_fullscreen()
	OS.set_window_fullscreen(fullscreen)
	global.conf.set_value("ui", "fullscreen", fullscreen)
	global.save()
