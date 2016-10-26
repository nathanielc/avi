extends Control

func _ready():
	# Called every time the node is added to the scene.
	# Initialization here
	get_node("join").connect("pressed", self, "_on_join_pressed")
	set_process(true)

func _on_join_pressed():
	global.host = get_node("host").get_text()
	global.port = get_node("port").get_val()
	global.goto_scene("res://ui/views/game.tscn")

