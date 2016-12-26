extends Control

var client= null

func _ready():
	# Called every time the node is added to the scene.
	# Initialization here
	get_node("configure").connect("pressed", self, "_on_congfigure_pressed")
	_connect()
	#set_process(true)

func _connect():
	client = preload("res://client.gd").new()
	var err = client.connect("127.0.0.1", 4242)
	if err != OK:
		return
	get_node("status").set_text("Connected")
	global.client = client


#func _process(delta):
#	if !client.is_connected():
#		get_node("status").set_text("Disconnected")

func _on_congfigure_pressed():
	if !client.is_connected():
		_connect()

