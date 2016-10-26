extends Control

func _ready():
	# Called every time the node is added to the scene.
	# Initialization here
	get_node("host_btn").connect("pressed", self, "_on_host_pressed")
	set_process(true)
	
func _on_host_pressed():
	global.host = get_node("host").get_text()
	global.port = get_node("port").get_val()
	var err = OS.execute("avi", ["-host", global.host, "-port", global.port, "-dir", "user://"], false)
	if err != OK:
		print("failed to start AVI server")
		return
	
	global.goto_scene("res://ui/views/game.tscn")